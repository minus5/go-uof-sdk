package uof

import (
	"encoding/xml"
	"strings"
)

// Markets betradar api response
type MarketsRsp struct {
	Markets MarketDescriptions `xml:"market,omitempty" json:"markets,omitempty"`
	// unused
	//ResponseCode string   `xml:"response_code,attr,omitempty" json:"responseCode,omitempty"`
	//Location     string   `xml:"location,attr,omitempty" json:"location,omitempty"`
}

type MarketDescriptions []MarketDescription

func (md MarketDescriptions) Find(id int) *MarketDescription {
	for _, m := range md {
		if m.ID == id {
			return &m
		}
	}
	return nil
}

func (md MarketDescriptions) Groups() map[string][]int {
	marketGroups := make(map[string][]int)
	for _, m := range md {
		for _, group := range m.Groups {
			a := marketGroups[group]
			a = append(a, m.ID)
			marketGroups[group] = a
		}
	}
	return marketGroups
}

type MarketDescription struct {
	ID                     int               `xml:"id,attr" json:"id"`
	VariantID              int               `json:"variantID,omitempty"`
	Name                   string            `xml:"name,attr" json:"name,omitempty"`
	Description            string            `xml:"description,attr,omitempty" json:"description,omitempty"`
	IncludesOutcomesOfType string            `xml:"includes_outcomes_of_type,attr,omitempty" json:"includesOutcomesOfType,omitempty"`
	Variant                string            `xml:"variant,attr,omitempty" json:"variant,omitempty"`
	OutcomeType            OutcomeType       `json:"outcomeType,omitempty"`
	Groups                 []string          `json:"groups,omitempty"`
	Outcomes               []MarketOutcome   `xml:"outcomes>outcome,omitempty" json:"outcomes,omitempty"`
	Specifiers             []MarketSpecifier `xml:"specifiers>specifier,omitempty" json:"specifiers,omitempty"`
	Attributes             []MarketAttribute `xml:"attributes>attribute,omitempty" json:"attributes,omitempty"`
	//Mappings               []Mapping         `xml:"mappings>mapping,omitempty" json:"mappings,omitempty"`
}

type MarketOutcome struct {
	ID          int    `json:"id"`
	Name        string `xml:"name,attr" json:"name,omitempty"`
	Description string `xml:"description,attr,omitempty" json:"description,omitempty"`
}

type MarketSpecifier struct {
	Type        SpecifierType `json:"type"`
	Name        string        `xml:"name,attr" json:"name,omitempty"`
	Description string        `xml:"description,attr,omitempty" json:"description,omitempty"`
}

type MarketAttribute struct {
	Name        string `xml:"name,attr" json:"name,omitempty"`
	Description string `xml:"description,attr" json:"description,omitempty"`
}

// // currently unused but parsing is valid
// type Mapping struct {
// 	MappingOutcome []MappingOutcome `xml:"mapping_outcome,omitempty" json:"mappingOutcome,omitempty"`
// 	ProductID      int              `xml:"product_id,attr" json:"productID"`
// 	ProductIDs     string           `xml:"product_ids,attr" json:"productIDs"`
// 	SportID        string           `xml:"sport_id,attr" json:"sportID"`
// 	MarketID       string           `xml:"market_id,attr" json:"marketID"`
// 	SovTemplate    string           `xml:"sov_template,attr,omitempty" json:"sovTemplate,omitempty"`
// 	ValidFor       string           `xml:"valid_for,attr,omitempty" json:"validFor,omitempty"`
// }

// type MappingOutcome struct {
// 	OutcomeID          string `xml:"outcome_id,attr" json:"outcomeID"`
// 	ProductOutcomeID   string `xml:"product_outcome_id,attr" json:"productOutcomeID"`
// 	ProductOutcomeName string `xml:"product_outcome_name,attr,omitempty" json:"productOutcomeName,omitempty"`
// }

func (t *MarketDescription) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type T MarketDescription
	var overlay struct {
		*T
		Groups      string `xml:"groups,attr"`
		OutcomeType string `xml:"outcome_type,attr,omitempty"`
	}
	overlay.T = (*T)(t)
	if err := d.DecodeElement(&overlay, &start); err != nil {
		return err
	}
	t.VariantID = toVariantID(overlay.Variant)
	t.Groups = toGroups(overlay.Groups)
	t.OutcomeType = toOutcomeType(overlay.OutcomeType)
	return nil
}

func (t *MarketOutcome) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type T MarketOutcome
	var overlay struct {
		*T
		ID string `xml:"id,attr"`
	}
	overlay.T = (*T)(t)
	if err := d.DecodeElement(&overlay, &start); err != nil {
		return err
	}
	t.ID = toOutcomeID(overlay.ID)
	return nil
}

func (t *MarketSpecifier) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type T MarketSpecifier
	var overlay struct {
		*T
		Type string `xml:"type,attr"`
	}
	overlay.T = (*T)(t)
	if err := d.DecodeElement(&overlay, &start); err != nil {
		return err
	}
	t.Type = toSpecifierType(overlay.Type)
	return nil
}

func toVariantID(id string) int {
	if id == "" {
		return 0
	}
	return hash32(id)
}

func toGroups(groups string) []string {
	if groups == "" {
		return nil
	}
	sg := strings.Split(groups, "|") // split string to array
	for i, g := range sg {           // remove all group
		if g == "all" {
			sg = append(sg[:i], sg[i+1:]...)
		}
	}
	if len(sg) == 0 {
		return nil
	}

	return sg
}

func toOutcomeType(ot string) OutcomeType {
	switch ot {
	case "":
		return OutcomeTypeDefault
	case "player":
		return OutcomeTypePlayer
	case "competitor":
		return OutcomeTypeCompetitor
	case "competitors":
		return OutcomeTypeCompetitors
	case "free_text":
		return OutcomeTypeFreeText
	default:
		return OutcomeTypeUnknown
	}
}

func toSpecifierType(st string) SpecifierType {
	switch st {
	case "string":
		return SpecifierTypeString
	case "integer":
		return SpecifierTypeInteger
	case "decimal":
		return SpecifierTypeDecimal
	case "variable_text":
		return SpecifierTypeVariableText
	default:
		return SpecifierTypeUnknown
	}
}
