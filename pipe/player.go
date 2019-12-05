package pipe

import (
	"sync"
	"time"

	uof "github.com/minus5/go-uof-sdk"
)

type playerApi interface {
	Player(lang uof.Lang, playerID int) (*uof.Player, error)
}

type player struct {
	api       playerApi
	em        *expireMap
	languages []uof.Lang // suported languages
	errc      chan<- error
	out       chan<- *uof.Message
	rateLimit chan struct{}
	subProcs  *sync.WaitGroup
}

func Player(api playerApi, languages []uof.Lang) stage {
	p := &player{
		api:       api,
		languages: languages,
		em:        newExpireMap(time.Hour),
		subProcs:  &sync.WaitGroup{},
		rateLimit: make(chan struct{}, ConcurentApiCallsLimit),
	}
	return StageWithSubProcesses(p.loop)
}

func (p *player) loop(in <-chan *uof.Message, out chan<- *uof.Message, errc chan<- error) *sync.WaitGroup {
	p.errc, p.out = errc, out

	for m := range in {
		out <- m
		if m.Is(uof.MessageTypeOddsChange) {
			m.OddsChange.EachPlayer(p.get)
		}
	}
	return p.subProcs
}

func (p *player) get(playerID int) {
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
			ap, err := p.api.Player(lang, playerID)
			if err != nil {
				p.errc <- err
				return
			}
			p.out <- uof.NewPlayerMessage(lang, ap)
			p.em.insert(key)
		}(lang)
	}
}
