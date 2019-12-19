package pipe

import (
	"sync"
	"time"

	"github.com/minus5/go-uof-sdk"
)

type playerAPI interface {
	Player(lang uof.Lang, playerID int) (*uof.Player, error)
}

type player struct {
	api       playerAPI
	em        *expireMap
	languages []uof.Lang // suported languages
	errc      chan<- error
	out       chan<- *uof.Message
	rateLimit chan struct{}
	subProcs  *sync.WaitGroup
}

func Player(api playerAPI, languages []uof.Lang) InnerStage {
	p := &player{
		api:       api,
		languages: languages,
		em:        newExpireMap(time.Hour),
		subProcs:  &sync.WaitGroup{},
		rateLimit: make(chan struct{}, ConcurentAPICallsLimit),
	}
	return StageWithSubProcessesSync(p.loop)
}

func (p *player) loop(in <-chan *uof.Message, out chan<- *uof.Message, errc chan<- error) *sync.WaitGroup {
	p.errc, p.out = errc, out

	for m := range in {
		out <- m
		if m.Is(uof.MessageTypeOddsChange) {
			m.OddsChange.EachPlayer(func(playerID int) {
				p.get(playerID, m.ReceivedAt)
			})
		}
	}
	return p.subProcs
}

func (p *player) get(playerID, requestedAt int) {
	p.subProcs.Add(len(p.languages))
	for _, lang := range p.languages {
		go func(lang uof.Lang) {
			defer p.subProcs.Done()
			p.rateLimit <- struct{}{}
			defer func() { <-p.rateLimit }()

			key := uof.UIDWithLang(playerID, lang)
			if p.em.fresh(key) {
				return
			}
			p.em.insert(key)
			ap, err := p.api.Player(lang, playerID)
			if err != nil {
				p.em.remove(key)
				p.errc <- err
				return
			}
			p.out <- uof.NewPlayerMessage(lang, ap, requestedAt)
		}(lang)
	}
}
