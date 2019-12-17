package uof

import (
	"encoding/xml"
)

// The bet_stop message is an optimized signal to indicate that all, or a set of
// markets should be instantly suspended (continue to display odds, but don't
// accept tickets). The same effect can also be achieved by sending an
// odds_change message that lists all the affected markets and moves them to
// status="-1" (suspended).
// It is important to keep in mind that only active
// markets should be set to suspended, and not markets that are already
// deactivated, settled or cancelled. This is also the case for the attribute
// market_status (explained HERE). If it is not present, the market should be
// moved to suspended. However, if the market is already deactivated, settled or
// cancelled this is not a good practice. Only move ACTIVE markets to suspended.
type BetStop struct {
	EventID   int          `json:"eventID"`
	EventURN  URN          `xml:"event_id,attr" json:"eventURN"`
	Timestamp int          `xml:"timestamp,attr" json:"timestamp"`
	RequestID *int         `xml:"request_id,attr,omitempty" json:"requestID,omitempty"`
	Groups    []string     `json:"groups"`
	MarketIDs []int        `json:"marketsIDs"`
	Producer  Producer     `xml:"product,attr" json:"producer"`
	Status    MarketStatus `json:"status,omitempty"`
}

func (t *BetStop) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type T BetStop
	var overlay struct {
		*T
		Groups       string `xml:"groups,attr" json:"groups"`
		MarketStatus *int   `xml:"market_status,attr,omitempty"` // May be one of 1, 0, -1, -2, -3, -4
	}
	overlay.T = (*T)(t)
	if err := d.DecodeElement(&overlay, &start); err != nil {
		return err
	}
	t.EventID = t.EventURN.EventID()
	t.Status = toMarketStatus(overlay.MarketStatus)
	t.Groups = toGroups(overlay.Groups)
	return nil
}

// If not present, the markets specified should be moved to suspended. If
// present, they should be either suspended or deactivated based on the value of
// this field.
func toMarketStatus(ms *int) MarketStatus {
	if ms == nil {
		return MarketStatusSuspended
	}
	switch *ms {
	case 0:
		return MarketStatusInactive
	case 1:
		return MarketStatusActive
	case -1:
		return MarketStatusSuspended
	case -2:
		return MarketStatusHandedOver
	case -3:
		return MarketStatusSettled
	case -4:
		return MarketStatusCancelled
	default:
		return MarketStatusInactive
	}
}
