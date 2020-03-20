package pipe

import (
	"sync"
	"time"

	"github.com/minus5/go-uof-sdk"
)

type competitorAPI interface {
	Competitor(lang uof.Lang, competitorID int) (*uof.CompetitorPlayer, error)
}

type competitor struct {
	api       competitorAPI
	em        *expireMap
	languages []uof.Lang // suported languages
	errc      chan<- error
	out       chan<- *uof.Message
	rateLimit chan struct{}
	subProcs  *sync.WaitGroup
}

func Competitor(api competitorAPI, languages []uof.Lang) InnerStage {
	p := &competitor{
		api:       api,
		languages: languages,
		em:        newExpireMap(time.Hour),
		subProcs:  &sync.WaitGroup{},
		rateLimit: make(chan struct{}, ConcurentAPICallsLimit),
	}
	return StageWithSubProcessesSync(p.loop)
}

func (p *competitor) loop(in <-chan *uof.Message, out chan<- *uof.Message, errc chan<- error) *sync.WaitGroup {
	p.errc, p.out = errc, out

	for m := range in {
		out <- m
		if m.Is(uof.MessageTypeOddsChange) {
			m.OddsChange.EachCompetitor(func(competitorID int) {
				p.get(competitorID, m.ReceivedAt)
			})
		}
	}
	return p.subProcs
}

func (p *competitor) get(competitorID, requestedAt int) {
	p.subProcs.Add(len(p.languages))
	for _, lang := range p.languages {
		go func(lang uof.Lang) {
			defer p.subProcs.Done()
			p.rateLimit <- struct{}{}
			defer func() { <-p.rateLimit }()

			key := uof.UIDWithLang(competitorID, lang)
			if p.em.fresh(key) {
				return
			}
			p.em.insert(key)
			cp, err := p.api.Competitor(lang, competitorID)
			if err != nil {
				p.em.remove(key)
				p.errc <- err
				return
			}
			p.out <- uof.NewCompetitorMessage(lang, cp, requestedAt)
		}(lang)
	}
}
