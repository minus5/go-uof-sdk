package pipe

import (
	"sync"
	"time"

	"github.com/minus5/uof"
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
	subProcs  *sync.WaitGroup
}

func Player(api playerApi, languages []uof.Lang) stage {
	var wg sync.WaitGroup
	p := &player{
		api:       api,
		languages: languages,
		em:        newExpireMap(time.Hour),
		subProcs:  &wg,
	}
	return StageWithDrain(p.loop)

}

func (p *player) loop(in <-chan *uof.Message, out chan<- *uof.Message, errc chan<- error) *sync.WaitGroup {
	p.errc, p.out = errc, out

	for m := range in {
		out <- m
		if m.Is(uof.MessageTypeOddsChange) {
			p.subProcs.Add(1)
			go p.get(m.OddsChange)
		}
	}
	return p.subProcs
}

func (p *player) get(oc *uof.OddsChange) {
	defer p.subProcs.Done()

	oc.EachPlayer(func(playerID int) {
		for _, lang := range p.languages {
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
		}
	})
}
