package uof

import (
	"encoding/xml"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBetCancel(t *testing.T) {
	buf, err := ioutil.ReadFile("./testdata/bet_cancel.xml")
	assert.Nil(t, err)

	bc := &BetCancel{}
	err = xml.Unmarshal(buf, bc)
	assert.Nil(t, err)

	assert.Equal(t, 18941600, bc.EventID)
	assert.Equal(t, 62, bc.Markets[0].ID)
	assert.Equal(t, 2296512168, bc.Markets[0].LineID)
}

func TestRollbackBetCancel(t *testing.T) {
	buf, err := ioutil.ReadFile("./testdata/rollback_bet_cancel.xml")
	assert.Nil(t, err)

	bc := &RollbackBetCancel{}
	err = xml.Unmarshal(buf, bc)
	assert.Nil(t, err)

	assert.Equal(t, 4444, bc.EventID)
	assert.Equal(t, 48, bc.Markets[0].ID)
	assert.Equal(t, 2701050930, bc.Markets[0].LineID)
}

func TestBetSettlement(t *testing.T) {
	buf, err := ioutil.ReadFile("./testdata/bet_settlement.xml")
	assert.Nil(t, err)

	bs := &BetSettlement{}
	err = xml.Unmarshal(buf, bs)
	assert.Nil(t, err)

	assert.Equal(t, 16807109, bs.EventID)
	assert.Equal(t, 193, bs.Markets[0].ID)
	assert.Equal(t, 0, bs.Markets[0].LineID)

	assert.Equal(t, 204, bs.Markets[1].ID)
	assert.Equal(t, 1683548904, bs.Markets[1].LineID)

	assert.Equal(t, OutcomeResultLose, bs.Markets[0].Outcomes[0].Result)
	assert.Equal(t, OutcomeResultWin, bs.Markets[0].Outcomes[1].Result)

	assert.Equal(t, OutcomeResultVoid, bs.Markets[2].Outcomes[0].Result)
	assert.Equal(t, OutcomeResultHalfLose, bs.Markets[2].Outcomes[1].Result)
	assert.Equal(t, OutcomeResultHalfWin, bs.Markets[2].Outcomes[2].Result)
}

func TestBetStop(t *testing.T) {
	buf := []byte(`<bet_stop timestamp="12345" product="3" event_id="sr:match:471123" groups="all"/>`)

	bc := &BetStop{}
	err := xml.Unmarshal(buf, bc)
	assert.Nil(t, err)

	assert.Equal(t, 471123, bc.EventID)
	assert.Equal(t, MarketStatusSuspended, bc.Status)
	assert.Equal(t, []string(nil), bc.Groups)
	assert.Len(t, bc.Groups, 0)

	buf = []byte(`<bet_stop timestamp="12345" product="3" event_id="sr:match:471123" groups="10_min|180s"/>`)
	bc = &BetStop{}
	err = xml.Unmarshal(buf, bc)
	assert.Nil(t, err)
	assert.Len(t, bc.Groups, 2)
}

func TestFixtureChange(t *testing.T) {
	buf := []byte(`<fixture_change event_id="sr:match:1234" product="3"/>`)
	fc := &FixtureChange{}
	err := xml.Unmarshal(buf, fc)
	assert.Nil(t, err)
	assert.Equal(t, 1234, fc.EventID)
	assert.Nil(t, fc.ChangeType)

	buf = []byte(`<fixture_change event_id="sr:match:1234" change_type="5" product="3"/>`)
	err = xml.Unmarshal(buf, fc)
	assert.Nil(t, err)
	assert.Equal(t, FixtureChangeTypeCoverage, *fc.ChangeType)
}

func TestMarkets(t *testing.T) {
	buf, err := ioutil.ReadFile("./testdata/markets-0.xml")
	assert.Nil(t, err)

	ms := &MarketsRsp{}
	err = xml.Unmarshal(buf, ms)
	assert.Nil(t, err)

	assert.Len(t, ms.Markets, 7)
	m := ms.Markets[0]
	assert.Equal(t, 1, m.ID)
	assert.Equal(t, 0, m.VariantID)
	assert.Len(t, m.Groups, 2)
	assert.Len(t, m.Outcomes, 3)
	assert.Equal(t, OutcomeTypeDefault, m.OutcomeType)

	m = ms.Markets[3]
	assert.Equal(t, 21, m.ID)
	assert.Equal(t, 1686878731, m.VariantID)
	assert.Len(t, m.Groups, 0)
	assert.Len(t, m.Outcomes, 7)
	assert.Equal(t, 1644387477, m.Outcomes[0].ID)
	assert.Equal(t, 1627609858, m.Outcomes[1].ID)

	m = ms.Markets[4]
	assert.Equal(t, 575, m.ID)
	assert.Len(t, m.Groups, 2)
	assert.Len(t, m.Outcomes, 2)
	assert.Len(t, m.Specifiers, 3)
	assert.Equal(t, SpecifierTypeDecimal, m.Specifiers[0].Type)
	assert.Equal(t, SpecifierTypeInteger, m.Specifiers[1].Type)

	m = ms.Markets[5]
	assert.Equal(t, 892, m.ID)
	assert.Equal(t, 0, m.VariantID)
	assert.Len(t, m.Groups, 2)
	assert.Len(t, m.Outcomes, 0)
	assert.Equal(t, OutcomeTypePlayer, m.OutcomeType)
	assert.Equal(t, SpecifierTypeVariableText, m.Specifiers[0].Type)
	assert.Equal(t, SpecifierTypeInteger, m.Specifiers[1].Type)
	assert.Equal(t, SpecifierTypeString, m.Specifiers[2].Type)

	m = ms.Markets[6]
	assert.Equal(t, 892, m.ID)
	assert.Equal(t, 3487053313, m.VariantID)
	assert.Len(t, m.Groups, 0)
	assert.Len(t, m.Outcomes, 3)
	//testu.PP(ms)
}

func TestPlayer(t *testing.T) {
	buf, err := ioutil.ReadFile("./testdata/player_profile.xml")
	assert.Nil(t, err)

	pp := &PlayerProfile{}
	err = xml.Unmarshal(buf, pp)
	assert.Nil(t, err)

	p := pp.Player
	assert.Equal(t, 947, p.ID)
	assert.Equal(t, Male, p.Gender)
	assert.Equal(t, "forward", p.Type)
	assert.Equal(t, "1984-07-18", p.DateOfBirth.Format(apiDateFormat))
	//testu.PP(p)
}

func TestFixture(t *testing.T) {
	buf, err := ioutil.ReadFile("./testdata/fixture-0.xml")
	assert.Nil(t, err)

	fr := &FixtureRsp{}
	err = xml.Unmarshal(buf, fr)
	assert.Nil(t, err)

	f := fr.Fixture
	assert.Equal(t, 18001015, f.ID)
	assert.Equal(t, "2019-05-08 19:00", f.StartTime.Format("2006-01-02 15:04"))
	assert.Len(t, f.Competitors, 2)
	assert.Len(t, f.TvChannels, 30)

	assert.Equal(t, "Soccer", f.Sport.Name)
	assert.Equal(t, 1, f.Sport.ID)
	assert.Equal(t, "International Clubs", f.Category.Name)
	assert.Equal(t, 393, f.Category.ID)
	assert.Equal(t, "UEFA Champions League", f.Tournament.Name)
	assert.Equal(t, 7, f.Tournament.ID)

	assert.Equal(t, 2953, f.Home.ID)
	assert.Equal(t, 33, f.Away.ID)
	//testu.PP(f)
}
