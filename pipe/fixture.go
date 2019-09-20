package pipe

import (
	"fmt"
	"sync"
	"time"

	"github.com/minus5/uof"
)

type fixtureApi interface {
	Fixture(lang uof.Lang, eventURN uof.URN) (*uof.Fixture, error)
	Fixtures(lang uof.Lang, to time.Time) (<-chan uof.Fixture, <-chan error)
}

type fixture struct {
	api       fixtureApi
	languages []uof.Lang           // suported languages
	requests  map[string]time.Time // last sucessful request time
	errc      chan<- error
	out       chan<- *uof.Message
	preloadTo time.Time
	sync.Mutex
}

func Fixture(api fixtureApi, languages []uof.Lang, preloadTo time.Time) stage {
	f := &fixture{
		api:       api,
		languages: languages,
		requests:  make(map[string]time.Time),
		preloadTo: preloadTo,
	}
	return Stage(f.loop)
}

// Na sto sve pazim ovdje:
//  * na pocetku napravim preload
//  * za vrijeme preload-a ne pokrecem pojedinacne
//  * za vrijeme preload-a za zaustavljam lanaca, saljem dalje in -> out
//  * nakon sto zavrsi preload napravim one koje preload nije ubacio
//  * ne radim request cesce od svakih x (bitno za replay, da ne proizvedem puno requesta)
//  * kada radim scenario replay htio bi da samo jednom opali, dok je neki in process da na pokrece isti
func (f *fixture) loop(in <-chan *uof.Message, out chan<- *uof.Message, errc chan<- error) {
	f.errc, f.out = errc, out
	defer func() {
		f.errc, f.out = nil, nil
	}()

	for _, u := range f.preloadLoop(in) {
		f.getFixture(u)
	}
	for m := range in {
		out <- m
		if u := f.eventURN(m); u != uof.NoURN {
			f.getFixture(u)
		}
	}

}

func (f *fixture) eventURN(m *uof.Message) uof.URN {
	if m.Type != uof.MessageTypeFixtureChange || m.FixtureChange == nil {
		return uof.NoURN
	}
	return m.FixtureChange.EventURN
}

// returns list of fixture changes urns appeared in 'in' during preload
func (f *fixture) preloadLoop(in <-chan *uof.Message) []uof.URN {
	done := make(chan struct{})
	go func() {
		f.preload()
		close(done)
	}()

	var urns []uof.URN
	for {
		select {
		case m, ok := <-in:
			if !ok {
				return nil
			}
			f.out <- m
			if u := f.eventURN(m); u != uof.NoURN {
				urns = append(urns, u)
			}
		case <-done:
			return urns
		}
	}
}

func (f *fixture) preload() {
	if f.preloadTo.IsZero() {
		return
	}
	var wg sync.WaitGroup
	wg.Add(len(f.languages))
	for _, lang := range f.languages {
		go func(lang uof.Lang) {
			defer wg.Done()
			in, errc := f.api.Fixtures(lang, f.preloadTo)
			for x := range in {
				f.out <- uof.NewFixtureMessage(lang, x)
				f.done(lang, x.URN)
			}
			for err := range errc {
				f.errc <- err
			}
		}(lang)
	}
	wg.Wait()
}

func (f *fixture) getFixture(eventURN uof.URN) {
	for _, lang := range f.languages {
		if f.requestedRecently(lang, eventURN) {
			continue
		}
		go func(lang uof.Lang) {
			x, err := f.api.Fixture(lang, eventURN)
			if err != nil {
				return
			}
			f.out <- uof.NewFixtureMessage(lang, *x)
			f.done(lang, eventURN)
		}(lang)
	}
}

func (f *fixture) done(lang uof.Lang, u uof.URN) {
	f.Lock()
	defer f.Unlock()

	f.requests[f.key(lang, u)] = time.Now()
}

func (f *fixture) key(lang uof.Lang, u uof.URN) string {
	return fmt.Sprintf("%s %d", u, lang)
}

func (f *fixture) requestedRecently(lang uof.Lang, eventURN uof.URN) bool {
	f.Lock()
	defer f.Unlock()

	key := f.key(lang, eventURN)
	if last, ok := f.requests[key]; ok {
		return last.After(f.checkpoint())
	}
	return false
}

func (f *fixture) checkpoint() time.Time {
	return time.Now().Add(-time.Minute)
}
