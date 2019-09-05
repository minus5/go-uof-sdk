package uof

import (
	"encoding/xml"
	"time"
)

type PlayerProfile struct {
	Player      Player    `xml:"player" json:"player"`
	GeneratedAt time.Time `xml:"generated_at,attr,omitempty" json:"generatedAt,omitempty"`
}

type Player struct {
	ID           int       `xml:"id,attr" json:"id"`
	Type         string    `xml:"type,attr,omitempty" json:"type,omitempty"`
	DateOfBirth  time.Time `xml:"date_of_birth,attr,omitempty" json:"dateOfBirth,omitempty"`
	Nationality  string    `xml:"nationality,attr,omitempty" json:"nationality,omitempty"`
	CountryCode  string    `xml:"country_code,attr,omitempty" json:"countryCode,omitempty"`
	Height       int       `xml:"height,attr,omitempty" json:"height,omitempty"`
	Weight       int       `xml:"weight,attr,omitempty" json:"weight,omitempty"`
	JerseyNumber int       `xml:"jersey_number,attr,omitempty" json:"jerseyNumber,omitempty"`
	Name         string    `xml:"name,attr,omitempty" json:"name,omitempty"`
	FullName     string    `xml:"full_name,attr,omitempty" json:"fullName,omitempty"`
	Nickname     string    `xml:"nickname,attr,omitempty" json:"nickname,omitempty"`
	Gender       Gender    `xml:"gender,attr,omitempty" json:"gender,omitempty"`
}

func (t *Player) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type T Player
	var overlay struct {
		*T
		ID          string `xml:"id,attr" json:"id"`
		DateOfBirth string `xml:"date_of_birth,attr,omitempty" json:"dateOfBirth,omitempty"`
		Gender      string `xml:"gender,attr,omitempty" json:"gender,omitempty"`
	}
	overlay.T = (*T)(t)
	if err := d.DecodeElement(&overlay, &start); err != nil {
		return err
	}
	t.ID = toPlayerID(overlay.ID)
	t.DateOfBirth = dateToTime(overlay.DateOfBirth)
	t.Gender = toGender(overlay.Gender)
	return nil
}

func toGender(g string) Gender {
	switch g {
	case "male":
		return Male
	case "female":
		return Female
	default:
		return GenderUnknown
	}
}

const apiDateFormat = "2006-01-02"

func dateToTime(date string) time.Time {
	t, _ := time.Parse(apiDateFormat, date)
	return t
}
