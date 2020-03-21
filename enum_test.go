package uof

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProducer(t *testing.T) {
	assert.Equal(t, "pre", ProducerPrematch.String())
	assert.Equal(t, "pre", Producer(3).Code())
	assert.Equal(t, "Ctrl", Producer(3).Name())
	assert.Equal(t, "Betradar Ctrl", Producer(3).Description())
	assert.Equal(t, InvalidName, Producer(-1).String())
	assert.Equal(t, InvalidName, Producer(-1).Name())
	assert.Equal(t, InvalidName, Producer(-1).Description())
	assert.Equal(t, 0, Producer(-1).RecoveryWindow())

	assert.Equal(t, 259200000, Producer(3).RecoveryWindow())
	assert.True(t, Producer(3).Prematch())

	assert.True(t, Producer(3).Sports())
	assert.False(t, Producer(3).Virtuals())

	assert.False(t, Producer(11).Sports())
	assert.True(t, Producer(11).Virtuals())
}

func TestURN(t *testing.T) {
	u := URN("sr:match:123")
	assert.Equal(t, 123, u.ID())
	//assert.Equal(t, URNTypeMatch, u.Type())
	assert.Equal(t, URN("sr:match:123"), NewEventURN(123))
	assert.Equal(t, "sr:match:123", URN("sr:match:123").String())

	assert.Equal(t, 0, URN("").ID())
	assert.True(t, URN("").Empty())
	assert.Equal(t, 0, URN("").EventID())
	assert.Equal(t, 0, URN("pero").ID())

	//assert.Equal(t, URNTypeUnknown, URN("").Type())
	//assert.Equal(t, URNTypeUnknown, URN("pero").Type())
	assert.Equal(t, 0, URN("pero").EventID())

	u.Parse("123")
	assert.Equal(t, URN("sr:match:123"), u)
	assert.Equal(t, 123, u.EventID())
}

func TestLanguage(t *testing.T) {
	var l Lang
	l.Parse("hr")
	assert.Equal(t, LangHR, l)
	assert.Equal(t, "hr", l.Code())
	assert.Equal(t, "hr", l.String())
	assert.Equal(t, "Croatian", l.Name())

	ls := Languages("hr,en,de")
	assert.Len(t, ls, 3)
	assert.Equal(t, LangHR, ls[0])
	assert.Equal(t, LangEN, ls[1])
	assert.Equal(t, LangDE, ls[2])

	assert.Equal(t, "", Lang(-128).Code())
	assert.Equal(t, "", Lang(-128).Name())
}

func TestMessageTypes(t *testing.T) {
	for i, n := range messageTypeNames {
		m := messageTypes[i]
		assert.Equal(t, m.String(), n)

		var m2 MessageType
		m2.Parse(n)
		assert.Equal(t, m, m2)
	}

	assert.Equal(t, InvalidName, MessageType(127).String())

	assert.Equal(t, MessageKindEvent, MessageType(1).Kind())
	assert.Equal(t, MessageKindLexicon, MessageType(32).Kind())
	assert.Equal(t, MessageKindSystem, MessageType(64).Kind())
}

func TestURNEventID(t *testing.T) {
	assert.Equal(t, -0xff01, URN("sr:stage:255").EventID())
	assert.Equal(t, -0xff02, URN("sr:season:255").EventID())
	assert.Equal(t, -0xff1C, URN("vti:tournament:255").EventID())
	assert.Equal(t, 0, URN("pero:zdero:255").EventID())

	data := []struct {
		u  string
		id int
	}{
		{"sr:match:127", 127},

		{"sr:stage:127", -0x7f01},
		{"sr:season:255", -0xff02},
		{"sr:tournament:255", -0xff03},
		{"sr:simple_tournament:255", -0xff04},

		{"test:match:255", -0xff0F},

		{"vf:match:255", -0xff10},
		{"vf:season:255", -0xff11},
		{"vf:tournament:255", -0xff12},

		{"vbl:match:255", -0xff13},
		{"vbl:season:255", -0xff14},
		{"vbl:tournament:255", -0xff15},

		{"vto:match:255", -0xff16},
		{"vto:season:255", -0xff17},
		{"vto:tournament:255", -0xff18},

		{"vdr:stage:255", -0xff19},
		{"vhc:stage:255", -0xff1A},

		{"vti:match:255", -0xff1B},
		{"vti:tournament:255", -0xff1C},

		{"wns:draw:255", -0xff1D},
		// invalid
		{"pero:zdero:255", 0},
	}
	for _, d := range data {
		assert.Equal(t, d.id, URN(d.u).EventID())
	}
}

// func TestURNType(t *testing.T) {
// 	data := []struct {
// 		u string
// 		t int8
// 	}{
// 	{"sr:match:127", }
// 	}
// }

func TestProducersChange(t *testing.T) {
	var pc ProducersChange
	pc.Add(ProducerLiveOdds, 123)
	pc.Add(ProducerPrematch, 456)
	assert.Len(t, pc, 2)
}
