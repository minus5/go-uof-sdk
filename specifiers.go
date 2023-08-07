package uof

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
)

func ParseSpecifiers(name string, specifiers map[string]string, players map[int]Player, fixture Fixture) (string, error) {
	for key, val := range specifiers {
		switch {
		case strings.Contains(name, "{"+key+"}"):
			name = strings.ReplaceAll(name, "{"+key+"}", val)
		case strings.Contains(name, "{!"+key+"}"):
			intVal, err := strconv.Atoi(val)
			if err != nil {
				return "", fmt.Errorf("invalid number in specifier with ordinal operator: %s", val)
			}
			name = strings.ReplaceAll(name, "{!"+key+"}", humanize.Ordinal(intVal))
		case strings.Contains(name, "{"+key+"-"):
			i := strings.Index(name, "{"+key+"-")
			j := strings.Index(name[i:], "}") + i
			nStr := name[i+len(key)+2 : j]
			n, err := strconv.Atoi(nStr)
			if err != nil {
				return "", fmt.Errorf("invalid number in name with sub(-) specifier: %s", name)
			}
			intVal, err := strconv.Atoi(val)
			if err != nil {
				return "", fmt.Errorf("invalid number in specifier with sub(-) operator: %s", val)
			}
			result := intVal - n
			name = name[:i] + strconv.Itoa(result) + name[j+1:]
		case strings.Contains(name, "{"+key+"+"):
			i := strings.Index(name, "{"+key+"+")
			j := strings.Index(name[i:], "}") + i
			nStr := name[i+len(key)+2 : j]
			n, err := strconv.Atoi(nStr)
			if err != nil {
				return "", fmt.Errorf("invalid number in name with sum(+) specifier: %s", name)
			}
			intVal, err := strconv.Atoi(val)
			if err != nil {
				return "", fmt.Errorf("invalid number in specifier with sum(+) operator: %s", val)
			}
			result := intVal + n
			name = name[:i] + strconv.Itoa(result) + name[j+1:]
		case strings.Contains(name, "{!"+key+"-"):
			i := strings.Index(name, "{!"+key+"-")
			j := strings.Index(name[i:], "}") + i
			nStr := name[i+len(key)+3 : j]
			n, err := strconv.Atoi(nStr)
			if err != nil {
				return "", fmt.Errorf("invalid number in name with sub(-) ordinal specifier: %s", name)
			}
			intVal, err := strconv.Atoi(val)
			if err != nil {
				return "", fmt.Errorf("invalid number in specifier with sub(-) ordinal operator: %s", val)
			}
			result := intVal - n
			name = name[:i] + humanize.Ordinal(result) + name[j+1:]
		case strings.Contains(name, "{!"+key+"+"):
			i := strings.Index(name, "{!"+key+"+")
			j := strings.Index(name[i:], "}") + i
			nStr := name[i+len(key)+3 : j]
			n, err := strconv.Atoi(nStr)
			if err != nil {
				return "", fmt.Errorf("invalid number in name with sum(+) ordinal specifier: %s", name)
			}
			intVal, err := strconv.Atoi(val)
			if err != nil {
				return "", fmt.Errorf("invalid number in specifier with sum(+) ordinal operator: %s", val)
			}
			result := intVal + n
			name = name[:i] + humanize.Ordinal(result) + name[j+1:]
		case strings.Contains(name, "{+"+key+"}"):
			i := strings.Index(name, "{")
			j := strings.Index(name[i:], "}") + i
			intVal, err := strconv.Atoi(val)
			if err != nil {
				return "", fmt.Errorf("invalid number in specifier with signed operator: %s", val)
			}
			name = name[:i] + "+" + strconv.Itoa(intVal) + name[j+1:]
		case strings.Contains(name, "{-"+key+"}"):
			i := strings.Index(name, "{")
			j := strings.Index(name[i:], "}") + i
			intVal, err := strconv.Atoi(val)
			if err != nil {
				return "", fmt.Errorf("invalid number in specifier with signed operator: %s", val)
			}
			name = name[:i] + "-" + strconv.Itoa(intVal) + name[j+1:]
		case strings.Contains(name, "{%player}"):
			playerID := URN(val).ID()
			player, ok := players[playerID]
			if !ok {
				return "", fmt.Errorf("player with id %d not found", playerID)
			}
			name = strings.ReplaceAll(name, "{%player}", player.FullName)
		case strings.Contains(name, "{$event}"):
			name = strings.ReplaceAll(name, "{$event}", fixture.Name)
		case strings.Contains(name, "{$competitor"):
			i := strings.Index(name, "{$competitor")
			j := strings.Index(name[i:], "}") + i
			index := name[i+len("{$competitor") : j]
			indexInt, err := strconv.Atoi(index)
			if err != nil {
				return "", fmt.Errorf("invalid number in specifier with competitor operator: %s", index)
			}
			if indexInt > len(fixture.Competitors) {
				return "", fmt.Errorf("invalid number in specifier with competitor operator: %s", index)
			}
			competitor := fixture.Competitors[indexInt-1]
			name = name[:i] + competitor.Name + name[j+1:]
		}
	}
	return name, nil
}