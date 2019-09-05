package uof

import (
	"hash/fnv"
	"strconv"
	"strings"
)

type Producer int8

var producers = []struct {
	id          Producer
	name        string
	description string
}{
	{id: 1, name: "LO", description: "Live Odds"},
	{id: 3, name: "Ctrl", description: "Betradar Ctrl"},
	{id: 4, name: "BetPal", description: "BetPal"},
	{id: 5, name: "PremiumCricket", description: "Premium Cricket"},
	{id: 6, name: "VF", description: "Virtual football"},
	{id: 7, name: "WNS", description: "World Number Service"},
	{id: 8, name: "VBL", description: "Virtual Basketball League"},
	{id: 9, name: "VTO", description: "Virtual Tennis Open"},
}

func (p Producer) String() string {
	return p.Name()
}

func (p Producer) Name() string {
	for _, d := range producers {
		if p == d.id {
			return d.name
		}
	}
	return InvalidName
}

func (p Producer) Description() string {
	for _, d := range producers {
		if p == d.id {
			return d.description
		}
	}
	return InvalidName
}

// Prematch means that producer markets are valid only for betting before the
// match starts.
func (p Producer) Prematch() bool {
	return p == 3
}

const (
	InvalidName = "?"
	srPlayer    = "sr:player:"
)

type URN string

func (u URN) ID() int {
	if u == "" {
		return 0
	}
	p := strings.Split(string(u), ":")
	if len(p) != 3 {
		return 0
	}
	i, _ := strconv.ParseUint(p[2], 10, 64)
	return int(i)
}

func (u URN) Type() int8 {
	if u == "" {
		return URNTypeUnknown
	}
	p := strings.Split(string(u), ":")
	if len(p) != 3 {
		return URNTypeUnknown
	}
	switch p[1] {
	case "match":
		return URNTypeMatch
	case "stage":
		return URNTypeStage
	case "tournament":
		return URNTypeTournament
	case "simple_tournament":
		return URNTypeSimpleTournament
	case "season":
		return URNTypeSeason
	case "draw":
		return URNTypeDraw
	case "lottery":
		return URNTypeLottery
	case "player":
		return URNTypePlayer
	}
	return URNTypeUnknown
}

const (
	URNTypeMatch int8 = iota
	URNTypeStage
	URNTypeTournament
	URNTypeSimpleTournament
	URNTypeSeason
	URNTypeDraw
	URNTypeLottery
	URNTypePlayer

	URNTypeUnknown = int8(-1)
)

func toLineID(specifiers string) int {
	if specifiers == "" {
		return 0
	}
	return hash32(specifiers)
}

func hash32(s string) int {
	h := fnv.New32a()
	h.Write([]byte(s))
	return int(h.Sum32())
}

// The default value is active if status is not present.
type MarketStatus int8

// Reference: https://docs.betradar.com/display/BD/UOF+-+Market+status
const (
	// Active/suspended/inactive could be sent in odds change message:

	// Odds are provided and you can accept bets on the market.
	MarketStatusActive MarketStatus = 1
	// Odds continue to be provided but you should not accept bets on the market
	// for a short time (e.g. from right before a goal and until the goal has been
	// observed/confirmed).
	MarketStatusSuspended MarketStatus = -1
	// Odds are no longer provided for this market. A market can go back to Active
	// again i.e.: A total 3.5 market is deactivated since 0.5, 1.5 or 2.5 is the
	// most balanced market. However, if a goal is scored, then the 3.5 market
	// becomes the most balanced again, changing status to active. There are
	// numerous other reasons for this change as well, and it happens on a regular
	// basis.
	MarketStatusInactive MarketStatus = 0

	// During recovery the following additional status may also be sent:

	// Not a real market status. This status is normally seen under recovery, and
	// is a signal that the producer that sends this message is no longer sending
	// odds for this market. Odds will come from another producer going forward
	// (and might already have started coming from the new producer). Handed over
	// is also sent by the prematch producer when the Live Odds producer takes
	// over a market. If you have not received the live odds change yet, the
	// market should be suspended, otherwise the message can be ignored. If the
	// live odds change does not eventually appear, the market should likely be
	// deactivated.
	MarketStatusHandedOver MarketStatus = -2
	// Bet Settlement messages have been sent for this market, no further odds
	// will be provided. However, it should be noted that in rare cases (error
	// conditions), a settled market may be moved to cancelled by a bet_cancel
	// message.
	MarketStatusSettled MarketStatus = -3
	// This market has been cancelled. No further odds will be provided for this
	// market. This state is only seen during recovery for matches where the
	// system has sent out a cancellation message for that particular market.
	MarketStatusCancelled MarketStatus = -4
)

type CashoutStatus int8

const (
	// available for cashout
	CashoutStatusAvailable CashoutStatus = 1
	// temporarily unavailable for cashout
	CashoutStatusUnavailable CashoutStatus = -1
	// permanently unavailable for cashout
	CashoutStatusClosed CashoutStatus = -2
)

type Team int8

const (
	TeamHome Team = 1
	TeamAway Team = 2
)
