package pipe

import (
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
}

func Player(api playerApi, languages []uof.Lang) stage {
	p := &player{
		api:       api,
		languages: languages,
		em:        newExpireMap(time.Hour),
	}
	return Stage(p.loop)

}

func (p *player) loop(in <-chan *uof.Message, out chan<- *uof.Message, errc chan<- error) {
	p.errc, p.out = errc, out
	defer func() {
		p.errc, p.out = nil, nil
	}()

	for m := range in {
		out <- m
		if m.Type == uof.MessageTypeOddsChange {
			go p.get(m.OddsChange)
		}
	}
}

func (p *player) get(oc *uof.OddsChange) {
	if oc == nil {
		return
	}

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
