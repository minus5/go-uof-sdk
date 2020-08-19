package api

import (
	"encoding/xml"
	"time"

	"github.com/minus5/go-uof-sdk"
)

const (
	pathMarkets       = "/v1/descriptions/{{.Lang}}/markets.xml?include_mappings={{.IncludeMappings}}"
	pathMarketVariant = "/v1/descriptions/{{.Lang}}/markets/{{.MarketID}}/variants/{{.Variant}}?include_mappings={{.IncludeMappings}}"
	pathFixture       = "/v1/sports/{{.Lang}}/sport_events/{{.EventURN}}/fixture.xml"
	pathSummary       = "/v1/sports/{{.Lang}}/sport_events/{{.EventURN}}/summary.xml"
	pathPlayer        = "/v1/sports/{{.Lang}}/players/sr:player:{{.PlayerID}}/profile.xml"
	pathCompetitor    = "/v1/sports/{{.Lang}}/competitors/sr:competitor:{{.PlayerID}}/profile.xml"
	events            = "/v1/sports/{{.Lang}}/schedules/pre/schedule.xml?start={{.Start}}&limit={{.Limit}}"
	liveEvents        = "/v1/sports/{{.Lang}}/schedules/live/schedule.xml"
)

// Markets all currently available markets for a language
func (a *API) Markets(lang uof.Lang) (uof.MarketDescriptions, error) {
	var mr marketsRsp
	return mr.Markets, a.getAs(&mr, pathMarkets, &params{Lang: lang})
}

func (a *API) MarketVariant(lang uof.Lang, marketID int, variant string) (uof.MarketDescriptions, error) {
	var mr marketsRsp
	return mr.Markets, a.getAs(&mr, pathMarketVariant, &params{Lang: lang, MarketID: marketID, Variant: variant})
}

// Fixture lists the fixture for a specified sport event
func (a *API) Fixture(lang uof.Lang, eventURN uof.URN) (*uof.Fixture, error) {
	var fr fixtureRsp
	return &fr.Fixture, a.getAs(&fr, pathFixture, &params{Lang: lang, EventURN: eventURN})
}

// Fixture with extra information
func (a *API) Summary(lang uof.Lang, eventURN uof.URN) (*uof.Summary, error) {
	var s uof.Summary
	return &s, a.getAs(&s, pathSummary, &params{Lang: lang, EventURN: eventURN})
}

func (a *API) Tournament(lang uof.Lang, eventURN uof.URN) (*uof.FixtureTournament, error) {
	var ft uof.FixtureTournament
	return &ft, a.getAs(&ft, pathFixture, &params{Lang: lang, EventURN: eventURN})
}

func (a *API) Player(lang uof.Lang, playerID int) (*uof.Player, error) {
	var pr playerRsp
	return &pr.Player, a.getAs(&pr, pathPlayer, &params{Lang: lang, PlayerID: playerID})
}

func (a *API) Competitor(lang uof.Lang, playerID int) (*uof.CompetitorPlayer, error) {
	var cr competitorRsp
	return &cr.Competitor, a.getAs(&cr, pathCompetitor, &params{Lang: lang, PlayerID: playerID})
}

type marketsRsp struct {
	Markets uof.MarketDescriptions `xml:"market,omitempty" json:"markets,omitempty"`
	// unused
	// ResponseCode string   `xml:"response_code,attr,omitempty" json:"responseCode,omitempty"`
	// Location     string   `xml:"location,attr,omitempty" json:"location,omitempty"`
}

type playerRsp struct {
	Player      uof.Player `xml:"player" json:"player"`
	GeneratedAt time.Time  `xml:"generated_at,attr,omitempty" json:"generatedAt,omitempty"`
}

type competitorRsp struct {
	Competitor  uof.CompetitorPlayer `xml:"competitor" json:"competitor"`
	GeneratedAt time.Time            `xml:"generated_at,attr,omitempty" json:"generatedAt,omitempty"`
}

type fixtureRsp struct {
	Fixture     uof.Fixture `xml:"fixture" json:"fixture"`
	GeneratedAt time.Time   `xml:"generated_at,attr,omitempty" json:"generatedAt,omitempty"`
}

type scheduleRsp struct {
	Fixtures    []uof.Fixture `xml:"sport_event,omitempty" json:"sportEvent,omitempty"`
	GeneratedAt time.Time     `xml:"generated_at,attr,omitempty" json:"generatedAt,omitempty"`
}

// Fixtures gets all the fixtures with schedule before to
func (a *API) Fixtures(lang uof.Lang, to time.Time) (<-chan uof.Fixture, <-chan error) {
	errc := make(chan error, 1)
	out := make(chan uof.Fixture)
	go func() {
		defer close(out)
		defer close(errc)
		done := false

		parse := func(buf []byte) error {
			var sr scheduleRsp
			if err := xml.Unmarshal(buf, &sr); err != nil {
				return uof.Notice("unmarshal", err)
			}
			for _, f := range sr.Fixtures {
				out <- f
				if f.Scheduled.After(to) {
					done = true
				}
			}
			return nil
		}

		// first live events
		buf, err := a.get(liveEvents, &params{Lang: lang})
		if err != nil {
			errc <- err
			return
		}
		if err := parse(buf); err != nil {
			errc <- err
			return
		}

		// than all events which has scheduled before to
		limit := 1000
		for start := 0; true; start += limit {
			buf, err := a.get(events, &params{Lang: lang, Start: start, Limit: limit})
			if err != nil {
				errc <- err
				return
			}
			if err := parse(buf); err != nil {
				errc <- err
				return
			}
			if done {
				return
			}
		}
	}()

	return out, errc
}
