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
}

func TestURN(t *testing.T) {
	u := URN("sr:match:123")
	assert.Equal(t, 123, u.ID())
	assert.Equal(t, URNTypeMatch, u.Type())
}

func TestLanguage(t *testing.T) {
	var l Lang
	l.Parse("hr")
	assert.Equal(t, LangHR, l)
	assert.Equal(t, "hr", l.Code())
	assert.Equal(t, "Croatian", l.Name())

	ls := Languages("hr,en,de")
	assert.Len(t, ls, 3)
	assert.Equal(t, LangHR, ls[0])
	assert.Equal(t, LangEN, ls[1])
	assert.Equal(t, LangDE, ls[2])
}

func TestMessageTypes(t *testing.T) {
	for i, n := range messageTypeNames {
		m := messageTypes[i]
		assert.Equal(t, m.String(), n)

		var m2 MessageType
		m2.Parse(n)
		assert.Equal(t, m, m2)
	}
}
