package uof

import (
	"encoding/xml"
	"fmt"
	"time"
)

// FixtureRsp response
type FixtureRsp struct {
	Fixture     Fixture   `xml:"fixture" json:"fixture"`
	GeneratedAt time.Time `xml:"generated_at,attr,omitempty" json:"generatedAt,omitempty"`
}

// Fixture describe static or semi-static information about matches and races.
// Reference: https://docs.betradar.com/display/BD/UOF+-+Fixtures+in+the+API
type Fixture struct {
	ID                 int       `xml:"-" json:"id"`
	URN                URN       `xml:"id,attr,omitempty" json:"urn"`
	StartTime          time.Time `xml:"start_time,attr,omitempty" json:"startTime,omitempty"`
	StartTimeConfirmed bool      `xml:"start_time_confirmed,attr,omitempty" json:"startTimeConfirmed,omitempty"`
	StartTimeTbd       bool      `xml:"start_time_tbd,attr,omitempty" json:"startTimeTbd,omitempty"`
	NextLiveTime       time.Time `xml:"next_live_time,attr,omitempty" json:"nextLiveTime,omitempty"`
	Liveodds           string    `xml:"liveodds,attr,omitempty" json:"liveodds,omitempty"`
	Status             string    `xml:"status,attr,omitempty" json:"status,omitempty"`
	Name               string    `xml:"name,attr,omitempty" json:"name,omitempty"`
	Type               string    `xml:"type,attr,omitempty" json:"type,omitempty"`
	Scheduled          time.Time `xml:"scheduled,attr,omitempty" json:"scheduled,omitempty"`
	ScheduledEnd       time.Time `xml:"scheduled_end,attr,omitempty" json:"scheduledEnd,omitempty"`
	ReplacedBy         string    `xml:"replaced_by,attr,omitempty" json:"replacedBy,omitempty"`

	Sport      Sport      `xml:"sport" json:"sport"`
	Category   Category   `xml:"category" json:"category"`
	Tournament Tournament `xml:"tournament,omitempty" json:"tournament,omitempty"`

	Round  Round  `xml:"tournament_round,omitempty" json:"round,omitempty"`
	Season Season `xml:"season,omitempty" json:"season,omitempty"`
	Venue  Venue  `xml:"venue,omitempty" json:"venue,omitempty"`

	ProductInfo ProductInfo  `xml:"product_info,omitempty" json:"productInfo,omitempty"`
	Competitors []Competitor `xml:"competitors>competitor,omitempty" json:"competitors,omitempty"`
	TvChannels  []TvChannel  `xml:"tv_channels>tv_channel,omitempty" json:"tvChannels,omitempty"`

	Home Competitor `json:"home"`
	Away Competitor `json:"away"`
	// this also exists but we are skiping for the time being
	//ReferenceIDs         ReferenceIDs         `xml:"reference_ids,omitempty" json:"referenceIDs,omitempty"`
	//SportEventConditions SportEventConditions `xml:"sport_event_conditions,omitempty" json:"sportEventConditions,omitempty"`
	//DelayedInfo DelayedInfo `xml:"delayed_info,omitempty" json:"delayedInfo,omitempty"`
	//CoverageInfo CoverageInfo `xml:"coverage_info,omitempty" json:"coverageInfo,omitempty"`
	//Races        []SportEvent `xml:"races>sport_event,omitempty" json:"races,omitempty"`
	//ExtraInfo   ExtraInfo    `xml:"extra_info,omitempty" json:"extraInfo,omitempty"`
	//ScheduledStartTimeChanges []ScheduledStartTimeChange `xml:"scheduled_start_time_changes>scheduled_start_time_change,omitempty" json:"scheduledStartTimeChanges,omitempty"`
	//Parent *ParentStage `xml:"parent,omitempty" json:"parent,omitempty"`

}

// Tournament structure
type Tournament struct {
	ID   int    `json:"id"`
	Name string `xml:"name,attr" json:"name"`
}

// Sport structure
type Sport struct {
	ID   int    `json:"id"`
	Name string `xml:"name,attr" json:"name"`
}

// Category structure
// CountryCode is the three-letter ISO country-code
type Category struct {
	ID          int    `json:"id"`
	Name        string `xml:"name,attr" json:"name"`
	CountryCode string `xml:"country_code,attr,omitempty" json:"countryCode,omitempty"`
}

// Competitor structure
type Competitor struct {
	ID           int                `json:"id"`
	Qualifier    string             `xml:"qualifier,attr,omitempty" json:"qualifier,omitempty"`
	Name         string             `xml:"name,attr" json:"name"`
	Abbreviation string             `xml:"abbreviation,attr" json:"abbreviation"`
	Country      string             `xml:"country,attr,omitempty" json:"country,omitempty"`
	CountryCode  string             `xml:"country_code,attr,omitempty" json:"countryCode,omitempty"`
	Virtual      bool               `xml:"virtual,attr,omitempty" json:"virtual,omitempty"`
	Players      []CompetitorPlayer `xml:"players>player,omitempty" json:"players,omitempty"`
	//ReferenceIDs CompetitorReferenceIDs `xml:"reference_ids,omitempty" json:"referenceIDs,omitempty"`
}

// CompetitorPlayer structure
type CompetitorPlayer struct {
	ID           int    `json:"id"`
	Name         string `xml:"name,attr" json:"name"`
	Abbreviation string `xml:"abbreviation,attr" json:"abbreviation"`
	Nationality  string `xml:"nationality,attr,omitempty" json:"nationality,omitempty"`
}

// Venue represent where the event is taking place
type Venue struct {
	ID             int    `json:"id"`
	Name           string `xml:"name,attr" json:"name"`
	Capacity       int    `xml:"capacity,attr,omitempty" json:"capacity,omitempty"`
	CityName       string `xml:"city_name,attr,omitempty" json:"cityName,omitempty"`
	CountryName    string `xml:"country_name,attr,omitempty" json:"countryName,omitempty"`
	CountryCode    string `xml:"country_code,attr,omitempty" json:"countryCode,omitempty"`
	MapCoordinates string `xml:"map_coordinates,attr,omitempty" json:"mapCoordinates,omitempty"`
}

// TvChannel list of TV channels
type TvChannel struct {
	Name string `xml:"name,attr" json:"name"`
	// seams to be always zero
	// StartTime time.Time `xml:"start_time,attr,omitempty" json:"startTime,omitempty"`
}

// StreamingChannel details about streaming offering
type StreamingChannel struct {
	ID   int    `xml:"id,attr" json:"id"`
	Name string `xml:"name,attr" json:"name"`
}

// ProductInfoLink links to various pages within Betradar’s hosted solution offering for particular event
type ProductInfoLink struct {
	Name string `xml:"name,attr" json:"name"`
	Ref  string `xml:"ref,attr" json:"ref"`
}

// Round tournament info
type Round struct {
	ID                  int    `xml:"betradar_id,attr,omitempty" json:"id,omitempty"`
	Type                string `xml:"type,attr,omitempty" json:"type,omitempty"`
	Number              int    `xml:"number,attr,omitempty" json:"number,omitempty"`
	Name                string `xml:"name,attr,omitempty" json:"name,omitempty"`
	GroupLongName       string `xml:"group_long_name,attr,omitempty" json:"groupLongName,omitempty"`
	Group               string `xml:"group,attr,omitempty" json:"group,omitempty"`
	GroupID             string `xml:"group_id,attr,omitempty" json:"groupID,omitempty"`
	CupRoundMatches     int    `xml:"cup_round_matches,attr,omitempty" json:"cupRoundMatches,omitempty"`
	CupRoundMatchNumber int    `xml:"cup_round_match_number,attr,omitempty" json:"cupRoundMatchNumber,omitempty"`
	OtherMatchID        string `xml:"other_match_id,attr,omitempty" json:"otherMatchID,omitempty"`
}

// Season list structure
type Season struct {
	ID        int       `json:"id"`
	StartDate string    `xml:"start_date,attr" json:"startDate"`
	EndDate   string    `xml:"end_date,attr" json:"endDate"`
	StartTime time.Time `xml:"start_time,attr,omitempty" json:"startTime,omitempty"`
	EndTime   time.Time `xml:"end_time,attr,omitempty" json:"endTime,omitempty"`
	Year      string    `xml:"year,attr,omitempty" json:"year,omitempty"`
	Name      string    `xml:"name,attr" json:"name"`
	//TournamentID string    `xml:"tournament_id,attr,omitempty" json:"tournamentID,omitempty"`
}

// type ParentStage struct {
// 	URN          URN       `xml:"id,attr,omitempty" json:"urn,omitempty"`
// 	Name         string    `xml:"name,attr,omitempty" json:"name,omitempty"`
// 	Type         string    `xml:"type,attr,omitempty" json:"type,omitempty"`
// 	Scheduled    time.Time `xml:"scheduled,attr,omitempty" json:"scheduled,omitempty"`
// 	StartTimeTbd bool      `xml:"start_time_tbd,attr,omitempty" json:"startTimeTbd,omitempty"`
// 	ScheduledEnd time.Time `xml:"scheduled_end,attr,omitempty" json:"scheduledEnd,omitempty"`
// 	ReplacedBy   string    `xml:"replaced_by,attr,omitempty" json:"replacedBy,omitempty"`
// }

// type ScheduledStartTimeChange struct {
// 	OldTime   time.Time `xml:"old_time,attr" json:"oldTime"`
// 	NewTime   time.Time `xml:"new_time,attr" json:"newTime"`
// 	ChangedAt time.Time `xml:"changed_at,attr" json:"changedAt"`
// }

// ProductInfo lists additional information about the fixture found inside the product_info attribute
type ProductInfo struct {
	Streaming            []StreamingChannel `xml:"streaming>channel,omitempty" json:"streaming,omitempty"`
	IsInLiveScore        string             `xml:"is_in_live_score,omitempty" json:"isInLiveScore,omitempty"`
	IsInHostedStatistics string             `xml:"is_in_hosted_statistics,omitempty" json:"isInHostedStatistics,omitempty"`
	IsInLiveCenterSoccer string             `xml:"is_in_live_center_soccer,omitempty" json:"isInLiveCenterSoccer,omitempty"`
	IsAutoTraded         string             `xml:"is_auto_traded,omitempty" json:"isAutoTraded,omitempty"`
	Links                []ProductInfoLink  `xml:"links>link,omitempty" json:"links,omitempty"`
}

// UnmarshalXML *Fixture data
func (t *Fixture) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type T Fixture
	var overlay struct {
		*T
		Tournament *struct {
			URN      URN      `xml:"id,attr"`
			Name     string   `xml:"name,attr"`
			Sport    Sport    `xml:"sport"`
			Category Category `xml:"category"`
		} `xml:"tournament,omitempty"`
	}
	overlay.T = (*T)(t)
	if err := d.DecodeElement(&overlay, &start); err != nil {
		return err
	}
	t.ID = overlay.URN.EventID()
	t.Sport = overlay.Tournament.Sport
	t.Category = overlay.Tournament.Category
	t.Tournament.ID = overlay.Tournament.URN.ID()
	t.Tournament.Name = overlay.Tournament.Name

	for _, c := range t.Competitors {
		if c.Qualifier == "home" {
			t.Home = c
		}
		if c.Qualifier == "away" {
			t.Away = c
		}
	}
	return nil
}

// UnmarshalXML *Sport data
func (t *Sport) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type T Sport
	var overlay struct {
		*T
		URN URN `xml:"id,attr"`
	}
	overlay.T = (*T)(t)
	if err := d.DecodeElement(&overlay, &start); err != nil {
		return err
	}
	t.ID = overlay.URN.ID()
	return nil
}

// UnmarshalXML *Category data
func (t *Category) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type T Category
	var overlay struct {
		*T
		URN URN `xml:"id,attr"`
	}
	overlay.T = (*T)(t)
	if err := d.DecodeElement(&overlay, &start); err != nil {
		return err
	}
	t.ID = overlay.URN.ID()
	return nil
}

// UnmarshalXML *Season data
func (t *Season) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type T Season
	var overlay struct {
		*T
		URN URN `xml:"id,attr"`
	}
	overlay.T = (*T)(t)
	if err := d.DecodeElement(&overlay, &start); err != nil {
		return err
	}
	t.ID = overlay.URN.ID()
	return nil
}

// UnmarshalXML *Venue data
func (t *Venue) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type T Venue
	var overlay struct {
		*T
		URN URN `xml:"id,attr"`
	}
	overlay.T = (*T)(t)
	if err := d.DecodeElement(&overlay, &start); err != nil {
		return err
	}
	t.ID = overlay.URN.ID()
	return nil
}

// UnmarshalXML *Competitor data
func (t *Competitor) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type T Competitor
	var overlay struct {
		*T
		URN URN `xml:"id,attr"`
	}
	overlay.T = (*T)(t)
	if err := d.DecodeElement(&overlay, &start); err != nil {
		return err
	}
	t.ID = overlay.URN.ID()
	return nil
}

// UnmarshalXML *CompetitorPlayer data
func (t *CompetitorPlayer) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type T CompetitorPlayer
	var overlay struct {
		*T
		URN URN `xml:"id,attr"`
	}
	overlay.T = (*T)(t)
	if err := d.DecodeElement(&overlay, &start); err != nil {
		return err
	}
	t.ID = overlay.URN.ID()
	return nil
}

// PP pretty prints fixure row
func (t *Fixture) PP() string {
	name := fmt.Sprintf("%s - %s", t.Home.Name, t.Away.Name)
	return fmt.Sprintf("%-90s %12s %15s", name, t.Scheduled.Format("02.01. 15:04"), t.Status)
}
