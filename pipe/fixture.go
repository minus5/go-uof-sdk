package pipe

import (
	"sync"
	"time"

	"github.com/minus5/svckit/log"
	"github.com/minus5/uof"
	"github.com/minus5/uof/api"
)

func Fixture(api *api.Api, languages []uof.Lang, in <-chan *uof.Message) <-chan *uof.Message {
	out := make(chan *uof.Message, 16)
	requestTimes := make(map[uof.URN]time.Time)
	go func() {
		var wg sync.WaitGroup
		defer close(out)

		for m := range in {
			out <- m
			if m.Type != uof.MessageTypeFixtureChange {
				continue
			}

			// zako je zadnji bio prije minute nemoj kretati u novi
			if last, ok := requestTimes[m.EventURN]; ok {
				if last.After(time.Now().Add(-time.Minute)) {
					continue
				}
			}
			requestTimes[m.EventURN] = time.Now()

			for _, lang := range languages {
				wg.Add(1)
				go func(m *uof.Message, lang uof.Lang) {
					defer wg.Done()
					buf, err := api.Fixture(lang, m.EventURN)
					if err != nil {
						log.S("urn", m.EventURN.String()).Error(err)
						return
					}
					fm, err := m.AsFixture(lang, buf)
					if err != nil {
						log.S("urn", m.EventURN.String()).Error(err)
						return
					}
					out <- fm
				}(m, lang)
			}
		}

		wg.Wait()
	}()
	return out
}
