package pipe

import (
	"sync"
	"time"

	"github.com/minus5/go-uof-sdk"
)

// Number of concurent api calls of one type. For example: no more than x
// running Player api calls in any point in time.
const ConcurentAPICallsLimit = 16

// What is a pipeline?
// There's no formal definition of a pipeline in Go; it's just one of many kinds
// of concurrent programs. Informally, a pipeline is a series of stages
// connected by channels, where each stage is a group of goroutines running the
// same function. In each stage, the goroutines:
//   * receive values from upstream via inbound channels
//   * perform some function on that data, usually producing new values
//   * send values downstream via outbound channels
//
// Each stage has any number of inbound and outbound channels, except the first
// and last stages, which have only outbound or inbound channels, respectively.
// The first stage is sometimes called the source or producer; the last stage,
// the sink or consumer.
// Reference: https://blog.golang.org/pipelines

type sourceStage func() (<-chan *uof.Message, <-chan error)
type InnerStage func(<-chan *uof.Message) (<-chan *uof.Message, <-chan error)
type ConsumerStage func(in <-chan *uof.Message) error
type stageFunc func(in <-chan *uof.Message, out chan<- *uof.Message, errc chan<- error)
type stageWithDrainFunc func(in <-chan *uof.Message, out chan<- *uof.Message, errc chan<- error) *sync.WaitGroup

func Build(source sourceStage, stages ...InnerStage) <-chan error {
	errors := make([]<-chan error, 0, len(stages)+2)
	in, errc := source()
	errors = append(errors, errc)

	for _, stage := range stages {
		out, errc := stage(in)
		errors = append(errors, errc)
		in = out
	}

	errors = append(errors, sink(in))
	return mergeErrors(errors)
}

// sink for the messages channel
// ensure that returned errors channel is closed after all messages chanels are closed
func sink(in <-chan *uof.Message) <-chan error {
	errc := make(chan error)
	go func() {
		for range in {
		}
		close(errc)
	}()
	return errc
}

// stolen from: https://blog.golang.org/pipelines
func mergeErrors(errors []<-chan error) <-chan error {
	var wg sync.WaitGroup
	out := make(chan error)

	// Start an output goroutine for each input channel in errors.  output
	// copies values from c to out until c is closed, then calls wg.Done.
	output := func(errc <-chan error) {
		for n := range errc {
			out <- n
		}
		wg.Done()
	}
	for _, errc := range errors {
		if errc == nil {
			continue
		}
		wg.Add(1)
		go output(errc)
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func Consumer(consumer ConsumerStage) InnerStage {
	return BufferedConsumer(consumer, 0)
}

func BufferedConsumer(consumer ConsumerStage, buffer int) InnerStage {
	return func(in <-chan *uof.Message) (<-chan *uof.Message, <-chan error) {
		out := make(chan *uof.Message)
		looperIn := make(chan *uof.Message, buffer)
		errc := make(chan error, 1)

		go func() { // tee in to out na looperIn
			defer close(out)
			defer close(looperIn)
			for m := range in {
				looperIn <- m
				out <- m
			}
		}()

		go func() {
			defer close(errc)

			if err := consumer(looperIn); err != nil {
				errc <- err
			}
			go func() { // for unclean exit; drain this chan
				for range looperIn {
				}
			}()
		}()
		return out, errc
	}
}

func Stage(looper stageFunc) InnerStage {
	return func(in <-chan *uof.Message) (<-chan *uof.Message, <-chan error) {
		out := make(chan *uof.Message)
		errc := make(chan error)

		go func() {
			defer close(out)
			defer close(errc)

			// looper has to range over in chan until it is closed
			// out channel will be consumed by following stages
			looper(in, out, errc)
		}()

		return out, errc
	}
}

func StageWithSubProcesses(looper stageWithDrainFunc) InnerStage {
	return func(in <-chan *uof.Message) (<-chan *uof.Message, <-chan error) {
		out := make(chan *uof.Message)
		errc := make(chan error)

		go func() {
			looperOut := make(chan *uof.Message)
			looperErrc := make(chan error)
			looperDone := make(chan struct{})
			defer close(looperOut)
			defer close(looperErrc)

			// copy form looper chans to output chanels
			go func() {
				defer close(out)
				defer close(errc)

				for {
					select {
					case <-looperDone:
						return
					case m, ok := <-looperOut:
						if !ok {
							return
						}
						out <- m
					case e, ok := <-looperErrc:
						if !ok {
							return
						}
						errc <- e
					}
				}
			}()

			// looper has to range over in chan until it is closed
			subProcs := looper(in, looperOut, looperErrc)
			// stop coping from looperOut/Errc to out/errc chans
			// and close the out/errc chans
			close(looperDone)
			// drain looper chans to allow sub processes to finish
			go func() {
				for range looperOut {
				}
			}()
			go func() {
				for range looperErrc {
				}
			}()
			subProcs.Wait()
		}()

		return out, errc
	}
}

func StageWithSubProcessesSync(looper stageWithDrainFunc) InnerStage {
	return func(in <-chan *uof.Message) (<-chan *uof.Message, <-chan error) {
		out := make(chan *uof.Message)
		errc := make(chan error)

		go func() {
			defer close(out)
			defer close(errc)

			// looper has to range over in chan until it is closed
			subProcs := looper(in, out, errc)
			subProcs.Wait()

		}()

		return out, errc
	}
}

func Simple(each func(m *uof.Message) error) InnerStage {
	return func(in <-chan *uof.Message) (<-chan *uof.Message, <-chan error) {
		out := make(chan *uof.Message)
		errc := make(chan error)

		go func() {
			defer close(out)
			defer close(errc)

			for m := range in {
				if err := each(m); err != nil {
					errc <- err
				}
				out <- m
			}
		}()
		return out, errc
	}
}

type expireMap struct {
	m        map[int]int
	interval time.Duration
	sync.Mutex
}

func newExpireMap(expireAfter time.Duration) *expireMap {
	em := &expireMap{
		m:        make(map[int]int),
		interval: expireAfter,
	}
	go func() {
		time.Sleep(expireAfter * 2)
		em.cleanup()
	}()
	return em
}

func (em *expireMap) cleanup() {
	em.Lock()
	defer em.Unlock()

	for k, v := range em.m {
		if em.expired(v) {
			delete(em.m, k)
		}
	}
}

func (em *expireMap) expired(v int) bool {
	return v < em.checkpoint()
}

func (em *expireMap) fresh(k int) bool {
	em.Lock()
	defer em.Unlock()

	if v, ok := em.m[k]; ok {
		return v > em.checkpoint()
	}
	return false
}

func (em *expireMap) checkpoint() int {
	return int(time.Now().UnixNano()) - int(em.interval)
}

func (em *expireMap) insert(key int) {
	em.Lock()
	defer em.Unlock()

	em.m[key] = int(time.Now().UnixNano())
}

func (em *expireMap) remove(key int) {
	em.Lock()
	defer em.Unlock()

	delete(em.m, key)
}
