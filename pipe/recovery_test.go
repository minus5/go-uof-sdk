package pipe

import (
	"testing"

	uof "github.com/minus5/go-uof-sdk"
	"github.com/stretchr/testify/assert"
)

type requestRecoveryParams struct {
	producer  uof.Producer
	timestamp int
	requestID int
}

type recoveryApiMock struct {
	calls chan requestRecoveryParams
}

func (m *recoveryApiMock) RequestRecovery(producer uof.Producer, timestamp int, requestID int) error {
	m.calls <- requestRecoveryParams{
		producer:  producer,
		timestamp: timestamp,
		requestID: requestID,
	}
	return nil
}

func TestRecoveryTimestamp(t *testing.T) {
	cs := uof.CurrentTimestamp()
	rp := &recoveryProducer{
		producer:       uof.ProducerLiveOdds,
		aliveTimestamp: cs,
	}
	assert.Equal(t, cs, rp.recoveryTimestamp())
	rp.aliveTimestamp = cs - rp.producer.RecoveryWindow() + 1
	assert.Equal(t, rp.aliveTimestamp, rp.recoveryTimestamp())
	rp.aliveTimestamp = cs - rp.producer.RecoveryWindow()
	assert.Equal(t, int(0), rp.recoveryTimestamp())
}

func TestRecoveryStateMachine(t *testing.T) {
	// setup
	timestamp := uof.CurrentTimestamp() - 10*1000
	var ps uof.ProducersChange
	ps.Add(uof.ProducerPrematch, timestamp)
	ps.Add(uof.ProducerLiveOdds, timestamp+1)
	m := &recoveryApiMock{calls: make(chan requestRecoveryParams, 16)}
	r := newRecovery(m, ps)

	// 0. initilay all producers are down
	for _, p := range r.producers {
		assert.Equal(t, uof.ProducerStatusDown, p.status)
	}

	// 1. connection up, triggers recovery requests for all producers
	r.connectionUp()
	// all produers status are changed to in recovery
	for _, p := range r.producers {
		assert.Equal(t, uof.ProducerStatusInRecovery, p.status)
	}
	// two recovery requests are sent
	recoveryRequestPrematch := <-m.calls
	recoveryRequestLive := <-m.calls
	if recoveryRequestPrematch.producer == uof.ProducerLiveOdds {
		// reorder because order is not garanteed (called in goroutines)
		recoveryRequestPrematch, recoveryRequestLive = recoveryRequestLive, recoveryRequestPrematch
	}
	// prematch producer request
	assert.Equal(t, uof.ProducerPrematch, recoveryRequestPrematch.producer)
	assert.Equal(t, timestamp, recoveryRequestPrematch.timestamp)
	prematch := r.find(uof.ProducerPrematch)
	assert.Equal(t, prematch.requestID, recoveryRequestPrematch.requestID)
	assert.Equal(t, uof.ProducerStatusInRecovery, prematch.status)
	// live producer request
	assert.Equal(t, uof.ProducerLiveOdds, recoveryRequestLive.producer)
	assert.Equal(t, timestamp+1, recoveryRequestLive.timestamp)
	live := r.find(uof.ProducerLiveOdds)
	assert.Equal(t, live.requestID, recoveryRequestLive.requestID)
	assert.Equal(t, uof.ProducerStatusInRecovery, live.status)

	// 2. snapshot complete, changes status to active for that producer
	assert.Equal(t, uof.ProducerStatusInRecovery, prematch.status)
	r.snapshotComplete(uof.ProducerPrematch, recoveryRequestPrematch.requestID)
	assert.Equal(t, uof.ProducerStatusActive, prematch.status)

	// 3. on alive, updates timestamp for that producer
	r.alive(uof.ProducerPrematch, timestamp+2, 1)
	assert.Equal(t, timestamp+2, prematch.aliveTimestamp)
	assert.Equal(t, timestamp+1, live.aliveTimestamp)

	// 4. on alive with subscribed = 0, forces recovery request for that producer
	// changes status to in recovery until snapshot complete
	assert.Equal(t, uof.ProducerStatusActive, prematch.status)
	r.alive(uof.ProducerPrematch, timestamp+3, 0)
	assert.Equal(t, uof.ProducerStatusInRecovery, prematch.status)
	recoveryRequestPrematch = <-m.calls
	assert.Equal(t, uof.ProducerPrematch, recoveryRequestPrematch.producer)
	assert.Equal(t, timestamp+2, recoveryRequestPrematch.timestamp)
	assert.Equal(t, prematch.requestID, recoveryRequestPrematch.requestID)

	// 5. snapshot complete, changes status to active for that producer
	assert.Equal(t, uof.ProducerStatusInRecovery, prematch.status)
	r.snapshotComplete(uof.ProducerPrematch, recoveryRequestPrematch.requestID)
	assert.Equal(t, uof.ProducerStatusActive, prematch.status)

	// 6. connection down, retruns to the start of the cycle
	r.connectionDown()
	for _, p := range r.producers {
		assert.Equal(t, uof.ProducerStatusDown, p.status)
	}
}

func TestRecoveryRequests(t *testing.T) {
	timestamp := uof.CurrentTimestamp() - 10*1000
	var ps uof.ProducersChange
	ps.Add(uof.ProducerPrematch, timestamp)
	ps.Add(uof.ProducerLiveOdds, timestamp+1)

	m := &recoveryApiMock{calls: make(chan requestRecoveryParams, 16)}
	r := newRecovery(m, ps)
	in := make(chan *uof.Message)
	out := make(chan *uof.Message, 16)
	errc := make(chan error, 16)
	go r.loop(in, out, errc)

	// 1. connection status triggers recovery requests
	in <- uof.NewConnnectionMessage(uof.ConnectionStatusUp)
	recoveryRequestPrematch := <-m.calls
	recoveryRequestLive := <-m.calls
	if recoveryRequestPrematch.producer == uof.ProducerLiveOdds {
		// order is not garanteed
		recoveryRequestPrematch, recoveryRequestLive = recoveryRequestLive, recoveryRequestPrematch
	}
	assert.Equal(t, uof.ProducerPrematch, recoveryRequestPrematch.producer)
	assert.Equal(t, timestamp, recoveryRequestPrematch.timestamp)
	prematch := r.find(uof.ProducerPrematch)
	assert.Equal(t, prematch.requestID, recoveryRequestPrematch.requestID)
	assert.Equal(t, uof.ProducerStatusInRecovery, prematch.status)

	assert.Equal(t, uof.ProducerLiveOdds, recoveryRequestLive.producer)
	assert.Equal(t, timestamp+1, recoveryRequestLive.timestamp)
	live := r.find(uof.ProducerLiveOdds)
	assert.Equal(t, live.requestID, recoveryRequestLive.requestID)
	assert.Equal(t, uof.ProducerStatusInRecovery, live.status)
	<-out // skip connection status message
	producersChangeMessage := <-out
	// check out message
	assert.Equal(t, uof.MessageTypeProducersChange, producersChangeMessage.Type)
	assert.Equal(t, uof.MessageScopeSystem, producersChangeMessage.Scope)
	assert.Equal(t, uof.ProducerPrematch, producersChangeMessage.Producers[0].Producer)
	assert.Equal(t, uof.ProducerStatusInRecovery, producersChangeMessage.Producers[0].Status)
	assert.Equal(t, uof.ProducerLiveOdds, producersChangeMessage.Producers[1].Producer)
	assert.Equal(t, uof.ProducerStatusInRecovery, producersChangeMessage.Producers[1].Status)

	// 2. snapshot complete for the prematch is received
	in <- &uof.Message{
		Header: uof.Header{Type: uof.MessageTypeSnapshotComplete},
		Body: uof.Body{SnapshotComplete: &uof.SnapshotComplete{
			Producer:  uof.ProducerPrematch,
			RequestID: recoveryRequestPrematch.requestID},
		},
	}
	<-out //snapshot complete
	producersChangeMessage = <-out
	// status of the prematch is changed to the active
	assert.Equal(t, prematch.requestID, 0)
	assert.Equal(t, uof.ProducerStatusActive, prematch.status)
	assert.Equal(t, live.requestID, recoveryRequestLive.requestID)
	assert.Equal(t, uof.ProducerStatusInRecovery, live.status)
	// check out message
	assert.Equal(t, uof.ProducerPrematch, producersChangeMessage.Producers[0].Producer)
	assert.Equal(t, uof.ProducerStatusActive, producersChangeMessage.Producers[0].Status)
	assert.Equal(t, uof.ProducerLiveOdds, producersChangeMessage.Producers[1].Producer)
	assert.Equal(t, uof.ProducerStatusInRecovery, producersChangeMessage.Producers[1].Status)

	// 3. alive message
	in <- &uof.Message{
		Header: uof.Header{Type: uof.MessageTypeAlive},
		Body: uof.Body{Alive: &uof.Alive{
			Producer:   uof.ProducerPrematch,
			Timestamp:  timestamp + 2,
			Subscribed: 1,
		},
		},
	}

	// 4. alive with subscribed=0 triggers recovery request
	in <- &uof.Message{
		Header: uof.Header{Type: uof.MessageTypeAlive},
		Body: uof.Body{Alive: &uof.Alive{
			Producer:   uof.ProducerPrematch,
			Timestamp:  timestamp + 3,
			Subscribed: 0,
		},
		},
	}
	recoveryRequestPrematch = <-m.calls
	assert.Equal(t, uof.ProducerPrematch, recoveryRequestPrematch.producer)
	<-out //alive messages
	<-out
	producersChangeMessage = <-out
	// check out message, both producers are again in recovery
	assert.Equal(t, uof.ProducerPrematch, producersChangeMessage.Producers[0].Producer)
	assert.Equal(t, uof.ProducerStatusInRecovery, producersChangeMessage.Producers[0].Status)
	assert.Equal(t, uof.ProducerLiveOdds, producersChangeMessage.Producers[1].Producer)
	assert.Equal(t, uof.ProducerStatusInRecovery, producersChangeMessage.Producers[1].Status)
}
