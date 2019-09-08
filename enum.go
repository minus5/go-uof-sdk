package uof

import (
	"fmt"
	"hash/fnv"
	"strconv"
	"strings"
)

type Producer int8

const (
	ProducerLiveOdds Producer = 1
	ProducerPrematch Producer = 3
)

var producers = []struct {
	id             Producer
	name           string
	description    string
	code           string
	scope          string
	recoveryWindow int
}{
	{id: 1, name: "LO", description: "Live Odds", code: "liveodds", scope: "live", recoveryWindow: 4320},
	{id: 3, name: "Ctrl", description: "Betradar Ctrl", code: "pre", scope: "prematch", recoveryWindow: 4320},
	{id: 4, name: "BetPal", description: "BetPal", code: "betpal", scope: "live", recoveryWindow: 4320},
	{id: 5, name: "PremiumCricket", description: "Premium Cricket", code: "premium_cricket", scope: "live|prematch", recoveryWindow: 4320},
	{id: 6, name: "VF", description: "Virtual football", code: "vf", scope: "virtual", recoveryWindow: 180},
	{id: 7, name: "WNS", description: "Numbers Betting", code: "wns", scope: "prematch", recoveryWindow: 4320},
	{id: 8, name: "VBL", description: "Virtual Basketball League", code: "vbl", scope: "virtual", recoveryWindow: 180},
	{id: 9, name: "VTO", description: "Virtual Tennis Open", code: "vto", scope: "virtual", recoveryWindow: 180},
	{id: 10, name: "VDR", description: "Virtual Dog Racing", code: "vdr", scope: "virtual", recoveryWindow: 180},
	{id: 11, name: "VHC", description: "Virtual Horse Classics", code: "vhc", scope: "virtual", recoveryWindow: 180},
	{id: 12, name: "VTI", description: "Virtual Tennis In-Play", code: "vti", scope: "virtual", recoveryWindow: 180},
	{id: 15, name: "VBI", description: "Virtual Baseball In-Play", code: "vbi", scope: "virtual", recoveryWindow: 180},
}

func (p Producer) String() string {
	return p.Code()
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

func (p Producer) Code() string {
	for _, d := range producers {
		if p == d.id {
			return d.code
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
	srMatch     = "sr:match:"
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

func NewEventURN(eventID int) URN {
	return URN(fmt.Sprintf("%s%d", srMatch, eventID))
}

func (u URN) String() string {
	return string(u)
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

type EventStatus int8

const (
	EventStatusNotStarted EventStatus = 0
	EventStatusLive       EventStatus = 1
	EventStatusSuspended  EventStatus = 2 // Used by the Premium Cricket odds producer
	EventStatusEnded      EventStatus = 3
	EventStatusClosed     EventStatus = 4
	// Only one of the above statuses are possible in the odds_change message in
	// the feed. However please note that other states are available in the API,
	// but will not appear in the odds_change message. These are as following:
	EventStatusCancelled   EventStatus = 5
	EventStatusDelayed     EventStatus = 6
	EventStatusInterrupted EventStatus = 7
	EventStatusPostponed   EventStatus = 8
	EventStatusAbandoned   EventStatus = 9
)

type EventReporting int8

const (
	EventReportingNotAvailable EventReporting = 0
	EventReportingActive       EventReporting = 1
	EventReportingSuspended    EventReporting = -1
)

// Values must match the pattern [0-9]+:[0-9]+|[0-9]+
type ClockTime string

func (c *ClockTime) Minute() string {
	p := strings.Split(string(*c), ":")
	if len(p) > 0 {
		return p[0]
	}
	return ""
}

func (c *ClockTime) String() string {
	return string(*c)
}

type OutcomeResult int8

const (
	OutcomeResultUnknown  OutcomeResult = 0
	OutcomeResultLose     OutcomeResult = 1
	OutcomeResultWin      OutcomeResult = 2
	OutcomeResultVoid     OutcomeResult = 3
	OutcomeResultHalfLose OutcomeResult = 4
	OutcomeResultHalfWin  OutcomeResult = 5
)

// The change_type attribute (if present), describes what type of change that
// caused the message to be sent. In general, best practices are to always
// re-fetch the updated fixture from the API and not solely rely on the
// change_type and the message content. This is because multiple different
// changes could have been made.
// May be one of 1, 2, 3, 4, 5
type FixtureChangeType int8

const (
	// This is a new match/event that has been just added.
	FixtureChangeTypeNew FixtureChangeType = 1
	// Start-time update
	FixtureChangeTypeTime FixtureChangeType = 2
	// This sport event will not take place. It has been cancelled.
	FixtureChangeTypeCancelled FixtureChangeType = 3
	// The format of the sport-event has been updated (e.g. the number of sets to
	// play has been updated or the length of the match etc.)
	FixtureChangeTypeFromat FixtureChangeType = 4
	// Coverage update. Sent for example when liveodds coverage for some reason
	// cannot be offered for a match.
	FixtureChangeTypeCoverage FixtureChangeType = 5
)

type OutcomeType int8

const (
	OutcomeTypeDefault OutcomeType = iota
	OutcomeTypePlayer
	OutcomeTypeCompetitor
	OutcomeTypeCompetitors
	OutcomeTypeFreeText
	OutcomeTypeUnknown OutcomeType = -1
)

type SpecifierType int8

const (
	SpecifierTypeString SpecifierType = iota
	SpecifierTypeInteger
	SpecifierTypeDecimal
	SpecifierTypeVariableText
	SpecifierTypeUnknown SpecifierType = -1
)

type Gender int8

const (
	GenderUnknown Gender = iota
	Male
	Female
)

type MessageType int8

// amqp message types
const (
	MessageTypeUnknown    MessageType = -1
	MessageTypeOddsChange MessageType = iota
	MessageTypeFixtureChange
	MessageTypeBetCancel
	MessageTypeBetSettlement
	MessageTypeBetStop
	MessageTypeRollbackBetSettlement
	MessageTypeRollbackBetCancel
	MessageTypeSnapshotComplete
	MessageTypeAlive
)

// api message types
const (
	MessageTypeFixture MessageType = iota + 32
	MessageTypeMarkets
	MessageTypePlayer
)

const (
	MessageTypeConnection MessageType = 127
)

func (m *MessageType) Parse(name string) {
	v := func() MessageType {
		switch name {
		case "alive":
			return MessageTypeAlive
		case "bet_cancel":
			return MessageTypeBetCancel
		case "bet_settlement":
			return MessageTypeBetSettlement
		case "bet_stop":
			return MessageTypeBetStop
		case "fixture_change":
			return MessageTypeFixtureChange
		case "odds_change":
			return MessageTypeOddsChange
		case "rollback_bet_settlement":
			return MessageTypeRollbackBetSettlement
		case "rollback_bet_cancel":
			return MessageTypeRollbackBetCancel
		case "snapshot_complete":
			return MessageTypeSnapshotComplete
		default:
			return MessageTypeUnknown
		}
	}()
	*m = v
}

type MessageScope int8

// Scope of the message
const (
	MessageScopePrematch MessageScope = iota
	MessageScopeLive
	MessageScopePrematchAndLive
	MessageScopeVirtuals
	MessageScopeSystem // system scope messages, like alive, product down
)

func (s *MessageScope) Parse(prematchInterest, liveInterest string) {
	v := func() MessageScope {
		if prematchInterest == "pre" {
			if liveInterest == "live" {
				return MessageScopePrematchAndLive
			}
			return MessageScopePrematch
		}
		if prematchInterest == "virt" {
			return MessageScopeVirtuals
		}
		if liveInterest == "live" {
			return MessageScopeLive
		}
		return MessageScopeSystem
	}()
	*s = v
}

type MessagePriority int8

const (
	MessagePriorityLow MessagePriority = iota
	MessagePriorityHigh
)

func (p *MessagePriority) Parse(priority string) {
	v := func() MessagePriority {
		switch priority {
		case "hi":
			return MessagePriorityHigh
		default:
			return MessagePriorityLow
		}
	}()
	*p = v
}
