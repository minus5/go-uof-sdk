package pipe

import (
	"sync"
	"time"

	"github.com/minus5/go-uof-sdk"
)

type fixtureApi interface {
	Fixture(lang uof.Lang, eventURN uof.URN) (*uof.Fixture, error)
	Fixtures(lang uof.Lang, to time.Time) (<-chan uof.Fixture, <-chan error)
}

type fixture struct {
	api       fixtureApi
	languages []uof.Lang // suported languages
	em        *expireMap
	errc      chan<- error
	out       chan<- *uof.Message
	preloadTo time.Time
	subProcs  *sync.WaitGroup
	rateLimit chan struct{}
	sync.Mutex
}

func Fixture(api fixtureApi, languages []uof.Lang, preloadTo time.Time) InnerStage {
	f := &fixture{
		api:       api,
		languages: languages,
		em:        newExpireMap(time.Minute),
		//requests:  make(map[string]time.Time),
		subProcs:  &sync.WaitGroup{},
		rateLimit: make(chan struct{}, ConcurentApiCallsLimit),
		preloadTo: preloadTo,
	}
	return StageWithSubProcessesSync(f.loop)
}

// Na sto sve pazim ovdje:
//  * na pocetku napravim preload
//  * za vrijeme preload-a ne pokrecem pojedinacne
//  * za vrijeme preload-a za zaustavljam lanaca, saljem dalje in -> out
//  * nakon sto zavrsi preload napravim one koje preload nije ubacio
//  * ne radim request cesce od svakih x (bitno za replay, da ne proizvedem puno requesta)
//  * kada radim scenario replay htio bi da samo jednom opali, dok je neki in process da na pokrece isti
func (f *fixture) loop(in <-chan *uof.Message, out chan<- *uof.Message, errc chan<- error) *sync.WaitGroup {
	f.errc, f.out = errc, out

	for _, u := range f.preloadLoop(in) {
		f.getFixture(u)
	}
	for m := range in {
		out <- m
		if u := f.eventURN(m); u != uof.NoURN {
			f.getFixture(u)
		}
	}

	return f.subProcs
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

	f.subProcs.Add(1)
	go func() {
		defer f.subProcs.Done()
		f.preload()
		close(done)
	}()

	var urns []uof.URN
	for {
		select {
		case m, ok := <-in:
			if !ok {
				return urns
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
				f.em.insert(uof.UIDWithLang(x.URN.EventID(), lang))
			}
			for err := range errc {
				f.errc <- err
			}
		}(lang)
	}
	wg.Wait()
}

func (f *fixture) getFixture(eventURN uof.URN) {
	f.subProcs.Add(len(f.languages))
	for _, lang := range f.languages {
		go func(lang uof.Lang) {
			defer f.subProcs.Done()
			f.rateLimit <- struct{}{}
			defer func() { <-f.rateLimit }()

			key := uof.UIDWithLang(eventURN.EventID(), lang)
			if f.em.fresh(key) {
				return
			}
			x, err := f.api.Fixture(lang, eventURN)
			if err != nil {
				f.errc <- err
				return
			}
			f.out <- uof.NewFixtureMessage(lang, *x)
			f.em.insert(key)
		}(lang)
	}
}
