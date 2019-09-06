package pipe

import (
	"fmt"
	"sync"

	"github.com/minus5/svckit/log"
	"github.com/minus5/uof"
	"github.com/minus5/uof/api"
)

func VariantMarket(api *api.Api, languages []uof.Lang, in <-chan *uof.Message) <-chan *uof.Message {
	out := make(chan *uof.Message, 16)
	done := make(map[string]struct{})

	go func() {
		var wg sync.WaitGroup
		defer close(out)

		for m := range in {
			out <- m
			m.OddsChange.EachVariantMarket(func(marketID int, variant string) {
				k := fmt.Sprintf("%d %s", marketID, variant)
				if _, ok := done[k]; ok {
					return
				}
				done[k] = struct{}{}

				wg.Add(1)
				go func(marketID int, variant string) {
					defer wg.Done()
					for _, lang := range languages {
						buf, err := api.MarketVariant(lang, marketID, variant)
						if err != nil {
							log.I("marketID", int(marketID)).S("lang", lang.String()).S("variant", variant).Error(err)
							continue
						}
						mm, err := uof.NewMarketsMessage(lang, buf)
						if err != nil {
							log.Error(err)
							continue
						}
						out <- mm
					}
				}(marketID, variant)
			})
		}
		wg.Wait()
	}()
	return out
}

// osigurava da uvijek na startu prvo posalje markete za sve jezike
func Markets(api *api.Api, languages []uof.Lang, in <-chan *uof.Message) <-chan *uof.Message {
	out := make(chan *uof.Message, len(languages))

	for _, lang := range languages {
		buf, err := api.Markets(lang)
		if err != nil {
			log.Error(err)
			continue
		}
		mm, err := uof.NewMarketsMessage(lang, buf)
		if err != nil {
			log.Error(err)
			continue
		}
		out <- mm
	}

	go func() {
		defer close(out)
		for m := range in {
			out <- m
		}
	}()
	return out
}
