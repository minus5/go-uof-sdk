package pipe

import (
	"sync"

	"github.com/minus5/svckit/log"
	"github.com/minus5/uof"
	"github.com/minus5/uof/api"
)

func Player(api *api.Api, languages []uof.Lang, in <-chan *uof.Message) <-chan *uof.Message {
	out := make(chan *uof.Message, 16)
	done := make(map[int]struct{})

	go func() {
		var wg sync.WaitGroup
		defer close(out)

		for m := range in {
			out <- m
			if m.Type != uof.MessageTypeOddsChange {
				continue
			}
			m.OddsChange.EachPlayer(func(player int) {
				if _, ok := done[player]; ok {
					return
				}
				done[player] = struct{}{}

				wg.Add(1)
				go func(playerID int) {
					defer wg.Done()
					for _, lang := range languages {
						buf, err := api.Player(lang, playerID)
						if err != nil {
							log.I("player", playerID).Error(err)
							continue
						}
						pm, err := uof.NewPlayerMessage(lang, buf)
						if err != nil {
							log.Error(err)
							continue
						}
						out <- pm
					}
				}(player)
			})
		}
		wg.Wait()
	}()
	return out
}
