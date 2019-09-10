package pipe

import (
	"context"
	"fmt"
	"time"

	"github.com/minus5/uof"
	"github.com/pkg/errors"
)

// on start recover all after timestamp or full
// on reconnect recover all after timestamp
// on alive with subscribed = 0, revocer that producer with last valid ts
// TODO counting on number of requests per period

// Recovery requests limits: https://docs.betradar.com/display/BD/UOF+-+Access+restrictions+for+odds+recovery
// Recovery sequence expalined: https://docs.betradar.com/display/BD/UOF+-+Recovery+using+API

// A client should always store the last successfully received alive message (or
// its timestamp) from each producer. In case of a disconnection, recovery since
// after timestamp should be issued for each affected producer, using the
// timestamp of the last successfully processed alive message before issues
// occurred.

type recoveryProducer struct {
	producer              uof.Producer
	status                uof.ProducerStatus // current status of the producer
	aliveTimestamp        int64              // last alive timestamp
	requestID             int                // last recovery requestID
	statusChangedAt       int64              // last change of the status
	recoveryRequestCancel context.CancelFunc
}

func (p *recoveryProducer) setStatus(newStatus uof.ProducerStatus) {
	if p.status != newStatus {
		p.status = newStatus
		ct := uof.CurrentTimestamp()
		if p.statusChangedAt >= ct {
			// ensure monotonic increase (for tests)
			ct = p.statusChangedAt + 1
		}
		p.statusChangedAt = ct
	}
}

// If producer is back more than recovery window (defined for each producer)
// it has to make full recovery (forced with timestamp = 0).
// Otherwise recovery after timestamp is done.
func (p *recoveryProducer) recoveryTimestamp() int64 {
	if uof.CurrentTimestamp()-p.aliveTimestamp >= p.producer.RecoveryWindow() {
		return 0
	}
	return p.aliveTimestamp
}

type recovery struct {
	api       recoveryApi
	requestID int
	producers []*recoveryProducer
	errc      chan<- error
}

type recoveryApi interface {
	RequestRecovery(producer uof.Producer, timestamp int64, requestID int) error
}

func newRecovery(api recoveryApi, producers uof.ProducersChange) *recovery {
	r := &recovery{
		api: api,
	}
	ct := uof.CurrentTimestamp()
	for _, p := range producers {
		r.producers = append(r.producers, &recoveryProducer{
			producer:        p.Producer,
			aliveTimestamp:  p.Timestamp,
			status:          uof.ProducerStatusDown,
			statusChangedAt: ct,
		})
	}
	return r
}

func (r *recovery) log(err error) {
	select {
	case r.errc <- err:
	default:
	}
}

func (r *recovery) requestRecoveryForAll() {
	for _, p := range r.producers {
		if p.status == uof.ProducerStatusInRecovery {
			r.requestRecovery(p)
		}
	}
}

func (r *recovery) requestRecovery(p *recoveryProducer) {
	p.setStatus(uof.ProducerStatusInRecovery)
	p.requestID = r.nextRequestID()

	if cancel := p.recoveryRequestCancel; cancel != nil {
		cancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	p.recoveryRequestCancel = cancel

	go func(producer uof.Producer, timestamp int64, requestID int) {
		for {
			op := fmt.Sprintf("recovery for %s, timestamp: %d, requestID: %d", p.producer.Code(), p.recoveryTimestamp(), r.requestID)
			r.log(fmt.Errorf("staring %s", op))
			err := r.api.RequestRecovery(producer, timestamp, requestID)
			if err == nil {
				return
			}
			r.errc <- errors.Wrap(err, "failed "+op)
			// wait a minute
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Minute):
			}
		}
	}(p.producer, p.recoveryTimestamp(), p.requestID)
}

func (r *recovery) nextRequestID() int {
	r.requestID++
	return r.requestID
}

func (r *recovery) find(producer uof.Producer) *recoveryProducer {
	for _, rp := range r.producers {
		if rp.producer == producer {
			return rp
		}
	}
	return nil
}

// returns false if we are not interested in that producer
func (r *recovery) alive(producer uof.Producer, timestamp int64, subscribed int) {
	p := r.find(producer)
	if p == nil {
		return // this is expected we are getting alive for all producers in uof (with Subscribed=0)
	}
	if subscribed == 0 {
		r.requestRecovery(p)
		return
	}
	p.aliveTimestamp = timestamp
}

func (r *recovery) snapshotComplete(producer uof.Producer, requestID int) {
	p := r.find(producer)
	if p == nil {
		r.log(fmt.Errorf("unexpected producer %s", producer))
		return
	}
	if p.requestID != requestID {
		r.log(fmt.Errorf("unexpected requestID %d, expected %d, for producer %s", requestID, p.requestID, producer))
	}
	p.setStatus(uof.ProducerStatusActive)
	p.requestID = 0
}

func (r *recovery) connectionUp() {
	for _, p := range r.producers {
		if p.status == uof.ProducerStatusDown {
			r.requestRecovery(p)
		}
	}
}

func (r *recovery) connectionDown() {
	for _, p := range r.producers {
		p.setStatus(uof.ProducerStatusDown)
	}
}

func (r *recovery) statusChangedAt() int64 {
	var sc int64
	for _, r := range r.producers {
		if r.statusChangedAt > sc {
			sc = r.statusChangedAt
		}
	}
	return sc
}

func (r *recovery) loop(in <-chan *uof.Message, out chan<- *uof.Message, errc chan<- error) {
	r.errc = errc
	var statusChangedAt int64
	for m := range in {
		switch m.Type {
		case uof.MessageTypeAlive:
			r.alive(m.Alive.Producer, m.Alive.Timestamp, m.Alive.Subscribed)
		case uof.MessageTypeSnapshotComplete:
			r.snapshotComplete(m.SnapshotComplete.Producer, m.SnapshotComplete.RequestID)
		case uof.MessageTypeConnection:
			switch m.Connection.Status {
			case uof.ConnectionStatusUp:
				r.connectionUp()
			case uof.ConnectionStatusDown:
				r.connectionDown()
			}
		default:
			out <- m
			continue
		}
		if sc := r.statusChangedAt(); sc > statusChangedAt {
			statusChangedAt = sc
			out <- r.producersChangeMessage()
		}
	}
	r.errc = nil
}

func (r *recovery) producersChangeMessage() *uof.Message {
	var psc uof.ProducersChange
	for _, p := range r.producers {
		pc := uof.ProducerChange{
			Producer:  p.producer,
			Status:    p.status,
			Timestamp: p.statusChangedAt,
		}
		if p.status == uof.ProducerStatusInRecovery {
			pc.RecoveryID = p.requestID
		}
		psc = append(psc, pc)
	}
	return uof.NewProducersChangeMessage(psc)
}

func Recovery(api recoveryApi, producers uof.ProducersChange) stage {
	r := newRecovery(api, producers)
	return Stage(r.loop)
}
