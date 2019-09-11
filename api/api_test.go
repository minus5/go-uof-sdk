package api

import (
	"os"
	"testing"

	"github.com/minus5/uof"
	"github.com/stretchr/testify/assert"
)

func TestTemplate(t *testing.T) {
	path := runTemplate(startScenario, &params{ScenarioID: 1, Speed: 2, MaxDelay: 3})
	assert.Equal(t, "/v1/replay/scenario/play/1?speed=2&max_delay=3&use_replay_timestamp=false", path)
}

const EnvToken = "UOF_TOKEN"

// this test depends on UOF_TOKEN enviroment variable
// to be set to the staging access token
// run it as:
//    UOF_TOKEN=my-token go test -v
func TestIntegration(t *testing.T) {
	token, ok := os.LookupEnv(EnvToken)
	if !ok {
		t.Skip("integration token not found")
	}

	a, err := Staging(token)
	assert.NoError(t, err)

	tests := []struct {
		name string
		f    func(t *testing.T, a *Api)
	}{
		{"markets", testMarkets},
		{"marketVariant", testMarketVariant},
		{"fixture", testFixture},
		{"player", testPlayer},
	}
	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) { s.f(t, a) })
	}
}

func testMarkets(t *testing.T, a *Api) {
	lang := uof.LangEN
	buf, err := a.Markets(lang)
	assert.Nil(t, err)
	mm, err := uof.NewMarketsMessage(lang, buf)
	assert.Nil(t, err)

	assert.True(t, len(mm.Markets) >= 992)
	m := mm.Markets.Find(1)
	assert.Equal(t, "1x2", m.Name)
	//testu.PP(m)
	//testu.PP(m)
}

func testMarketVariant(t *testing.T, a *Api) {
	lang := uof.LangEN
	buf, err := a.MarketVariant(lang, 241, "sr:exact_games:bestof:5")
	assert.Nil(t, err)
	mm, err := uof.NewMarketsMessage(lang, buf)
	assert.Nil(t, err)
	assert.Len(t, mm.Markets, 1)
	m := mm.Markets[0]
	assert.Equal(t, "Exact games", m.Name)
	assert.Len(t, m.Outcomes, 3)
	//testu.PP(mm)
}

func testFixture(t *testing.T, a *Api) {
	lang := uof.LangEN
	buf, err := a.Fixture(lang, "sr:match:8696826")
	assert.Nil(t, err)

	fc := uof.Message{Header: uof.Header{Type: uof.MessageTypeFixtureChange}}
	fm, err := fc.AsFixture(lang, buf)
	assert.Nil(t, err)

	assert.Equal(t, "IK Oddevold", fm.Fixture.Home.Name)

	//testu.PP(fm)
}

func testPlayer(t *testing.T, a *Api) {
	lang := uof.LangEN
	buf, err := a.Player(lang, 947)
	assert.Nil(t, err)

	pm, err := uof.NewPlayerMessage(lang, buf)
	assert.Nil(t, err)

	assert.Equal(t, "Lee Barnard", pm.Player.FullName)
	//testu.PP(pm.Player)
}
