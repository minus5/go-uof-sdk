package pipe

import (
	"strings"
	"sync"
	"time"

	"github.com/minus5/go-uof-sdk"
)

type marketsAPI interface {
	Markets(lang uof.Lang) (uof.MarketDescriptions, error)
	MarketVariant(lang uof.Lang, marketID int, variant string) (uof.MarketDescriptions, error)
}

type markets struct {
	api       marketsAPI
	languages []uof.Lang
	em        *expireMap
	errc      chan<- error
	out       chan<- *uof.Message
	rateLimit chan struct{}
	subProcs  *sync.WaitGroup
}

// getting all markets on the start
func Markets(api marketsAPI, languages []uof.Lang) InnerStage {
	var wg sync.WaitGroup
	m := &markets{
		api:       api,
		languages: languages,
		em:        newExpireMap(24 * time.Hour),
		subProcs:  &wg,
		rateLimit: make(chan struct{}, ConcurentAPICallsLimit),
	}
	return StageWithSubProcessesSync(m.loop)
}

func (s *markets) loop(in <-chan *uof.Message, out chan<- *uof.Message, errc chan<- error) *sync.WaitGroup {
	s.out, s.errc = out, errc

	s.getAll()
	for m := range in {
		out <- m
		if m.Is(uof.MessageTypeOddsChange) {
			m.OddsChange.EachVariantMarket(func(marketID int, variant string) {
				s.variantMarket(marketID, variant, m.ReceivedAt)
			})
		}
	}
	return s.subProcs
}

func (s *markets) getAll() {
	s.subProcs.Add(len(s.languages))
	requestedAt := uof.CurrentTimestamp()

	for _, lang := range s.languages {
		go func(lang uof.Lang) {
			defer s.subProcs.Done()

			ms, err := s.api.Markets(lang)
			if err != nil {
				s.errc <- err
				return
			}
			s.out <- uof.NewMarketsMessage(lang, ms, requestedAt)
		}(lang)
	}
}

func (s *markets) variantMarket(marketID int, variant string, requestedAt int) {
	if strings.HasPrefix(variant, "pre:playerprops") {
		// TODO: it is not working for this type of variant markets
		return
	}
	s.subProcs.Add(len(s.languages))

	for _, lang := range s.languages {
		go func(lang uof.Lang) {
			defer s.subProcs.Done()
			s.rateLimit <- struct{}{}
			defer func() { <-s.rateLimit }()

			key := uof.UIDWithLang(uof.Hash(variant)<<32|marketID, lang)
			if s.em.fresh(key) {
				return
			}

			ms, err := s.api.MarketVariant(lang, marketID, variant)
			if err != nil {
				s.errc <- err
				return
			}
			s.out <- uof.NewMarketsMessage(lang, ms, requestedAt)
			s.em.insert(key)
		}(lang)
	}
}
