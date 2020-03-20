package uof

import (
	"time"
)

type CompetitorProfile struct {
	Competitor  CompetitorPlayer `xml:"competitor" json:"competitor"`
	GeneratedAt time.Time        `xml:"generated_at,attr,omitempty" json:"generatedAt,omitempty"`
}
