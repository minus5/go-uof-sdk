package uof

import (
	"encoding/xml"
	"sort"
	"strconv"
	"strings"
)

// OddsChange messages are sent whenever Betradar has new odds for some markets
// for a match. Odds changes can include a subset of all markets; if so, markets
// not reported remain unchanged. All outcomes possible within a market are
// reported.
// Reference: https://docs.betradar.com/display/BD/UOF+-+Odds+change
type OddsChange struct {
	EventID  int `json:"eventID"`
	EventURN URN `xml:"event_id,attr" json:"eventURN"`
	// Specifies which producer generated these odds. At any given point in time
	// there should only be one product generating odds for a particular event.
	Producer  Producer `xml:"product,attr" json:"producer"`
	Timestamp int      `xml:"timestamp,attr" json:"timestamp"`
	Markets   []Market `json:"market,omitempty"`
	// values in range 0-6   /v1/descriptions/betting_status.xml
	BettingStatus *int `json:"bettingStatus,omitempty"`
	// values in range 0-87  /v1/descriptions/betstop_reasons.xml
	BetstopReason    *int              `json:"betstopReason,omitempty"`
	OddsChangeReason *int              `xml:"odds_change_reason,attr,omitempty" json:"oddsChangeReason,omitempty"` // May be one of 1
	EventStatus      *SportEventStatus `xml:"sport_event_status,omitempty" json:"sportEventStatus,omitempty"`

	OddsGenerationProperties *OddsGenerationProperties `xml:"odds_generation_properties,omitempty" json:"oddsGenerationProperties,omitempty"`
	RequestID                *int                      `xml:"request_id,attr,omitempty" json:"requestID,omitempty"`
}

// Provided by the prematch odds producer only, and contains a few
// key-parameters that can be used in a client’s own special odds model, or
// even offer spread betting bets based on it.
type OddsGenerationProperties struct {
	ExpectedTotals    *float64 `xml:"expected_totals,attr,omitempty" json:"expectedTotals,omitempty"`
	ExpectedSupremacy *float64 `xml:"expected_supremacy,attr,omitempty" json:"expectedSupremacy,omitempty"`
}

// Market describes the odds updates for a particular market.
// Betradar Unified Odds utilizes markets and market lines. Each market is a bet
// type identified with a unique ID and within a market, multiple different lines
// are often provided. Each of these lines is uniquely identified by additional
// specifiers (e.g. Total Goals 2.5 is the same market as Total Goals 1.5, but it
// is two different market lines. The market ID for both are the same, but the
// first one has a specifier ((goals=2.5)) and the other one has a specifier
// ((goals=1.5)) that uniquely identifies them).
// LineID is hash of specifier field used to uniquely identify lines in one market.
// One market line is uniquely identified by market id and line id.
type Market struct {
	ID            int               `xml:"id,attr" json:"id"`
	LineID        int               `json:"lineID"`
	Specifiers    map[string]string `json:"sepcifiers,omitempty"`
	Status        MarketStatus      `xml:"status,attr,omitempty" json:"status,omitempty"`
	CashoutStatus *CashoutStatus    `xml:"cashout_status,attr,omitempty" json:"cashoutStatus,omitempty"`
	// If present, this is set to 1, which states that this is the most balanced
	// or recommended market line. This setting makes most sense for markets where
	// multiple lines are provided (e.g. the Totals market).
	Favourite *bool     `xml:"favourite,attr,omitempty" json:"favourite,omitempty"`
	Outcomes  []Outcome `xml:"outcome,omitempty" json:"outcome,omitempty"`
	// Timestamp in UTC when to betstop this market. Typically used for outrights
	// and typically is the start-time of the event the market refers to.
	NextBetstop *int `json:"nextBetstop,omitempty"`
}
type MarketMetadata struct {
	NextBetstop *int `xml:"next_betstop,attr,omitempty" json:"nextBetstop,omitempty"`
}

type Outcome struct {
	ID            int      `json:"id"`
	PlayerID      int      `json:"playerID"`
	Competitors   []int    `json:"competitors"`
	Odds          *float64 `xml:"odds,attr,omitempty" json:"odds,omitempty"`
	Probabilities *float64 `xml:"probabilities,attr,omitempty" json:"probabilities,omitempty"`
	Active        *bool    `xml:"active,attr,omitempty" json:"active,omitempty"`
	Team          *Team    `xml:"team,attr,omitempty" json:"team,omitempty"`
}

// UnmarshalXML
func (o *OddsChange) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type T OddsChange
	var overlay struct {
		*T
		Odds *struct {
			Markets       []Market `xml:"market,omitempty"`
			BettingStatus *int     `xml:"betting_status,attr,omitempty"`
			BetstopReason *int     `xml:"betstop_reason,attr,omitempty"`
		} `xml:"odds,omitempty"`
	}
	overlay.T = (*T)(o)
	if err := d.DecodeElement(&overlay, &start); err != nil {
		return err
	}
	if overlay.Odds != nil {
		o.BettingStatus = overlay.Odds.BettingStatus
		o.BetstopReason = overlay.Odds.BetstopReason
		o.Markets = overlay.Odds.Markets
	}
	o.EventID = o.EventURN.EventID()
	return nil
}

// Custom unmarshaling reasons:
//  * To cover the case that: 'The default value is active if status is not present.'
//  * To convert Specifiers and ExtendedSpecifiers fileds which are
//    lists of key value attributes encoded in string to the map.
//  * To calculate LineID; market line is uniquely identified by both
//    market id and line id
func (m *Market) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type T Market
	var overlay struct {
		*T
		Status             *int8  `xml:"status,attr,omitempty"`
		Specifiers         string `xml:"specifiers,attr,omitempty" json:"specifiers,omitempty"`
		ExtendedSpecifiers string `xml:"extended_specifiers,attr,omitempty" json:"extendedSpecifiers,omitempty"`
		MarketMetadata     *struct {
			NextBetstop *int `xml:"next_betstop,attr,omitempty"`
		} `xml:"market_metadata,omitempty"`
	}
	overlay.T = (*T)(m)
	if err := d.DecodeElement(&overlay, &start); err != nil {
		return err
	}
	m.Status = MarketStatusActive // default
	if overlay.Status != nil {
		m.Status = MarketStatus(*overlay.Status)
	}
	m.Specifiers = toSpecifiers(overlay.Specifiers, overlay.ExtendedSpecifiers)
	m.LineID = toLineID(overlay.Specifiers)
	if overlay.MarketMetadata != nil {
		m.NextBetstop = overlay.MarketMetadata.NextBetstop
	}
	return nil
}

func (t *Outcome) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type T Outcome
	var overlay struct {
		*T
		ID string `xml:"id,attr" json:"urn"`
	}
	overlay.T = (*T)(t)
	if err := d.DecodeElement(&overlay, &start); err != nil {
		return err
	}
	t.ID = toOutcomeID(overlay.ID)
	t.PlayerID = toPlayerID(overlay.ID)
	t.Competitors = toComptitors(overlay.ID)
	return nil
}

func (m Market) VariantSpecifier() string {
	for k, v := range m.Specifiers {
		if k == "variant" {
			return v
		}
	}
	return ""
}

func toSpecifiers(specifiers, extendedSpecifiers string) map[string]string {
	allSpecifiers := specifiers
	if extendedSpecifiers != "" {
		allSpecifiers = allSpecifiers + "|" + extendedSpecifiers
	}
	if len(allSpecifiers) < 2 {
		return nil
	}
	sm := make(map[string]string)
	for _, s := range strings.Split(allSpecifiers, "|") {
		if p := strings.Split(s, "="); len(p) == 2 {
			k := p[0]
			v := p[1]
			if k == "player" {
				v = strings.TrimPrefix(v, srPlayer)
			}
			sm[k] = v
		}
	}
	return sm
}

func toPlayerID(id string) int {
	if strings.HasPrefix(id, srPlayer) {
		return URN(id).ID()
	}
	return 0
}

func toComptitors(id string) []int {
	if !strings.Contains(id, srCompetitor) {
		return nil
	}
	var competitors []int
	for _, p := range strings.Split(id, ",") {
		v := strings.TrimPrefix(p, srCompetitor)
		if k, err := strconv.Atoi(v); err == nil {
			competitors = append(competitors, k)
		}
	}
	if len(competitors) == 0 {
		return nil
	}
	return competitors
}

func toOutcomeID(id string) int {
	if strings.HasPrefix(id, srPlayer) {
		return toPlayerID(id)
	}
	if i, err := strconv.ParseInt(id, 10, 64); err == nil {
		return int(i)
	}
	return hash32(id)
}

func (o *OddsChange) EachPlayer(handler func(int)) {
	if o == nil {
		return
	}
	for _, m := range o.Markets {
		for _, o := range m.Outcomes {
			if id := o.PlayerID; id != 0 {
				handler(id)
			}
		}
		// fetch player if provided as market specifier
		// <market id="888" specifiers="player=sr:player:575270">
		// <market id="891" specifiers="goalnr=1|player=sr:player:833167">
		if playerID, ok := m.Specifiers["player"]; ok {
			id, err := strconv.Atoi(playerID)
			if err == nil {
				handler(id)
			}
		}
	}
}

func (o *OddsChange) EachCompetitor(handler func(int)) {
	if o == nil {
		return
	}
	for _, id := range o.Competitors() {
		handler(id)
	}
}

// Competitors collects all competitors from outcomes
func (o *OddsChange) Competitors() []int {
	cm := make(map[int]struct{})
	for _, m := range o.Markets {
		for _, o := range m.Outcomes {
			if cs := o.Competitors; cs != nil {
				for _, c := range cs {
					cm[c] = struct{}{}
				}

			}
		}
	}
	if len(cm) == 0 {
		return nil
	}
	ca := make([]int, 0, len(cm))
	for id := range cm {
		ca = append(ca, id)
	}
	sort.Ints(ca)
	return ca
}

func (o *OddsChange) EachVariantMarket(handler func(int, string)) {
	if o == nil {
		return
	}
	for _, m := range o.Markets {
		if s := m.VariantSpecifier(); s != "" {
			handler(m.ID, s)
		}
	}
}
