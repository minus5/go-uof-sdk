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
	assert.Equal(t, "all", bc.Groups)
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
