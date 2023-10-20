package pipe

import (
	"io/ioutil"
	"sync"
	"testing"

	"github.com/pvotal-tech/go-uof-sdk"
	"github.com/stretchr/testify/assert"
)

type playerAPIMock struct {
	requests map[int]struct{}
	sync.Mutex
}

func (m *playerAPIMock) Player(lang uof.Lang, playerID int) (*uof.Player, error) {
	m.Lock()
	defer m.Unlock()
	m.requests[uof.UIDWithLang(playerID, lang)] = struct{}{}
	return nil, nil
}

func TestPlayerPipe(t *testing.T) {
	a := &playerAPIMock{requests: make(map[int]struct{})}
	p := Player(a, []uof.Lang{uof.LangEN, uof.LangDE})
	assert.NotNil(t, p)

	in := make(chan *uof.Message)
	out, _ := p(in)

	// this type of message is passing through
	m := uof.NewSimpleConnnectionMessage(uof.ConnectionStatusUp)
	in <- m
	om := <-out
	assert.Equal(t, m, om)

	m = oddsChangeMessage(t)
	in <- m
	om = <-out
	assert.Equal(t, m, om)

	close(in)
	cnt := 0
	for range out {
		cnt++
	}
	assert.Equal(t, 82, cnt)
	assert.Equal(t, 82, len(a.requests))
}

func oddsChangeMessage(t *testing.T) *uof.Message {
	buf, err := ioutil.ReadFile("../testdata/odds_change-0.xml")
	assert.NoError(t, err)
	m, err := uof.NewQueueMessage("hi.pre.-.odds_change.1.sr:match.1234.-", buf)
	assert.NoError(t, err)
	return m
}
