package uof

import "strings"

type Lang int8

var languages = []struct {
	code  string
	value Lang
	name  string
}{
	{"sqi", LangSQI, "Albanian"},
	{"aa", LangAA, "Arabic"},
	{"az", LangAZE, "Azerbaijan"},
	{"bs", LangBS, "Bosnian"},
	{"br", LangBR, "Brazilian Portuguese"},
	{"bg", LangBG, "Bulgarian"},
	{"zh", LangZH, "Chinese (simplified)"},
	{"zht", LangZHT, "Chinese (traditional)"},
	{"hr", LangHR, "Croatian"},
	{"cz", LangCZ, "Czech"},
	{"da", LangDA, "Danish"},
	{"nl", LangNL, "Dutch"},
	{"en", LangEN, "English"},
	{"et", LangET, "Estonian"},
	{"fi", LangFI, "Finnish"},
	{"fr", LangFR, "French"},
	{"ka", LangKA, "Georgian"},
	{"de", LangDE, "German"},
	{"el", LangEL, "Greek"},
	{"heb", LangHEB, "Hebrew"},
	{"hu", LangHU, "Hungarian"},
	{"Id", LangID, "Indonesian"},
	{"ja", LangJA, "Japanese"},
	{"kaz", LangKAZ, "Kazakh"},
	{"ko", LangKO, "Korean"},
	{"lv", LangLV, "Latvian"},
	{"lt", LangLT, "Lithuanian"},
	{"ml", LangML, "Macedonian"},
	{"no", LangNO, "Norwegian"},
	{"pl", LangPL, "Polish"},
	{"pt", LangPT, "Portuguese"},
	{"ro", LangRO, "Romanian"},
	{"ru", LangRU, "Russian"},
	{"sr", LangSR, "Serbian"},
	{"srl", LangSRL, "Serbian Latin"},
	{"sk", LangSK, "Slovak"},
	{"sl", LangSL, "Slovenian"},
	{"es", LangES, "Spanish"},
	{"se", LangSE, "Swedish"},
	{"th", LangTH, "Thai"},
	{"tr", LangTR, "Turkish"},
	{"ukr", LangUKR, "Ukrainian"},
	{"vi", LangVI, "Vietnamese"},
	{"it", LangIT, "Italian"},
}

const (
	LangNone Lang = iota
	LangSQI       // Albanian
	LangAA        // Arabic
	LangAZE       // Azerbaijan
	LangBS        // Bosnian
	LangBR        // Brazilian Portuguese
	LangBG        // Bulgarian
	LangZH        // Chinese (simplified)
	LangZHT       // Chinese (traditional)
	LangHR        // Croatian
	LangCZ        // Czech
	LangDA        // Danish
	LangNL        // Dutch
	LangEN        // English
	LangET        // Estonian
	LangFI        // Finnish
	LangFR        // French
	LangKA        // Georgian
	LangDE        // German
	LangEL        // Greek
	LangHEB       // Hebrew
	LangHU        // Hungarian
	LangID        // Indonesian
	LangJA        // Japanese
	LangKAZ       // Kazakh
	LangKO        // Korean
	LangLV        // Latvian
	LangLT        // Lithuanian
	LangML        // Macedonian
	LangNO        // Norwegian
	LangPL        // Polish
	LangPT        // Portuguese
	LangRO        // Romanian
	LangRU        // Russian
	LangSR        // Serbian
	LangSRL       // Serbian Latin
	LangSK        // Slovak
	LangSL        // Slovenian
	LangES        // Spanish
	LangSE        // Swedish
	LangTH        // Thai
	LangTR        // Turkish
	LangUKR       // Ukrainian
	LangVI        // Vietnamese
	LangIT        // Italian
)

// Languages transforms comma separatet list of language codes to
// Lang values array
func Languages(codes string) []Lang {
	var values []Lang
	for _, code := range strings.Split(codes, ",") {
		var l Lang
		l.Parse(code)
		if l != LangNone {
			values = append(values, l)
		}
	}
	return values
}

// Parse converts language code to Lang value
func (l *Lang) Parse(code string) {
	for _, v := range languages {
		if v.code == code {
			*l = v.value
			return
		}
	}
}

func (l Lang) String() string {
	return l.Code()
}

func (l Lang) Code() string {
	for _, v := range languages {
		if v.value == l {
			return v.code
		}
	}
	return ""
}

func (l Lang) Name() string {
	for _, v := range languages {
		if v.value == l {
			return v.name
		}
	}
	return ""
}
