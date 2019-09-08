package pipe

import (
	"sync"

	"github.com/minus5/uof"
)

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

type source func() (<-chan *uof.Message, <-chan error)
type stage func(<-chan *uof.Message) (<-chan *uof.Message, <-chan error)

func Build(source source, stages ...stage) <-chan error {
	var errors []<-chan error
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

func Stage(looper func(in <-chan *uof.Message, out chan<- *uof.Message, errc chan<- error)) stage {
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

func Simple(each func(m *uof.Message) error) stage {
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
