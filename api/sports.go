package api

import (
	"encoding/xml"
	"time"

	"github.com/minus5/uof"
)

const (
	pathMarkets       = "/v1/descriptions/{{.Lang}}/markets.xml?include_mappings={{.IncludeMappings}}"
	pathMarketVariant = "/v1/descriptions/{{.Lang}}/markets/{{.MarketID}}/variants/{{.Variant}}?include_mappings={{.IncludeMappings}}"
	pathFixture       = "/v1/sports/{{.Lang}}/sport_events/{{.EventURN}}/fixture.xml"
	pathPlayer        = "/v1/sports/{{.Lang}}/players/sr:player:{{.PlayerID}}/profile.xml"
	events            = "/v1/sports/{{.Lang}}/schedules/pre/schedule.xml?start={{.Start}}&limit={{.Limit}}"
	liveEvents        = "/v1/sports/{{.Lang}}/schedules/live/schedule.xml"
)

// Markets all currently available markets for a language
func (a *Api) Markets(lang uof.Lang) ([]byte, error) {
	return a.get(pathMarkets, &params{Lang: lang})
}

func (a *Api) MarketVariant(lang uof.Lang, marketID int, variant string) ([]byte, error) {
	return a.get(pathMarketVariant, &params{Lang: lang, MarketID: marketID, Variant: variant})
}

// Fixture lists the fixture for a specified sport event
func (a *Api) Fixture(lang uof.Lang, eventURN uof.URN) (*uof.Fixture, error) {
	buf, err := a.get(pathFixture, &params{Lang: lang, EventURN: eventURN})
	if err != nil {
		return nil, err
	}
	var fr fixtureRsp
	if err := xml.Unmarshal(buf, &fr); err != nil {
		return nil, err
	}
	return &fr.Fixture, nil
}

func (a *Api) Player(lang uof.Lang, playerID int) ([]byte, error) {
	return a.get(pathPlayer, &params{Lang: lang, PlayerID: playerID})
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
func (a *Api) Fixtures(lang uof.Lang, to time.Time) (<-chan uof.Fixture, <-chan error) {
	errc := make(chan error, 1)
	out := make(chan uof.Fixture)
	go func() {
		defer close(out)
		defer close(errc)
		done := false

		parse := func(buf []byte) error {
			var sr scheduleRsp
			if err := xml.Unmarshal(buf, &sr); err != nil {
				return err
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
		}
		if err := parse(buf); err != nil {
			errc <- err
			return
		}

		// than all witch schedules before to
		limit := 1000
		for start := 0; true; start = start + limit {
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
