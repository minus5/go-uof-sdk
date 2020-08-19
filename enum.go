package uof

import (
	"fmt"
	"hash/fnv"
	"strconv"
	"strings"
)

type Producer int8

const (
	ProducerUnknown  Producer = -1
	ProducerDefault  Producer = 0
	ProducerLiveOdds Producer = 1
	ProducerPrematch Producer = 3
)

var producers = []struct {
	id             Producer
	name           string
	description    string
	code           string
	scope          string
	recoveryWindow int // in minutes
}{
	{id: 0, name: "SR", description: "Sports", code: "sr"},
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

func (p Producer) Scope() string {
	for _, d := range producers {
		if p == d.id {
			return d.scope
		}
	}
	return InvalidName
}

// RecoveryWindow in milliseconds
func (p Producer) RecoveryWindow() int {
	for _, d := range producers {
		if p == d.id {
			return d.recoveryWindow * 60 * 1000
		}
	}
	return 0
}

// Prematch means that producer markets are valid only for betting before the
// match starts.
func (p Producer) Prematch() bool {
	return p == 3
}

func (p Producer) Sports() bool {
	return p == ProducerLiveOdds || p == ProducerPrematch
}

func (p Producer) Virtuals() bool {
	return p.Scope() == "virtual"
}

func VirtualProducers() []Producer {
	var v []Producer
	for _, d := range producers {
		p := d.id
		if p.Virtuals() {
			v = append(v, p)
		}
	}
	return v
}

const (
	InvalidName  = "?"
	srMatch      = "sr:match:"
	srPlayer     = "sr:player:"
	srCompetitor = "sr:competitor:"
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

func (u URN) Empty() bool {
	return string(u) == ""
}

func NewEventURN(eventID int) URN {
	return URN(fmt.Sprintf("%s%d", srMatch, eventID))
}

func (u URN) String() string {
	return string(u)
}

func (u *URN) Parse(s string) {
	r := URN(s)
	if id, err := strconv.Atoi(s); err == nil {
		r = NewEventURN(id)
	}
	*u = r
}

const NoURN = URN("")

// EventID tries to generate unique id for all types of events. Most comon are
// those with prefix sr:match for them we reserve positive id-s. All others got
// range in negative ids.
// Reference: https://docs.betradar.com/display/BD/MG+-+Entities
//            http://sdk.sportradar.com/content/unifiedfeedsdk/net/doc/html/e1f73019-73cd-c9f8-0d58-7fe25800abf2.htm
// List of currently existing event types is taken from the combo box in the
// integration control page. From method "Fixture for a specified sport event".
// !!! Refactored to Producer + EventID in producer.
//nolint:gocyclo //accepting complexity of 23
func (u URN) EventID() int {
	id, prefix := u.split()
	if id == 0 {
		return 0
	}
	if u.Producer() != ProducerUnknown {
		return id
	}

	suffixID := func(suffix int8) int {
		return -(id<<8 | int(suffix))
	}

	switch prefix {
	case "sr:match":
		return id
	case "sr:stage":
		return suffixID(1)
	case "sr:season":
		return suffixID(2)
	case "sr:tournament":
		return suffixID(3)
	case "sr:simple_tournament":
		return suffixID(4)
	case "test:match":
		return suffixID(15)
	case "vf:match":
		return suffixID(16)
	case "vf:season":
		return suffixID(17)
	case "vf:tournament":
		return suffixID(18)
	case "vbl:match":
		return suffixID(19)
	case "vbl:season":
		return suffixID(20)
	case "vbl:tournament":
		return suffixID(21)
	case "vto:match":
		return suffixID(22)
	case "vto:season":
		return suffixID(23)
	case "vto:tournament":
		return suffixID(24)
	case "vdr:stage":
		return suffixID(25)
	case "vhc:stage":
		return suffixID(26)
	case "vti:match":
		return suffixID(27)
	case "vti:tournament":
		return suffixID(28)
	case "wns:draw":
		return suffixID(29)
	default:
		return 0
	}
}

func (u URN) Producer() Producer {
	p := strings.Split(string(u), ":")
	if len(p) != 3 {
		return ProducerUnknown
	}
	for _, d := range producers {
		if d.code == p[0] {
			return Producer(d.id)
		}
	}
	return ProducerUnknown
}

// splits urn into id and prefix
func (u URN) split() (int, string) {
	if u == "" {
		return 0, ""
	}
	p := strings.Split(string(u), ":")
	if len(p) != 3 {
		return 0, ""
	}
	i, err := strconv.ParseUint(p[2], 10, 64)
	if err != nil {
		return 0, ""
	}
	id := int(i)
	prefix := p[0] + ":" + p[1]

	return id, prefix
}

func (u URN) IsTournament() bool {
	p := strings.Split(string(u), ":")
	if len(p) != 3 {
		return false
	}
	return p[1] == "season" || p[1] == "tournament"
}

func toLineID(specifiers string) int {
	if specifiers == "" {
		return 0
	}
	return hash32(specifiers)
}

func hash32(s string) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return int(h.Sum32())
}

func Hash(s string) int {
	return hash32(s)
}

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

func (c *ClockTime) PtrVal() *string {
	if c == nil {
		return nil
	}
	v := string(*c)
	return &v
}

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

type MessageType int8

const (
	MessageTypeUnknown MessageType = -1
)

// event related message types
const (
	MessageTypeOddsChange MessageType = iota
	MessageTypeFixtureChange
	MessageTypeBetCancel
	MessageTypeBetSettlement
	MessageTypeBetStop
	MessageTypeRollbackBetSettlement
	MessageTypeRollbackBetCancel
)

// api message types
const (
	MessageTypeFixture MessageType = iota + 32
	MessageTypeMarkets
	MessageTypePlayer
	MessageTypeCompetitor
	MessageTypeTournament
	MessageTypeSummary
)

// system message types
const (
	MessageTypeAlive MessageType = iota + 64
	MessageTypeSnapshotComplete
	MessageTypeConnection
	MessageTypeProducersChange
)

var messageTypes = []MessageType{
	MessageTypeUnknown,

	MessageTypeOddsChange,
	MessageTypeFixtureChange,
	MessageTypeBetCancel,
	MessageTypeBetSettlement,
	MessageTypeBetStop,
	MessageTypeRollbackBetSettlement,
	MessageTypeRollbackBetCancel,

	MessageTypeFixture,
	MessageTypeMarkets,
	MessageTypePlayer,
	MessageTypeCompetitor,
	MessageTypeTournament,

	MessageTypeAlive,
	MessageTypeSnapshotComplete,
	MessageTypeConnection,
	MessageTypeProducersChange,
}

var messageTypeNames = []string{
	InvalidName,

	"odds_change",
	"fixture_change",
	"bet_cancel",
	"bet_settlement",
	"bet_stop",
	"rollback_bet_settlement",
	"rollback_bet_cancel",

	"fixture",
	"market",
	"player",
	"competitor",
	"tournament",

	"alive",
	"snapshot_complete",
	"connection",
	"producer_change",
}

func (m *MessageType) Parse(name string) {
	v := MessageTypeUnknown
	for i, n := range messageTypeNames {
		if n == name {
			v = messageTypes[i]
			break
		}
	}
	*m = v
}

func (m MessageType) String() string {
	for i, t := range messageTypes {
		if t == m {
			return messageTypeNames[i]
		}
	}
	return InvalidName
}

func (m MessageType) Kind() MessageKind {
	if m < 32 {
		return MessageKindEvent
	}
	if m < 64 {
		return MessageKindLexicon
	}
	return MessageKindSystem
}

type MessageKind int8

const (
	MessageKindEvent MessageKind = iota
	MessageKindLexicon
	MessageKindSystem
)

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

type Environment int8

const (
	Production Environment = iota
	Staging
	Replay
)
