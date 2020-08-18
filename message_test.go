package uof

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessageParseRoutingKeys(t *testing.T) {
	data := []struct {
		key string
		rm  Message
	}{
		{
			key: "hi.-.live.bet_cancel.21.sr:match.13073610.-",
			rm: Message{
				Header: Header{
					Type:     MessageTypeBetCancel,
					Scope:    MessageScopeLive,
					Priority: MessagePriorityHigh,
					SportID:  21,
					EventURN: "sr:match:13073610",
					EventID:  13073610,
				},
			},
		},
		{
			key: "hi.pre.-.odds_change.1.sr:match.1234.-",
			rm: Message{
				Header: Header{
					Type:     MessageTypeOddsChange,
					Scope:    MessageScopePrematch,
					Priority: MessagePriorityHigh,
					SportID:  1,
					EventURN: "sr:match:1234",
					EventID:  1234,
				},
			},
		},
		{
			key: "hi.virt.-.odds_change.7.vf:match.12345.-",
			rm: Message{
				Header: Header{
					Type:     MessageTypeOddsChange,
					Scope:    MessageScopeVirtuals,
					Priority: MessagePriorityHigh,
					SportID:  7,
					EventURN: "vf:match:12345",
					EventID:  -3160336,
				},
			},
		},
		{
			key: "-.-.-.alive.-.-.-.-",
			rm: Message{
				Header: Header{
					Type:     MessageTypeAlive,
					Scope:    MessageScopeSystem,
					Priority: MessagePriorityLow,
				},
			},
		},
		{
			key: "-.-.-.snapshot_complete.-.-.-",
			rm: Message{
				Header: Header{
					Type:     MessageTypeSnapshotComplete,
					Scope:    MessageScopeSystem,
					Priority: MessagePriorityLow,
				},
			},
		},
		{
			key: "hi.-.live.odds_change.4.sr:match.11784628",
			rm: Message{
				Header: Header{
					Type:     MessageTypeOddsChange,
					Scope:    MessageScopeLive,
					Priority: MessagePriorityHigh,
					SportID:  4,
					EventURN: "sr:match:11784628",
					EventID:  11784628,
				},
			},
		},
		{
			key: "lo.pre.live.bet_settlement.8.sr:match.12.-",
			rm: Message{
				Header: Header{
					Type:     MessageTypeBetSettlement,
					Scope:    MessageScopePrematchAndLive,
					Priority: MessagePriorityLow,
					SportID:  8,
					EventURN: "sr:match:12",
					EventID:  12,
				},
			},
		},
	}

	for _, d := range data {
		rm, err := NewQueueMessage(d.key, nil)
		assert.Nil(t, err)
		assert.Equal(t, d.rm.Scope, rm.Scope)
		assert.Equal(t, d.rm.Type, rm.Type)
		assert.Equal(t, d.rm.Priority, rm.Priority)
		assert.Equal(t, d.rm.SportID, rm.SportID)
		assert.Equal(t, d.rm.EventURN, rm.EventURN)
		assert.Equal(t, d.rm.EventID, rm.EventID)
		assert.Equal(t, LangNone, rm.Lang)
	}

	_, err := NewQueueMessage("...", nil)
	assert.Error(t, err)

	_, err = NewQueueMessage("hi.-.live.unknown.4.sr:match.11784628", nil)
	assert.Error(t, err)

	_, err = NewQueueMessage("hi.-.live.odds_change.4.sr:match.pero", nil)
	assert.Error(t, err)
}

func TestMessageTypeParse(t *testing.T) {
	var y MessageType
	y.Parse("alive")
	assert.Equal(t, MessageTypeAlive, y)
}

func TestMessageWithRawMarshal(t *testing.T) {
	m := &Message{
		Header: Header{Type: MessageTypeAlive, Scope: MessageScopeSystem, Priority: MessagePriorityLow, ReceivedAt: 12345, Producer: ProducerPrematch, Timestamp: 12340},
		Raw:    []byte(`<alive product="3" timestamp="12340" subscribed="1"/>`),
		Body: Body{
			Alive: &Alive{
				Producer:   ProducerPrematch,
				Timestamp:  12340,
				Subscribed: 1,
			},
		},
	}

	buf := m.Marshal()
	expected := []byte(`{"type":64,"scope":4,"receivedAt":12345,"producer":3,"timestamp":12340}
<alive product="3" timestamp="12340" subscribed="1"/>`)
	assert.Equal(t, expected, buf)

	m2 := &Message{}
	err := m2.Unmarshal(buf)
	assert.NoError(t, err)
	assert.Equal(t, m, m2)
}

func TestMessageWithoutRaw(t *testing.T) {
	m := &Message{
		Header: Header{Type: MessageTypeConnection, Scope: MessageScopeSystem, Priority: MessagePriorityLow, ReceivedAt: 12345},
		Body: Body{
			Connection: &Connection{
				Status: ConnectionStatusDown,
			},
		},
	}

	buf := m.Marshal()
	expected := []byte(`{"type":66,"scope":4,"receivedAt":12345,"connection":{"status":1}}`)
	assert.Equal(t, expected, buf)

	m2 := &Message{}
	err := m2.Unmarshal(buf)
	assert.NoError(t, err)
	assert.Equal(t, m, m2)
}

func TestUID(t *testing.T) {
	data := []struct {
		m   Message
		uid int
	}{
		{
			Message{
				Header: Header{
					Type: MessageTypePlayer,
					Lang: LangHR,
				},
				Body: Body{
					Player: &Player{ID: 0x12345},
				},
			},
			0x1234509,
		},
		{
			Message{
				Header: Header{
					Type: MessageTypePlayer,
					Lang: LangTR,
				},
				Body: Body{
					Player: &Player{ID: 0x007fffffffffffff},
				},
			},
			0x7fffffffffffff29,
		},
		{
			Message{
				Header: Header{
					Type: MessageTypeFixture,
					Lang: LangIT,
				},
				Body: Body{
					Fixture: &Fixture{ID: -0x123},
				},
			},
			-0x1232c,
		},
		{
			Message{
				Header: Header{
					Type: MessageTypeAlive,
				},
			},
			0,
		},
	}

	for i, d := range data {
		assert.Equal(t, d.uid, d.m.UID(),
			"case: %d, actual %x", i, d.m.UID())
	}
}

func TestUIDWithLang(t *testing.T) {
	data := []struct {
		id   int
		lang Lang
		uid  int
	}{
		{0x01, LangHR, 0x0109},
		{0x007fffffffffffff, LangTR, 0x7fffffffffffff29},
		{-0x123, LangIT, -0x1232c},
	}

	for _, d := range data {
		assert.Equal(t, d.uid, UIDWithLang(d.id, d.lang))
	}
}

func TestNewMessage(t *testing.T) {
	m := NewConnnectionMessage(ConnectionStatusUp)
	assert.True(t, m.Is(MessageTypeConnection))

	m = NewPlayerMessage(LangEN, nil, 0)
	assert.True(t, m.Is(MessageTypePlayer))

	m = NewMarketsMessage(LangEN, nil, 0)
	assert.True(t, m.Is(MessageTypeMarkets))

	m = NewProducersChangeMessage(nil)
	assert.True(t, m.Is(MessageTypeProducersChange))

	m = NewFixtureMessage(LangEN, Fixture{}, 0)
	assert.True(t, m.Is(MessageTypeFixture))

	m.NewFixtureMessage(LangEN, Fixture{})
	assert.True(t, m.Is(MessageTypeFixture))
}

func TestUnpackFail(t *testing.T) {
	// void_reason below whould be int value
	buf := []byte(`<bet_cancel end_time="1564598513000" event_id="sr:match:18941600" product="1" start_time="1564597838000" timestamp="1564602448841">
	<market name="1st half - 1st goal" id="62" void_reason="int"/>
	</bet_cancel>`)

	_, err := NewQueueMessage("hi.pre.-.bet_cancel.1.sr:match.1234.-", buf)
	assert.Error(t, err)
	assert.Equal(t, `NOTICE uof error op: message.unpack, inner: strconv.ParseInt: parsing "int": invalid syntax`, err.Error())

	// height should be int
	buf = []byte(`
	<player_profile>
    	<player height="int" />
	</player_profile>
	`)
	_, err = NewAPIMessage(LangEN, MessageTypePlayer, buf)
	assert.Error(t, err)
	assert.Equal(t, `NOTICE uof error op: message.unpack, inner: strconv.ParseInt: parsing "int": invalid syntax`, err.Error())

	var m Message
	err = m.Unmarshal(nil)
	assert.Error(t, err)

	m.Raw = []byte{}
	m.Type = -1
	err = m.unpack()
	assert.Error(t, err)
	assert.Equal(t, `NOTICE uof error op: message.unpack, inner: unknown message type -1`, err.Error())
}

func TestEnrichHeaderAfterUnpack(t *testing.T) {
	// check if Producer and Timestamp get copied to Header after Unmarshal
	// all messages have the same producer and the same timestamp in ms (1234578910111)
	buf := []struct {
		key string
		raw []byte
	}{
		// snapshot_complete
		struct {
			key string
			raw []byte
		}{
			key: "-.-.-.snapshot_complete.-.-.-",
			raw: []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
			<snapshot_complete request_id="1234" timestamp="1234578910111" product="1"/>`),
		},
		// alive
		struct {
			key string
			raw []byte
		}{
			key: "-.-.-.alive.-.-.-.-",
			raw: []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
			<alive product="1" timestamp="1234578910111" subscribed="1"/>`),
		},
		// fixture_change
		struct {
			key string
			raw []byte
		}{
			key: "hi.pre.live.fixture_change.1.sr:match.18001015.-",
			raw: []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
			<fixture_change start_time="1557255600000" product="1" event_id="sr:match:18001015" timestamp="1234578910111"/>`),
		},
		// bet_stop
		struct {
			key string
			raw []byte
		}{
			key: "hi.-.live.bet_stop.1.sr:match.18001015.-",
			raw: []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
			<bet_stop groups="all" market_status="0" product="1" event_id="sr:match:18001015" timestamp="1234578910111"/>`),
		},
		// odds_change
		struct {
			key string
			raw []byte
		}{
			key: "hi.-.live.odds_change.1.sr:match.18001015.-",
			raw: []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
		<odds_change product="1" event_id="sr:match:18001015" timestamp="1234578910111">
			<sport_event_status status="0" match_status="0"/>
			<odds>
				<market status="1" id="1">
					<outcome id="1" odds="2.08" probabilities="0.450633" active="1"/>
					<outcome id="2" odds="3.31" probabilities="0.277172" active="1"/>
					<outcome id="3" odds="3.37" probabilities="0.272194" active="1"/>
				</market>
			</odds>
		</odds_change>`),
		},
		// bet_settlement
		struct {
			key string
			raw []byte
		}{
			key: "hi.-.live.bet_settlement.1.sr:match.18001015.-",
			raw: []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
		<bet_settlement certainty="1" product="1" event_id="sr:match:13369905" timestamp="1234578910111">
			<outcomes>
				<market id="6">
					<outcome id="4" result="1"/>
					<outcome id="5" result="0"/>
				</market>
			</outcomes>
		</bet_settlement>`),
		},
		// rollback_bet_settlement
		struct {
			key string
			raw []byte
		}{
			key: "hi.-.live.rollback_bet_settlement.1.sr:match.18001015.-",
			raw: []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
			<rollback_bet_settlement product="1" event_id="sr:match:18001015" timestamp="1234578910111">
				<market id="38" specifiers="goalnr=1|type=live"/>
			</rollback_bet_settlement>`),
		},
		// bet_cancel
		struct {
			key string
			raw []byte
		}{
			key: "hi.-.live.bet_cancel.1.sr:match.18001015.-",
			raw: []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
		<bet_cancel end_time="1564598513000" event_id="sr:match:18001015" product="1" start_time="1564597838000" timestamp="1234578910111">
			<market name="1st half - 1st goal" id="62" specifier="goalnr=1" void_reason="12"/>
		</bet_cancel>`),
		},
		// rollback_bet_cancel
		struct {
			key string
			raw []byte
		}{
			key: "hi.-.live.rollback_bet_cancel.1.sr:match.18001015.-",
			raw: []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
			<rollback_bet_cancel event_id="sr:match:4444" product="1" timestamp="1234578910111">
				<market id="48" specifiers="score=41.5"/>
			</rollback_bet_cancel>`),
		},
	}

	for _, m := range buf {
		qm, err := NewQueueMessage(m.key, m.raw)
		assert.Nil(t, err)
		if err != nil {
			t.Logf(err.Error())
		}
		assert.NotNil(t, qm)
		assert.Equal(t, ProducerLiveOdds, qm.Producer)
		assert.Equal(t, 1234578910111, qm.Timestamp)
		assert.Equal(t, getTsFromMsg(qm), qm.Timestamp)
		assert.Equal(t, getProducerFromMsg(qm), qm.Producer)
		t.Logf("prod: %s ts: %d msg type: %s \n", qm.Producer, qm.Timestamp, qm.Type.String())
	}
}

// get timestamp from embedded message type
func getTsFromMsg(msg *Message) int {
	switch msg.Type {
	case MessageTypeOddsChange:
		return msg.OddsChange.Timestamp
	case MessageTypeFixtureChange:
		return msg.FixtureChange.Timestamp
	case MessageTypeBetStop:
		return msg.BetStop.Timestamp
	case MessageTypeBetSettlement:
		return msg.BetSettlement.Timestamp
	case MessageTypeRollbackBetSettlement:
		return msg.RollbackBetSettlement.Timestamp
	case MessageTypeBetCancel:
		return msg.BetCancel.Timestamp
	case MessageTypeRollbackBetCancel:
		return msg.RollbackBetCancel.Timestamp
	case MessageTypeSnapshotComplete:
		return msg.SnapshotComplete.Timestamp
	case MessageTypeAlive:
		return msg.Alive.Timestamp
	}
	return 0
}

// get producer from embedded message type
func getProducerFromMsg(msg *Message) Producer {
	switch msg.Type {
	case MessageTypeOddsChange:
		return msg.OddsChange.Producer
	case MessageTypeFixtureChange:
		return msg.FixtureChange.Producer
	case MessageTypeBetStop:
		return msg.BetStop.Producer
	case MessageTypeBetSettlement:
		return msg.BetSettlement.Producer
	case MessageTypeRollbackBetSettlement:
		return msg.RollbackBetSettlement.Producer
	case MessageTypeBetCancel:
		return msg.BetCancel.Producer
	case MessageTypeRollbackBetCancel:
		return msg.RollbackBetCancel.Producer
	case MessageTypeSnapshotComplete:
		return msg.SnapshotComplete.Producer
	case MessageTypeAlive:
		return msg.Alive.Producer
	}
	return ProducerUnknown
}
