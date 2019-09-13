package uof

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

type Header struct {
	Type       MessageType     `json:"type,omitempty"`
	Scope      MessageScope    `json:"scope,omitempty"`
	Priority   MessagePriority `json:"priority,omitempty"`
	Lang       Lang            `json:"lang,omitempty"`
	SportID    int             `json:"sportID,omitempty"`
	EventID    int             `json:"eventID,omitempty"`
	EventURN   URN             `json:"eventURN,omitempty"`
	ReceivedAt int64           `json:"receivedAt,omitempty"`
}

type Body struct {
	Alive                 *Alive                 `json:"alive,omitempty"`
	BetCancel             *BetCancel             `json:"betCancel,omitempty"`
	RollbackBetSettlement *RollbackBetSettlement `json:"rollbackBetSettlement,omitempty"`
	RollbackBetCancel     *RollbackBetCancel     `json:"rollbackBetCancel,omitempty"`
	SnapshotComplete      *SnapshotComplete      `json:"snapshotComplete,omitempty"`
	OddsChange            *OddsChange            `json:"oddsChange,omitempty"`
	FixtureChange         *FixtureChange         `json:"fixtureChange,omitempty"`
	BetSettlement         *BetSettlement         `json:"betSettlement,omitempty"`
	BetStop               *BetStop               `json:"betStop,omitempty"`
	Fixture               *Fixture               `json:"fixture,omitempty"`
	Markets               MarketDescriptions     `json:"markets,omitempty"`
	Player                *Player                `json:"player,omitempty"`
	Connection            *Connection            `json:"connection,omitempty"`
	Producers             ProducersChange        `json:"producerChange,omitempty"`
}

type Message struct {
	Header `json:",inline"`
	Raw    []byte `json:"-"`
	Body   `json:",inline"`
}

var uniqTimestamp func() int64 // ensures unique timestamp value

func init() {
	// init makes clousure for lastTs and mu
	lastTs := CurrentTimestamp()
	var mu sync.Mutex

	uniqTimestamp = func() int64 {
		mu.Lock()
		defer mu.Unlock()
		ts := CurrentTimestamp()
		if ts <= lastTs {
			ts += 1
		}
		lastTs = ts
		return ts
	}
}

// CurrentTimestamp in milliseconds
func CurrentTimestamp() int64 {
	return timeToTimestamp(time.Now())
}
func timeToTimestamp(t time.Time) int64 {
	return t.UnixNano() / 1e6
}

func NewQueueMessage(routingKey string, body []byte) (*Message, error) {
	r := &Message{
		Header: Header{ReceivedAt: uniqTimestamp()},
		Raw:    body,
	}
	if err := r.parseRoutingKey(routingKey); err != nil {
		return nil, err
	}
	return r, r.unpack()
}

func (m *Message) parseRoutingKey(routingKey string) error {
	p := strings.Split(routingKey, ".")
	if len(p) < 7 {
		err := fmt.Errorf("unknown routing key: %s", routingKey)
		return errors.WithStack(err)
	}
	part := func(i int) string {
		if len(p) > i && p[i] != "-" {
			return p[i]
		}
		return ""
	}
	priority := part(0)
	prematchInterest := part(1)
	liveInterest := part(2)
	messageType := part(3)
	sportID := part(4)
	eventURN := part(5)
	eventID := part(6)
	//nodeID := part(7)  // currently unused

	m.Priority.Parse(priority)
	m.Type.Parse(messageType)
	m.Scope.Parse(prematchInterest, liveInterest)

	if m.Type == MessageTypeUnknown {
		err := fmt.Errorf("unknown message type for routing key: %s", routingKey)
		return errors.WithStack(err)
	}

	if eventID != "" {
		m.EventID, _ = strconv.Atoi(eventID)
	}
	if sportID != "" {
		m.SportID, _ = strconv.Atoi(sportID)
	}
	if eventURN != "" && sportID != "" {
		m.EventURN = URN(eventURN + ":" + eventID)
	}

	return nil
}

func (m *Message) unpack() error {
	if m.Raw == nil {
		return nil
	}
	var err error

	unmarshal := func(i interface{}) {
		err = xml.Unmarshal(m.Raw, i)
	}

	switch m.Type {
	case MessageTypeAlive:
		m.Alive = &Alive{}
		unmarshal(m.Alive)
	case MessageTypeBetCancel:
		m.BetCancel = &BetCancel{}
		unmarshal(m.BetCancel)
	case MessageTypeBetSettlement:
		m.BetSettlement = &BetSettlement{}
		unmarshal(m.BetSettlement)
	case MessageTypeBetStop:
		m.BetStop = &BetStop{}
		unmarshal(m.BetStop)
	case MessageTypeFixtureChange:
		m.FixtureChange = &FixtureChange{}
		unmarshal(m.FixtureChange)
	case MessageTypeOddsChange:
		m.OddsChange = &OddsChange{}
		unmarshal(m.OddsChange)
	case MessageTypeRollbackBetSettlement:
		m.RollbackBetSettlement = &RollbackBetSettlement{}
		unmarshal(m.RollbackBetSettlement)
	case MessageTypeRollbackBetCancel:
		m.RollbackBetCancel = &RollbackBetCancel{}
		unmarshal(m.RollbackBetCancel)
	case MessageTypeSnapshotComplete:
		m.SnapshotComplete = &SnapshotComplete{}
		unmarshal(m.SnapshotComplete)
	case MessageTypeFixture:
		fr := FixtureRsp{}
		unmarshal(&fr)
		m.Fixture = &fr.Fixture
	case MessageTypeMarkets:
		md := &MarketsRsp{}
		unmarshal(md)
		m.Markets = md.Markets
	case MessageTypePlayer:
		pp := PlayerProfile{}
		unmarshal(&pp)
		m.Player = &pp.Player
	default:
		return fmt.Errorf("unknown message type %d", m.Type)
	}
	return err
}

func NewMarketsMessage(lang Lang, body []byte) (*Message, error) {
	m := &Message{
		Header: Header{
			Type:       MessageTypeMarkets,
			Lang:       lang,
			ReceivedAt: uniqTimestamp(),
		},
		Raw: body,
	}
	return m, m.unpack()
}

func NewPlayerMessage(lang Lang, body []byte) (*Message, error) {
	m := &Message{
		Header: Header{
			Type:       MessageTypePlayer,
			Lang:       lang,
			ReceivedAt: uniqTimestamp(),
		},
		Raw: body,
	}
	return m, m.unpack()
}

func NewConnnectionMessage(status ConnectionStatus) *Message {
	ts := uniqTimestamp()
	return &Message{
		Header: Header{
			Type:       MessageTypeConnection,
			Scope:      MessageScopeSystem,
			ReceivedAt: ts,
		},
		Body: Body{
			Connection: &Connection{
				Status:    status,
				Timestamp: ts,
			},
		},
	}
}

func NewProducersChangeMessage(pc ProducersChange) *Message {
	return &Message{
		Header: Header{
			Type:       MessageTypeProducersChange,
			Scope:      MessageScopeSystem,
			ReceivedAt: uniqTimestamp(),
		},
		Body: Body{
			Producers: pc,
		},
	}
}

func NewFixtureMessage(lang Lang, x Fixture) *Message {
	return &Message{
		Header: Header{
			Type:       MessageTypeFixture,
			EventURN:   x.URN,
			EventID:    x.ID,
			Lang:       lang,
			ReceivedAt: uniqTimestamp(),
		},
		Body: Body{
			Fixture: &x,
		},
	}
}

// // AsFixture transforms fixture change message to the fixture message
// // Takes attributes from the first message with date from api.
// func (m *Message) AsFixture(lang Lang, body []byte) (*Message, error) {
// 	if m.Type != MessageTypeFixtureChange {
// 		return nil, fmt.Errorf("wrong parent message type")
// 	}
// 	c := &Message{
// 		Header: m.Header,
// 		Raw:    body,
// 	}
// 	c.Type = MessageTypeFixture
// 	c.Lang = lang
// 	return c, c.unpack()
// }

func (m *Message) NewFixtureMessage(lang Lang, f Fixture) *Message {
	c := &Message{
		Header: m.Header,
	}
	c.Type = MessageTypeFixture
	c.Lang = lang
	c.Fixture = &f
	return c
}

const separator = byte(10)

func (m Message) Marshal() []byte {
	if m.Raw == nil {
		buf, _ := json.Marshal(m)
		return buf
	}
	buf, _ := json.Marshal(m.Header)
	buf = append(buf, separator)
	return append(buf, m.Raw...)
}

func (m *Message) Unmarshal(buf []byte) error {
	parts := bytes.SplitN(buf, []byte{separator}, 2)
	if err := json.Unmarshal(parts[0], m); err != nil {
		return err
	}
	if len(parts) < 2 {
		return nil
	}
	m.Raw = parts[1]
	return m.unpack()
}
