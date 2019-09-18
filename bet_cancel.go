package uof

import "encoding/xml"

// A bet_cancel message is sent when a bet made on a particular market needs to
// be cancelled and refunded due to an error (which is different to a
// bet-settlement/refund).
// Reference: https://docs.betradar.com/display/BD/UOF+-+Bet+cancel
type BetCancel struct {
	EventID   int      `json:"eventID"`
	EventURN  URN      `xml:"event_id,attr" json:"eventURN"`
	Producer  Producer `xml:"product,attr" json:"producer"`
	Timestamp int      `xml:"timestamp,attr" json:"timestamp"`
	RequestID *int     `xml:"request_id,attr,omitempty" json:"requestID,omitempty"`
	// If start and end time are specified, they designate a range in time for
	// which bets made should be cancelled. If there is an end_time but no
	// start_time, this means cancel all bets placed before the specified time. If
	// there is a start_time but no end_time this means, cancel all bets placed
	// after the specified start_time.
	StartTime    *int              `xml:"start_time,attr,omitempty" json:"startTime,omitempty"`
	EndTime      *int              `xml:"end_time,attr,omitempty" json:"endTime,omitempty"`
	SupercededBy *string           `xml:"superceded_by,attr,omitempty" json:"supercededBy,omitempty"`
	Markets      []BetCancelMarket `xml:"market" json:"market"`
}

type BetCancelMarket struct {
	ID         int  `xml:"id,attr" json:"id"`
	LineID     int  `json:"lineID"`
	VoidReason *int `xml:"void_reason,attr,omitempty" json:"voidReason,omitempty"`
}

// A Rollback_bet_cancel message is sent when a previous bet cancel should be
// undone (if possible). This may happen, for example, if a Betradar operator
// mistakenly cancels the wrong market (resulting in a bet_cancel being sent)
// during the game; before realizing the mistake.
type RollbackBetCancel struct {
	EventID   int               `json:"eventID"`
	EventURN  URN               `xml:"event_id,attr" json:"eventURN"`
	Producer  Producer          `xml:"product,attr" json:"producer"`
	Timestamp int               `xml:"timestamp,attr" json:"timestamp"`
	RequestID *int              `xml:"request_id,attr,omitempty" json:"requestID,omitempty"`
	StartTime *int              `xml:"start_time,attr,omitempty" json:"startTime,omitempty"`
	EndTime   *int              `xml:"end_time,attr,omitempty" json:"endTime,omitempty"`
	Markets   []BetCancelMarket `xml:"market" json:"market"`
}

// UnmarshalXML
func (t *BetCancel) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type T BetCancel
	var overlay struct {
		*T
	}
	overlay.T = (*T)(t)
	if err := d.DecodeElement(&overlay, &start); err != nil {
		return err
	}
	t.EventID = t.EventURN.EventID()
	return nil
}

// UnmarshalXML
func (t *BetCancelMarket) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type T BetCancelMarket
	var overlay struct {
		*T
		Specifiers         string `xml:"specifiers,attr,omitempty"`
		ExtendedSpecifiers string `xml:"extended_specifiers,attr,omitempty"`
	}
	overlay.T = (*T)(t)
	if err := d.DecodeElement(&overlay, &start); err != nil {
		return err
	}
	t.LineID = toLineID(overlay.Specifiers)
	return nil
}

// UnmarshalXML
func (t *RollbackBetCancel) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type T RollbackBetCancel
	var overlay struct {
		*T
	}
	overlay.T = (*T)(t)
	if err := d.DecodeElement(&overlay, &start); err != nil {
		return err
	}
	t.EventID = t.EventURN.EventID()
	return nil
}
