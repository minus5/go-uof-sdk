package pipe

import (
	"testing"

	"github.com/minus5/go-uof-sdk"
	"github.com/stretchr/testify/assert"
)

func TestBetStop(t *testing.T) {
	b := betStop{
		marketGroups: marketGroups(),
	}
	m := &uof.Message{
		Header: uof.Header{
			Type: uof.MessageTypeBetStop,
		},
		Body: uof.Body{
			BetStop: &uof.BetStop{
				Groups: []string{"regular_play"},
			},
		},
	}
	assert.Nil(t, m.BetStop.MarketIDs)
	b.enrich(m)
	assert.NotNil(t, m.BetStop.MarketIDs)
	assert.Len(t, m.BetStop.MarketIDs, 217)
	m.BetStop.Groups = append(m.BetStop.Groups, "penalties")
	b.enrich(m)
	assert.Len(t, m.BetStop.MarketIDs, 218)

	m.BetStop.Groups = []string{"corners", "15_min"}
	b.enrich(m)
	assert.Len(t, m.BetStop.MarketIDs, 54)
}

func TestDedup(t *testing.T) {
	a := []int{4, 2, 3, 3, 1, 2, 5}
	b := dedup(a)
	assert.Equal(t, []int{1, 2, 3, 4, 5}, b)
}
