package uof

import (
	"encoding/xml"
	"time"
)

//You will receive a fixture change when you book a match, and also when/if the
// match is added to the live odds program.
// A fixture_change message is sent when a Betradar system has made a fixture
// change it deems is important. These are typically changes that affect events
// in the near-term (e.g. a match was added that starts in the next few minutes,
// a match was delayed and starts in a couple of minutes, etc.). The message is
// short and includes a bare minimum of relevant details about the
// addition/change. The recommended practice is to always to a follow-up API
// call to lookup the updated fixture information.
// Reference: https://docs.betradar.com/display/BD/UOF+-+Fixture+change
type FixtureChange struct {
	EventID      int                `json:"eventID"`
	EventURN     URN                `xml:"event_id,attr" json:"eventURN"`
	Producer     Producer           `xml:"product,attr" json:"producer"`
	Timestamp    int64              `xml:"timestamp,attr" json:"timestamp"`
	RequestID    *int               `xml:"request_id,attr,omitempty" json:"requestID,omitempty"`
	ChangeType   *FixtureChangeType `xml:"change_type,attr,omitempty" json:"changeType,omitempty"`
	StartTime    *int64             `xml:"start_time,attr" json:"startTime"`
	NextLiveTime *int64             `xml:"next_live_time,attr,omitempty" json:"nextLiveTime,omitempty"`
}

func (t *FixtureChange) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type T FixtureChange
	var overlay struct {
		*T
	}
	overlay.T = (*T)(t)
	if err := d.DecodeElement(&overlay, &start); err != nil {
		return err
	}
	t.EventID = t.EventURN.ID()
	return nil
}

func (fc *FixtureChange) Schedule() *time.Time {
	if fc.StartTime == nil {
		return nil
	}
	ts := time.Unix(0, *fc.StartTime*int64(time.Millisecond))
	return &ts
}
