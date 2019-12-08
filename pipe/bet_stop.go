package pipe

import (
	"encoding/json"
	"math"
	"sort"

	"github.com/minus5/go-uof-sdk"
)

func marketGroups() map[string][]int {
	var marketGroups map[string][]int
	err := json.Unmarshal([]byte(marketGroupsJSON), &marketGroups)
	if err != nil {
		panic(err)
	}
	return marketGroups
}

type betStop struct {
	marketGroups map[string][]int
}

// BetStop enriches bet stop messages with the list of the marketIDs which
// should be stopped. In initial message we got list of market groups. In other
// event messages we have only market ids. To allow client not to need to know
// the list of all markets to make connection between groups and ids we are here
// adding to the bet stop message those ids.
func BetStop() InnerStage {
	b := betStop{
		marketGroups: marketGroups(),
	}
	return Stage(b.loop)
}

func (b *betStop) loop(in <-chan *uof.Message, out chan<- *uof.Message, errc chan<- error) {
	for m := range in {
		switch m.Type {
		case uof.MessageTypeBetStop:
			b.enrich(m)
		case uof.MessageTypeMarkets:
			b.refresh(m)
		}
		out <- m
	}
}

func (b *betStop) refresh(m *uof.Message) {
	if m.Lang != uof.LangEN || m.Markets == nil {
		return
	}
	b.marketGroups = m.Markets.Groups()
}

func (b *betStop) enrich(m *uof.Message) {
	bs := m.BetStop
	if bs == nil || bs.Groups == nil {
		return
	}
	var marketIDs []int
	for _, k := range bs.Groups {
		if ids, ok := b.marketGroups[k]; ok {
			marketIDs = append(marketIDs, ids...)
		}
	}
	bs.MarketIDs = dedup(marketIDs)
}

func dedup(in []int) []int {
	sort.Ints(in)
	out := make([]int, 0, len(in))
	l := int(math.MinInt64)
	for _, c := range in {
		if c != l {
			out = append(out, c)
			l = c
		}
	}
	return out
}
