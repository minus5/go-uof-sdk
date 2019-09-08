package pipe

import (
	"github.com/minus5/uof"
	"github.com/minus5/uof/api"
)

// on start recover all after timestamp or full
// on reconnect recover all after timestamp
// on alive with subscribed = 0, revocer that producer with last valid ts

func Recovery(api *api.Api, producers map[uof.Producer]int64, in <-chan *uof.Message) <-chan *uof.Message {
	out := make(chan *uof.Message)

	recoveryID := 1

	for producer, timestamp := range producers {
		go recover(api, producer, timestamp, recoveryID)
		recoveryID++
	}

	go func() {
		defer close(out)
		for m := range in {
			if m.Type != uof.MessageTypeAlive {
				out <- m
				continue
			}
			if _, ok := producers[m.Alive.Producer]; ok {
				producers[m.Alive.Producer] = m.Alive.Timestamp
				out <- m
			}
		}
	}()
	return out
}

func recover(api *api.Api, producer uof.Producer, timestamp int64, recoveryID int) {
	api.RequestRecovery(producer, timestamp, recoveryID)
}
