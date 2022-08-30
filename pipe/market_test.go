package pipe

import (
	"fmt"
	"sync"
	"testing"

	"github.com/minus5/go-uof-sdk"
	"github.com/stretchr/testify/assert"
)

type marketsAPIMock struct {
	requests map[string]struct{}
	sync.Mutex
}

func (m *marketsAPIMock) Markets(lang uof.Lang) (uof.MarketDescriptions, error) {
	return nil, nil
}

func (m *marketsAPIMock) MarketVariant(lang uof.Lang, marketID int, variant string) (uof.MarketDescriptions, error) {
	m.Lock()
	defer m.Unlock()
	m.requests[fmt.Sprintf("%s %d %s", lang, marketID, variant)] = struct{}{}
	return nil, nil
}

func TestMarketsPipe(t *testing.T) {
	a := &marketsAPIMock{requests: make(map[string]struct{})}
	ms := Markets(a, []uof.Lang{uof.LangEN, uof.LangDE})
	assert.NotNil(t, ms)

	in := make(chan *uof.Message)
	out, _ := ms(in)

	// this type of message is passing through
	m := uof.NewSimpleConnnectionMessage(uof.ConnectionStatusUp)
	in <- m
	om := <-out
	assert.Equal(t, m, om)

	m = oddsChangeMessage(t)
	in <- m
	// om = <-out
	// assert.Equal(t, m, om)

	close(in)
	cnt := 0
	for range out {
		cnt++
	}
	assert.Equal(t, 5, cnt)

	_, found := a.requests["en 145 sr:point_range:76+"]
	assert.True(t, found)

	_, found = a.requests["de 145 sr:point_range:76+"]
	assert.True(t, found)

}
