package pipe

import (
	"fmt"

	"github.com/minus5/uof"
	"github.com/minus5/uof/api"
	"github.com/pkg/errors"
)

// on start recover all after timestamp or full
// on reconnect recover all after timestamp
// on alive with subscribed = 0, revocer that producer with last valid ts
// TODO counting on number of requests per period

type producer struct {
	id             uof.Producer
	aliveTimestamp int64
}

type recovery struct {
	api        *api.Api
	recoveryID int
	producers  map[uof.Producer]*producer
}

func newRecovery(api *api.Api, producers map[uof.Producer]int64) *recovery {
	r := &recovery{
		api:       api,
		producers: make(map[uof.Producer]*producer),
	}
	for id, timestamp := range producers {
		r.producers[id] = &producer{
			id:             id,
			aliveTimestamp: timestamp,
		}
	}
	return r
}

func (r *recovery) requestRecoveryForAll(errc chan<- error) {
	r.recoveryID++
	for _, p := range r.producers {
		go func(producer uof.Producer, timestamp int64, recoveryID int) {
			// TODO log message as error ?!
			errc <- fmt.Errorf("requesting recovery for %s, timestamp: %d", producer.Code(), timestamp)
			if err := r.api.RequestRecovery(producer, timestamp, recoveryID); err != nil {
				errc <- errors.Wrap(err, "api request failed")
			}
		}(p.id, p.aliveTimestamp, r.recoveryID)
		r.recoveryID++
	}

}

func (r *recovery) requestRecovery(p *producer, errc chan<- error) {
	r.recoveryID++
	go func(producer uof.Producer, timestamp int64, recoveryID int) {
		errc <- fmt.Errorf("requesting recovery for %s, timestamp: %d\n", producer.Code(), timestamp)
		if err := r.api.RequestRecovery(producer, timestamp, recoveryID); err != nil {
			errc <- errors.Wrap(err, "api request failed")
		}
	}(p.id, p.aliveTimestamp, r.recoveryID)
}

// returns false if we are not interested in that producer
func (r *recovery) alive(a *uof.Alive, errc chan<- error) bool {
	p, ok := r.producers[a.Producer]
	if !ok {
		return false // not interested in this producer
	}
	if a.Subscribed == 0 {
		r.requestRecovery(p, errc)
		return true
	}
	p.aliveTimestamp = a.Timestamp
	return true
}

func (r *recovery) loop(in <-chan *uof.Message, out chan<- *uof.Message, errc chan<- error) {
	for m := range in {
		if a := m.Alive; m.Type == uof.MessageTypeAlive && a != nil {
			if !r.alive(a, errc) {
				continue // do not propagate
			}
		}
		if c := m.Connection; m.Type == uof.MessageTypeConnection && c != nil && c.Status == uof.ConnectionStatusUp {
			r.requestRecoveryForAll(errc)
		}
		out <- m
	}
}

func Recovery(api *api.Api, producers map[uof.Producer]int64) stage {
	r := newRecovery(api, producers)
	return Stage(r.loop)
}
