package uof

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type Message struct {
	Type                  MessageType            `json:"type,omitempty"`
	Scope                 MessageScope           `json:"scope,omitempty"`
	Priority              MessagePriority        `json:"priority,omitempty"`
	Lang                  Lang                   `json:"lang,omitempty"`
	SportID               int                    `json:"sportID,omitempty"`
	EventID               int                    `json:"eventID,omitempty"`
	EventURN              URN                    `json:"eventURN,omitempty"`
	ReceivedAt            int64                  `json:"receivedAt,omitempty"`
	Body                  []byte                 `json:"-"`
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
}

func NewQueueMessage(routingKey string, timestamp int64, body []byte) (*Message, error) {
	r := &Message{
		Body:       body,
		ReceivedAt: timestamp,
	}
	if err := r.parseRoutingKey(routingKey); err != nil {
		return nil, err
	}
	return r, nil //r.unpack()
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
	if m.Body == nil {
		return nil
	}
	var err error

	unmarshal := func(i interface{}) {
		err = xml.Unmarshal(m.Body, i)
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
		Type: MessageTypeMarkets,
		Body: body,
		Lang: lang,
		//ReceivedAt: uniqTimestamp(),
	}
	return m, m.unpack()
}

func NewPlayerMessage(lang Lang, body []byte) (*Message, error) {
	m := &Message{
		Type: MessageTypePlayer,
		Body: body,
		Lang: lang,
		//ReceivedAt: uniqTimestamp(),
	}
	return m, m.unpack()
}

// AsFixture transforms fixture change message to the fixture message
// Takes attributes from the first message with date from api.
func (m *Message) AsFixture(lang Lang, body []byte) (*Message, error) {
	if m.Type != MessageTypeFixtureChange {
		return nil, fmt.Errorf("wrong parent message type")
	}
	c := &Message{
		Type:       MessageTypeFixture,
		Priority:   m.Priority,
		Scope:      m.Scope,
		SportID:    m.SportID,
		EventID:    m.EventID,
		EventURN:   m.EventURN,
		ReceivedAt: m.ReceivedAt,
		Lang:       lang,
		Body:       body,
	}
	return c, c.unpack()
}
