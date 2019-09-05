package uof

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProducer(t *testing.T) {
	assert.Equal(t, "Ctrl", Producer(3).String())
	assert.Equal(t, "Ctrl", Producer(3).Name())
	assert.Equal(t, "Betradar Ctrl", Producer(3).Description())
	assert.Equal(t, InvalidName, Producer(-1).String())
}

func TestURN(t *testing.T) {
	u := URN("sr:match:123")
	assert.Equal(t, 123, u.ID())
	assert.Equal(t, URNTypeMatch, u.Type())
}
