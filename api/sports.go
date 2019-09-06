package api

import (
	"github.com/minus5/uof"
)

const (
	pathMarkets       = "/v1/descriptions/{{.Lang}}/markets.xml?include_mappings={{.IncludeMappings}}"
	pathMarketVariant = "/v1/descriptions/{{.Lang}}/markets/{{.MarketID}}/variants/{{.Variant}}?include_mappings={{.IncludeMappings}}"
	pathFixture       = "/v1/sports/{{.Lang}}/sport_events/{{.EventURN}}/fixture.xml"
	pathPlayer        = "/v1/sports/{{.Lang}}/players/sr:player:{{.PlayerID}}/profile.xml"
)

// Markets all currently available markets for a language
func (a *Api) Markets(lang uof.Lang) ([]byte, error) {
	return a.get(pathMarkets, &params{Lang: lang})
}

func (a *Api) MarketVariant(lang uof.Lang, marketID int, variant string) ([]byte, error) {
	return a.get(pathMarketVariant, &params{Lang: lang, MarketID: marketID, Variant: variant})
}

// Fixture lists the fixture for a specified sport event
func (a *Api) Fixture(lang uof.Lang, eventURN uof.URN) ([]byte, error) {
	return a.get(pathFixture, &params{Lang: lang, EventURN: eventURN})
}

func (a *Api) Player(lang uof.Lang, playerID int) ([]byte, error) {
	return a.get(pathPlayer, &params{Lang: lang, PlayerID: playerID})
}
