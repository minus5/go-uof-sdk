package pipe

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExpireMap(t *testing.T) {
	em := newExpireMap(time.Minute)
	em.insert(1)
	assert.True(t, em.fresh(1))
}
