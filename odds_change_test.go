package uof

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOddsChange(t *testing.T) {
	buf, err := ioutil.ReadFile("./testdata/odds_change-0.xml")
	assert.Nil(t, err)

	oc := &OddsChange{}
	err = xml.Unmarshal(buf, oc)
	assert.Nil(t, err)

	tests := []struct {
		name string
		f    func(t *testing.T, oc *OddsChange)
	}{
		{"urn", testOddsChangeURN},
		{"specifier", testOddsChangeSpecifiers},
		{"status", testOddsChangeStatus},
	}
	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) { s.f(t, oc) })
	}

	//testu.PP(oc)
}

func testOddsChangeStatus(t *testing.T, oc *OddsChange) {
	m0 := oc.Odds.Markets[0]
	m1 := oc.Odds.Markets[1]
	m2 := oc.Odds.Markets[2]
	m3 := oc.Odds.Markets[3]
	m6 := oc.Odds.Markets[6]

	assert.Equal(t, MarketStatusActive, m0.Status)
	assert.Equal(t, MarketStatusActive, m1.Status)
	assert.Equal(t, MarketStatusInactive, m2.Status)
	assert.Equal(t, MarketStatusSuspended, m3.Status)
	assert.Equal(t, MarketStatusCancelled, m6.Status)
}

func testOddsChangeURN(t *testing.T, oc *OddsChange) {
	assert.Equal(t, 1234, oc.EventID())
	assert.Equal(t, URNTypeMatch, oc.EventURN.Type())
}

func testOddsChangeSpecifiers(t *testing.T, oc *OddsChange) {
	s := oc.Odds.Markets[0].Specifiers
	assert.Equal(t, 1, len(s))
	assert.Equal(t, "41.5", s["score"])

	s = oc.Odds.Markets[3].Specifiers
	assert.Equal(t, 4, len(s))
	assert.Equal(t, "2", s["pero"])
}

func TestSpecifiersParsing(t *testing.T) {
	data := []struct {
		specifiers        string
		extendedSpecifers string
		specifiersMap     map[string]string
		variantSpecifier  string
	}{
		{
			specifiers:    "total=1.5|from=1|to=15",
			specifiersMap: map[string]string{"total": "1.5", "from": "1", "to": "15"},
		},
		{
			specifiers:        "total=1.5|from=1",
			extendedSpecifers: "to=15",
			specifiersMap:     map[string]string{"total": "1.5", "from": "1", "to": "15"},
		},
		{
			extendedSpecifers: "to=15",
			specifiersMap:     map[string]string{"to": "15"},
		},
		{
			specifiers:        "from=1",
			extendedSpecifers: "||",
			specifiersMap:     map[string]string{"from": "1"},
		},

		{
			specifiers:       "total=1.5|variant=sr:exact_goals:4+|from=1|to=15",
			specifiersMap:    map[string]string{"total": "1.5", "from": "1", "to": "15", "variant": "sr:exact_goals:4+"},
			variantSpecifier: "sr:exact_goals:4+",
		},
	}
	for i, d := range data {
		s := toSpecifiers(d.specifiers, d.extendedSpecifers)
		assert.Equal(t, len(d.specifiersMap), len(s))
		m := Market{Specifiers: d.specifiersMap}
		assert.Equal(t, d.variantSpecifier, m.VariantSpecifier())
		for k, v := range d.specifiersMap {
			assert.Equal(t, v, s[k], fmt.Sprintf("failed on %d", i))
		}
	}

}
