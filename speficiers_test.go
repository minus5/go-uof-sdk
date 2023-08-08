package uof

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseNameWithSpecifiers(t *testing.T) {
	players := map[int]Player{
		1234: Player{
			FullName: "John Rodriquez",
		},
	}
	fixture := Fixture{
		Name: "Euro2016",
		Competitors: []Competitor{
			{
				Name: "France",
			},
			{
				Name: "Germany",
			},
		},
	}

	testCases := []struct {
		description string
		name        string
		specifiers  map[string]string
		expected    string
	}{
		{
			description: "Replace {X} with value of specifier",
			name:        "Race to {pointnr} points",
			specifiers:  map[string]string{"pointnr": "3"},
			expected:    "Race to 3 points",
		},
		{
			description: "Replace {!X} with the ordinal value of specifier X",
			name:        "{!periodnr} period - total",
			specifiers:  map[string]string{"periodnr": "2"},
			expected:    "2nd period - total",
		},
		{
			description: "(-) Replace {X+/-c} with the value of the specifier X + or - the number c",
			name:        "Score difference {pointnr-3}",
			specifiers:  map[string]string{"pointnr": "10"},
			expected:    "Score difference 7",
		},
		{
			description: "(+) Replace {X+/-c} with the value of the specifier X + or - the number c",
			name:        "Score difference {pointnr+3}",
			specifiers:  map[string]string{"pointnr": "10"},
			expected:    "Score difference 13",
		},
		{
			description: "(-) Replace {!X+c} with the ordinal value of specifier X + c",
			name:        "{!inningnr-1} inning",
			specifiers:  map[string]string{"inningnr": "2"},
			expected:    "1st inning",
		},
		{
			description: "Replace {!X+c} with the ordinal value of specifier X + c",
			name:        "{!inningnr+1} inning",
			specifiers:  map[string]string{"inningnr": "2"},
			expected:    "3rd inning",
		},
		{
			description: "Replace {+X} with the value of specifier X with a +/- sign in front",
			name:        "Goal Diff {+goals} goals",
			specifiers:  map[string]string{"goals": "2"},
			expected:    "Goal Diff +2 goals",
		},
		{
			description: "Replace {-X} with the negated value of the specifier with a +/- sign in front",
			name:        "Goal Diff {-goals} goals",
			specifiers:  map[string]string{"goals": "2"},
			expected:    "Goal Diff -2 goals",
		},
		{
			description: "Name with 2 normal specifiers",
			name:        "Holes {from} to {to} - head2head (1x2) groups",
			specifiers:  map[string]string{"from": "1", "to": "18"},
			expected:    "Holes 1 to 18 - head2head (1x2) groups",
		},
		{
			description: "Name with 1 normal, 1 ordinal specifier",
			name:        "{!periodnr} period - {pointnr+3} points",
			specifiers:  map[string]string{"periodnr": "2", "pointnr": "10"},
			expected:    "2nd period - 13 points",
		},
		{
			description: "Name with 1 ordinal, 1 +/- specifier",
			name:        "{!half} half, {+goals} goals",
			specifiers:  map[string]string{"half": "1", "goals": "2"},
			expected:    "1st half, +2 goals",
		},
		{
			description: "Replace {%player} with name of specifier",
			name:        "{%player} total dismissals",
			specifiers:  map[string]string{"player": "1234"},
			expected:    "John Rodriquez total dismissals",
		},
		{
			description: "Player with 1 normal, 1 ordinal specifier",
			name:        "{!half} half - {%player} {goals} goals",
			specifiers:  map[string]string{"half": "1", "player": "1234", "goals": "2"},
			expected:    "1st half - John Rodriquez 2 goals",
		},
		{
			description: "Market ID 368",
			name:        "{!inningnr} innings - {%player} total",
			specifiers:  map[string]string{"inningnr": "1", "maxovers": "20", "player": "1234", "playernr": "2", "total": "27.5"},
			expected:    "1st innings - John Rodriquez total",
		},
		{
			description: "Replace {$event} with the name of the event",
			name:        "Winner of {$event}",
			specifiers:  nil,
			expected:    "Winner of Euro2016",
		},
		{
			description: "Replace {$competitorN} with the Nth competitor in the event (empty map)",
			name:        "Winner is {$competitor2}",
			specifiers:  nil,
			expected:    "Winner is Germany",
		},
		{
			description: "Replace {X} with value of specifier for decimals",
			name:        "{!periodnr} period - total {pointnr} points",
			specifiers:  map[string]string{"periodnr": "3", "pointnr": "3.5"},
			expected:    "3rd period - total 3.5 points",
		},
		{
			description: "2 Competitor with corners",
			name:        "{$competitor1}, {$competitor2} exact corners {cornernr}",
			specifiers:  map[string]string{"cornernr": "2"},
			expected:    "France, Germany exact corners 2",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			actual, err := ParseSpecifiers(tc.name, tc.specifiers, players, fixture)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
