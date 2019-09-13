package pipe

import (
	"fmt"
	"sync"
	"time"

	"github.com/minus5/uof"
	"github.com/pkg/errors"
)

type fixtureApi interface {
	Fixture(lang uof.Lang, eventURN uof.URN) (*uof.Fixture, error)
	Fixtures(lang uof.Lang, to time.Time) (<-chan uof.Fixture, <-chan error)
}

type fixture struct {
	api       fixtureApi
	languages []uof.Lang
	requests  map[uof.URN]time.Time
	errc      chan<- error
	out       chan<- *uof.Message
	sync.Mutex
}

func (f *fixture) loop(preloadTo time.Time) stageFunc {
	return func(in <-chan *uof.Message, out chan<- *uof.Message, errc chan<- error) {
		f.errc, f.out = errc, out

		preloadDone := make(chan struct{})
		go func() {
			f.preload(preloadTo)
			close(preloadDone)
		}()
		for m := range in {
			out <- m
			if fc := m.FixtureChange; m.Type == uof.MessageTypeFixtureChange && fc != nil {
				go func() {
					<-preloadDone
					f.getFixture(fc.EventURN)
				}()
			}
		}

		f.errc, f.out = nil, nil
	}
}

func Fixture(api fixtureApi, languages []uof.Lang, preloadTo time.Time) stage {
	f := &fixture{
		api:       api,
		languages: languages,
		requests:  make(map[uof.URN]time.Time),
	}
	return Stage(f.loop(preloadTo))
}

func (f *fixture) preload(preloadTo time.Time) {
	var wg sync.WaitGroup
	wg.Add(len(f.languages))

	for _, lang := range f.languages {
		go func(lang uof.Lang) {
			defer wg.Done()
			in, errc := f.api.Fixtures(lang, preloadTo)
			for x := range in {
				f.out <- uof.NewFixtureMessage(lang, x)
				f.done(x.URN)
			}
			for err := range errc {
				f.errc <- err
			}
		}(lang)
	}
	wg.Wait()
}

func (f *fixture) checkpoint() time.Time {
	return time.Now().Add(-time.Minute)
}

func (f *fixture) requestedRecently(eventURN uof.URN) bool {
	f.Lock()
	defer f.Unlock()
	if last, ok := f.requests[eventURN]; ok {
		return last.After(f.checkpoint())
	}
	f.requests[eventURN] = time.Now()
	return false
}

func (f *fixture) done(eventURN uof.URN) {
	f.Lock()
	defer f.Unlock()
	f.requests[eventURN] = time.Now()
}

func (f *fixture) getFixture(eventURN uof.URN) {
	if f.requestedRecently(eventURN) {
		return
	}
	for _, lang := range f.languages {
		go func(lang uof.Lang) {

			x, err := f.api.Fixture(lang, eventURN)
			if err != nil {
				f.errc <- errors.Wrap(err, fmt.Sprintf("fixture request event: %s, lang: %s", eventURN, lang))
				return
			}
			f.out <- uof.NewFixtureMessage(lang, *x)
		}(lang)
	}
}
