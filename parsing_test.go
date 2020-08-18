package uof

import (
	"encoding/xml"
	"io/ioutil"
	"testing"
	"time"

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

	// use message entry point
	m, err := NewQueueMessage("hi.pre.-.bet_cancel.1.sr:match.1234.-", buf)
	assert.NoError(t, err)
	assert.True(t, m.Is(MessageTypeBetCancel))
	assert.NotNil(t, m.BetCancel)
	assert.Equal(t, bc, m.BetCancel)
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

	m, err := NewQueueMessage("hi.pre.-.rollback_bet_cancel.1.sr:match.1234.-", buf)
	assert.NoError(t, err)
	assert.True(t, m.Is(MessageTypeRollbackBetCancel))
	assert.NotNil(t, m.RollbackBetCancel)
	assert.Equal(t, bc, m.RollbackBetCancel)
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

	assert.Equal(t, OutcomeResultWinWithDeadHead, bs.Markets[3].Outcomes[0].Result)
	assert.Equal(t, OutcomeResultWinWithDeadHead, bs.Markets[3].Outcomes[1].Result)
	assert.Equal(t, OutcomeResultLose, bs.Markets[3].Outcomes[2].Result)

	assert.Equal(t, 0.5, bs.Markets[3].Outcomes[0].DeadHeatFactor)
	assert.Equal(t, 0.5, bs.Markets[3].Outcomes[1].DeadHeatFactor)
	assert.Equal(t, 0.0, bs.Markets[3].Outcomes[2].DeadHeatFactor)

	m4 := bs.Markets[4]
	assert.Equal(t, 10077, m4.Outcomes[0].ID)
	assert.Equal(t, 10077, m4.Outcomes[0].PlayerID)
	assert.Equal(t, 38560, m4.Outcomes[1].ID)
	assert.Equal(t, 38560, m4.Outcomes[1].PlayerID)
	for _, o := range m4.Outcomes {
		assert.Equal(t, o.PlayerID, o.ID)
	}

	m, err := NewQueueMessage("hi.pre.-.bet_settlement.1.sr:match.1234.-", buf)
	assert.NoError(t, err)
	assert.True(t, m.Is(MessageTypeBetSettlement))
	assert.NotNil(t, m.BetSettlement)
	assert.Equal(t, bs, m.BetSettlement)

}

func TestRollbackBetSettlement(t *testing.T) {
	buf := []byte(`<rollback_bet_settlement event_id="sr:match:299321" timestamp="1236" product="1">
	<market id="47"/>
 </rollback_bet_settlement>`)
	rbs := &RollbackBetSettlement{}
	err := xml.Unmarshal(buf, rbs)
	assert.Nil(t, err)
	assert.Len(t, rbs.Markets, 1)
	assert.Equal(t, 47, rbs.Markets[0].ID)

	m, err := NewQueueMessage("hi.pre.-.rollback_bet_settlement.1.sr:match.1234.-", buf)
	assert.NoError(t, err)
	assert.True(t, m.Is(MessageTypeRollbackBetSettlement))
	assert.NotNil(t, m.RollbackBetSettlement)
	assert.Equal(t, rbs, m.RollbackBetSettlement)
}

func TestBetStop(t *testing.T) {
	buf := []byte(`<bet_stop timestamp="12345" product="3" event_id="sr:match:471123" groups="all"/>`)

	bs := &BetStop{}
	err := xml.Unmarshal(buf, bs)
	assert.Nil(t, err)

	assert.Equal(t, 471123, bs.EventID)
	assert.Equal(t, MarketStatusSuspended, bs.Status)
	assert.Equal(t, []string(nil), bs.Groups)
	assert.Len(t, bs.Groups, 0)

	buf = []byte(`<bet_stop timestamp="12345" product="3" event_id="sr:match:471123" groups="10_min|180s"/>`)
	bs = &BetStop{}
	err = xml.Unmarshal(buf, bs)
	assert.Nil(t, err)
	assert.Len(t, bs.Groups, 2)

	m, err := NewQueueMessage("hi.pre.-.bet_stop.1.sr:match.1234.-", buf)
	assert.NoError(t, err)
	assert.True(t, m.Is(MessageTypeBetStop))
	assert.NotNil(t, m.BetStop)
	assert.Equal(t, bs, m.BetStop)
}

func TestFixtureChange(t *testing.T) {
	buf := []byte(`<fixture_change event_id="sr:match:1234" product="3" start_time="1511107200000"/>`)
	fc := &FixtureChange{}
	err := xml.Unmarshal(buf, fc)
	assert.Nil(t, err)
	assert.Equal(t, 1234, fc.EventID)
	assert.Nil(t, fc.ChangeType)
	assert.Equal(t, "2017-11-19T16:00:00Z", fc.Schedule().UTC().Format(time.RFC3339))

	fc2 := &FixtureChange{}
	buf = []byte(`<fixture_change event_id="sr:match:1234" change_type="5" product="3"/>`)
	err = xml.Unmarshal(buf, fc2)
	assert.Nil(t, err)
	assert.Equal(t, FixtureChangeTypeCoverage, *fc2.ChangeType)

	m, err := NewQueueMessage("hi.pre.-.fixture_change.1.sr:match.1234.-", buf)
	assert.NoError(t, err)
	assert.True(t, m.Is(MessageTypeFixtureChange))
	assert.NotNil(t, m.FixtureChange)
	assert.Equal(t, fc2, m.FixtureChange)
}

func TestSnaphotComplete(t *testing.T) {
	buf := []byte(`<snapshot_complete request_id="1234" timestamp="1234578" product="3"/>`)
	sc := &SnapshotComplete{}
	err := xml.Unmarshal(buf, sc)
	assert.Nil(t, err)
	assert.Equal(t, Producer(3), sc.Producer)

	m, err := NewQueueMessage("-.-.-.snapshot_complete.-.-.-", buf)
	assert.NoError(t, err)
	assert.True(t, m.Is(MessageTypeSnapshotComplete))
	assert.NotNil(t, m.SnapshotComplete)
	assert.Equal(t, sc, m.SnapshotComplete)
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

	assert.Equal(t, &ms.Markets[4], ms.Markets.Find(575))
	assert.Len(t, ms.Markets.Groups(), 5)

	//test nil return
	assert.Nil(t, nil, ms.Markets.Find(1111))

	msg, err := NewAPIMessage(LangEN, MessageTypeMarkets, buf)
	assert.NoError(t, err)
	assert.Equal(t, ms.Markets, msg.Markets)
}

func TestPlayerMale(t *testing.T) {
	buf, err := ioutil.ReadFile("./testdata/player_profile_m.xml")
	assert.Nil(t, err)

	pp := &PlayerProfile{}
	err = xml.Unmarshal(buf, pp)
	assert.Nil(t, err)

	p := pp.Player
	assert.Equal(t, 947, p.ID)
	assert.Equal(t, Male, p.Gender)
	assert.Equal(t, "forward", p.Type)
	assert.Equal(t, "1984-07-18", p.DateOfBirth.Format(apiDateFormat))

	msg, err := NewAPIMessage(LangEN, MessageTypePlayer, buf)
	assert.NoError(t, err)
	assert.Equal(t, p, *msg.Player)
}

func TestPlayerFemale(t *testing.T) {
	buf, err := ioutil.ReadFile("./testdata/player_profile_f.xml")
	assert.Nil(t, err)

	pp := &PlayerProfile{}
	err = xml.Unmarshal(buf, pp)
	assert.Nil(t, err)

	p := pp.Player
	assert.Equal(t, 948, p.ID)
	assert.Equal(t, Female, p.Gender)
	assert.Equal(t, "forward", p.Type)
	assert.Equal(t, "1989-09-19", p.DateOfBirth.Format(apiDateFormat))

	msg, err := NewAPIMessage(LangEN, MessageTypePlayer, buf)
	assert.NoError(t, err)
	assert.Equal(t, p, *msg.Player)
}

func TestPlayerUnknown(t *testing.T) {
	buf, err := ioutil.ReadFile("./testdata/player_profile_u.xml")
	assert.Nil(t, err)

	pp := &PlayerProfile{}
	err = xml.Unmarshal(buf, pp)
	assert.Nil(t, err)

	p := pp.Player
	assert.Equal(t, 949, p.ID)
	assert.Equal(t, GenderUnknown, p.Gender)
	assert.Equal(t, "forward", p.Type)
	assert.Equal(t, "1985-01-01", p.DateOfBirth.Format(apiDateFormat))

	msg, err := NewAPIMessage(LangEN, MessageTypePlayer, buf)
	assert.NoError(t, err)
	assert.Equal(t, p, *msg.Player)
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
	assert.Equal(t, "Ajax Amsterdam - Tottenham Hotspur                                                         08.05. 19:00          closed", f.PP())

	assert.Equal(t, 2953, f.Home.ID)
	assert.Equal(t, 33, f.Away.ID)

	// <extra_info>
	// 	<info key="neutral_ground" value="false"/>
	//  ...
	// </extra_info>
	assert.Equal(t, "neutral_ground", f.ExtraInfo[0].Key)
	assert.Equal(t, "false", f.ExtraInfo[0].Value)
	assert.Equal(t, "period_length", f.ExtraInfo[1].Key)
	assert.Equal(t, "45", f.ExtraInfo[1].Value)
	assert.Equal(t, "overtime_length", f.ExtraInfo[2].Key)
	assert.Equal(t, "15", f.ExtraInfo[2].Value)
	assert.Equal(t, "coverage_source", f.ExtraInfo[3].Key)
	assert.Equal(t, "tv", f.ExtraInfo[3].Value)
	assert.Equal(t, "extended_live_markets_offered", f.ExtraInfo[4].Key)
	assert.Equal(t, "true", f.ExtraInfo[4].Value)

	// test creating uof.Message
	msgAPI, err := NewAPIMessage(LangEN, MessageTypeFixture, buf)
	assert.NoError(t, err)
	assert.Equal(t, f, *msgAPI.Fixture)
}

func TestFixutreWithPlayers(t *testing.T) {
	buf, err := ioutil.ReadFile("./testdata/fixture-2.xml")
	assert.Nil(t, err)

	msg, err := NewAPIMessage(LangEN, MessageTypeFixture, buf)
	assert.NoError(t, err)
	assert.Len(t, msg.Fixture.Competitors[0].Players, 2)
	assert.Len(t, msg.Fixture.Competitors[1].Players, 2)
	assert.Equal(t, "Goldhoff, George", msg.Fixture.Competitors[1].Players[0].Name)
}

func TestFixtureTournament(t *testing.T) {
	buf, err := ioutil.ReadFile("./testdata/fixture-3.xml")
	assert.Nil(t, err)

	ft := FixtureTournament{}
	err = xml.Unmarshal(buf, &ft)
	assert.Nil(t, err)

	assert.Equal(t, 13933, ft.ID)
	assert.Equal(t, "vf:tournament:13933", string(ft.URN))
	assert.Equal(t, 1111, ft.Category.ID)
	assert.Equal(t, 13933, ft.Tournament.ID)
	assert.Len(t, ft.Groups, 6)
	assert.Len(t, ft.Groups[0].Competitors, 4)
	assert.Equal(t, "Jamaica", ft.Groups[0].Competitors[2].Name)
	pp(ft)
}

func TestBetSettlementToResult(t *testing.T) {
	data := []struct {
		result         int
		voidFactor     float64
		deadHeatFactor float64
		outcomeResult  OutcomeResult
	}{
		{0, 0, 0, OutcomeResultLose},
		{1, 0, 0, OutcomeResultWin},
		{0, 1, 0, OutcomeResultVoid},
		{1, 0.5, 0, OutcomeResultHalfWin},
		{0, 0.5, 0, OutcomeResultHalfLose},
		{1, 0, 1, OutcomeResultWinWithDeadHead},
		// wrong inputs
		{-1, 0, 0, OutcomeResultUnknown},
		{2, 0, 0, OutcomeResultUnknown},
	}

	for _, d := range data {
		r := toResult(&d.result, &d.voidFactor, &d.deadHeatFactor)
		assert.Equal(t, d.outcomeResult, r)
	}

	r := toResult(nil, nil, nil)
	assert.Equal(t, OutcomeResultUnknown, r)
}

func TestBetStopToMarketStatus(t *testing.T) {
	data := []struct {
		in  int
		out MarketStatus
	}{
		{0, MarketStatusInactive},
		{1, MarketStatusActive},
		{-1, MarketStatusSuspended},
		{-2, MarketStatusHandedOver},
		{-3, MarketStatusSettled},
		{-4, MarketStatusCancelled},
		// wrong inputs
		{-5, MarketStatusInactive},
		{2, MarketStatusInactive},
	}

	for _, d := range data {
		assert.Equal(t, d.out, toMarketStatus(&d.in))
	}
	assert.Equal(t, MarketStatusSuspended, toMarketStatus(nil))
}

func TestToOutcomeType(t *testing.T) {
	data := []struct {
		in  string
		out OutcomeType
	}{
		{"", OutcomeTypeDefault},
		{"player", OutcomeTypePlayer},
		{"competitor", OutcomeTypeCompetitor},
		{"competitors", OutcomeTypeCompetitors},
		{"free_text", OutcomeTypeFreeText},
		// wrong inputs
		{"pero", OutcomeTypeUnknown},
	}

	for _, d := range data {
		assert.Equal(t, d.out, toOutcomeType(d.in))
	}
}

func TestToSpecifierType(t *testing.T) {
	data := []struct {
		in  string
		out SpecifierType
	}{
		{"string", SpecifierTypeString},
		{"integer", SpecifierTypeInteger},
		{"decimal", SpecifierTypeDecimal},
		{"variable_text", SpecifierTypeVariableText},
		// wrong inputs
		{"pero", SpecifierTypeUnknown},
	}

	for _, d := range data {
		assert.Equal(t, d.out, toSpecifierType(d.in))
	}
}

func TestConnectionStatus(t *testing.T) {
	data := []struct {
		in  string
		out ConnectionStatus
	}{
		{"down", ConnectionStatusDown},
		{"up", ConnectionStatusUp},
		// default
		{"?", -1},
	}

	for _, d := range data {
		assert.Equal(t, d.in, d.out.String())
	}

}
