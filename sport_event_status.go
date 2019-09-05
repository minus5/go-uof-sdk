package uof

// The element "sport_event_status" is provided in the odds_change message.
// Status is the only required attribute for this element, and this attribute
// describes the current status of the sport-event itself (not started, live,
// ended, closed). Additional attributes are live-only attributes, and only
// provided while the match is live; additionally, which attributes are provided
// depends on the sport.
// Reference: https://docs.betradar.com/display/BD/UOF+-+Sport+event+status
type SportEventStatus struct {
	// High-level generic status of the match.
	Status EventStatus `xml:"status,attr" json:"status"`
	// Does Betradar have a scout watching the game.
	Reporting *EventReporting `xml:"reporting,attr,omitempty" json:"reporting,omitempty"`
	// Current score for the home team.
	HomeScore *int `xml:"home_score,attr,omitempty" json:"homeScore,omitempty"`
	// Current score for the away team.
	AwayScore *int `xml:"away_score,attr,omitempty" json:"awayScore,omitempty"`
	// Sports-specific integer code the represents the live match status (first period, 2nd break, etc.).
	MatchStatus *int `xml:"match_status,attr" json:"matchStatus"`
	// The player who has the serve at that moment.
	CurrentServer *Team `xml:"current_server,attr,omitempty" json:"currentServer,omitempty"`
	// The point score of the "home" player. The score will be 50 if the "home"
	// player has advantage. This attribute is also used for the tiebreak score
	// when the game is in a tiebreak.
	// (15 30 40 50)
	HomeGamescore *int `xml:"home_gamescore,attr,omitempty" json:"homeGamescore,omitempty"`
	// The point score of the "away" player. The score will be 50 if the "away"
	// player has advantage. This attribute is also used for the tiebreak score
	// when the game is in a tiebreak.
	AwayGamescore *int `xml:"away_gamescore,attr,omitempty" json:"awayGamescore,omitempty"`

	HomePenaltyScore   *int    `xml:"home_penalty_score,attr,omitempty" json:"homePenaltyScore,omitempty"`
	AwayPenaltyScore   *int    `xml:"away_penalty_score,attr,omitempty" json:"awayPenaltyScore,omitempty"`
	HomeLegscore       *int    `xml:"home_legscore,attr,omitempty" json:"homeLegscore,omitempty"`
	AwayLegscore       *int    `xml:"away_legscore,attr,omitempty" json:"awayLegscore,omitempty"`
	ExpediteMode       *bool   `xml:"expedite_mode,attr,omitempty" json:"expediteMode,omitempty"`
	Tiebreak           *bool   `xml:"tiebreak,attr,omitempty" json:"tiebreak,omitempty"`
	HomeSuspend        *int    `xml:"home_suspend,attr,omitempty" json:"homeSuspend,omitempty"`
	AwaySuspend        *int    `xml:"away_suspend,attr,omitempty" json:"awaySuspend,omitempty"`
	Balls              *int    `xml:"balls,attr,omitempty" json:"balls,omitempty"`
	Strikes            *int    `xml:"strikes,attr,omitempty" json:"strikes,omitempty"`
	Outs               *int    `xml:"outs,attr,omitempty" json:"outs,omitempty"`
	Bases              *string `xml:"bases,attr,omitempty" json:"bases,omitempty"`
	HomeBatter         *int    `xml:"home_batter,attr,omitempty" json:"homeBatter,omitempty"`
	AwayBatter         *int    `xml:"away_batter,attr,omitempty" json:"awayBatter,omitempty"`
	Possession         *int    `xml:"possession,attr,omitempty" json:"possession,omitempty"`
	Position           *int    `xml:"position,attr,omitempty" json:"position,omitempty"`
	Try                *int    `xml:"try,attr,omitempty" json:"try,omitempty"`
	Yards              *int    `xml:"yards,attr,omitempty" json:"yards,omitempty"`
	Throw              *int    `xml:"throw,attr,omitempty" json:"throw,omitempty"`
	Visit              *int    `xml:"visit,attr,omitempty" json:"visit,omitempty"`
	RemainingReds      *int    `xml:"remaining_reds,attr,omitempty" json:"remainingReds,omitempty"`
	Delivery           *int    `xml:"delivery,attr,omitempty" json:"delivery,omitempty"`
	HomeRemainingBowls *int    `xml:"home_remaining_bowls,attr,omitempty" json:"homeRemainingBowls,omitempty"`
	AwayRemainingBowls *int    `xml:"away_remaining_bowls,attr,omitempty" json:"awayRemainingBowls,omitempty"`
	CurrentEnd         *int    `xml:"current_end,attr,omitempty" json:"currentEnd,omitempty"`
	Innings            *int    `xml:"innings,attr,omitempty" json:"innings,omitempty"`
	Over               *int    `xml:"over,attr,omitempty" json:"over,omitempty"`
	HomePenaltyRuns    *int    `xml:"home_penalty_runs,attr,omitempty" json:"homePenaltyRuns,omitempty"`
	AwayPenaltyRuns    *int    `xml:"away_penalty_runs,attr,omitempty" json:"awayPenaltyRuns,omitempty"`
	HomeDismissals     *int    `xml:"home_dismissals,attr,omitempty" json:"homeDismissals,omitempty"`
	AwayDismissals     *int    `xml:"away_dismissals,attr,omitempty" json:"awayDismissals,omitempty"`
	CurrentCtTeam      *Team   `xml:"current_ct_team,attr,omitempty" json:"currentCtTeam,omitempty"`

	Clock        *Clock        `xml:"clock,omitempty" json:"clock,omitempty"`
	PeriodScores []PeriodScore `xml:"period_scores>period_score,omitempty" json:"periodScores,omitempty"`
	Results      []Result      `xml:"results>result,omitempty" json:"results,omitempty"`
	Statistics   *Statistics   `xml:"statistics,omitempty" json:"statistics,omitempty"`
}

// The sport_event_status may contain a clock element. This clock element
// includes various clock/time attributes that are sports specific.
type Clock struct {
	// The playing minute of the match (or minute:second if available)
	// mm:ss (42:10)
	MatchTime *ClockTime `xml:"match_time,attr,omitempty" json:"matchTime,omitempty"`
	// How far into stoppage time is the match in minutes
	// mm:ss
	StoppageTime *ClockTime `xml:"stoppage_time,attr,omitempty" json:"stoppageTime,omitempty"`
	// Set to what the announce stoppage time is announced
	// mm:ss
	StoppageTimeAnnounced *ClockTime `xml:"stoppage_time_announced,attr,omitempty" json:"stoppageTimeAnnounced,omitempty"`
	// How many minutes remains of the match
	// mm:ss
	RemainingTime *ClockTime `xml:"remaining_time,attr,omitempty" json:"remainingTime,omitempty"`
	// How much time remains in the current period
	// mm:ss
	RemainingTimeInPeriod *ClockTime `xml:"remaining_time_in_period,attr,omitempty" json:"remainingTimeInPeriod,omitempty"`
	// true if the match clock is stopped otherwise false
	Stopped *bool `xml:"stopped,attr,omitempty" json:"stopped,omitempty"`
}

type PeriodScore struct {
	// The match status of an event gives an indication of which context the
	// current match is in. Complete list available at:
	// /v1/descriptions/en/match_status.xml
	MatchStatusCode *int `xml:"match_status_code,attr" json:"matchStatusCode"`
	// Indicates what regular period this is
	Number *int `xml:"number,attr" json:"number"`
	// The number of points/goals/games the competitor designated as "home" has
	// scored for this period.
	HomeScore *int `xml:"home_score,attr" json:"homeScore"`
	// The number of points/goals/games the competitor designated as "away" has
	// scored for this period.
	AwayScore *int `xml:"away_score,attr" json:"awayScore"`
}

type Result struct {
	MatchStatusCode *int `xml:"match_status_code,attr" json:"matchStatusCode"`
	HomeScore       *int `xml:"home_score,attr" json:"homeScore"`
	AwayScore       *int `xml:"away_score,attr" json:"awayScore"`
}

type Statistics struct {
	YellowCards    *StatisticsScore `xml:"yellow_cards,omitempty" json:"yellowCards,omitempty"`
	RedCards       *StatisticsScore `xml:"red_cards,omitempty" json:"redCards,omitempty"`
	YellowRedCards *StatisticsScore `xml:"yellow_red_cards,omitempty" json:"yellowRedCards,omitempty"`
	Corners        *StatisticsScore `xml:"corners,omitempty" json:"corners,omitempty"`
}

type StatisticsScore struct {
	Home int `xml:"home,attr" json:"home"`
	Away int `xml:"away,attr" json:"away"`
}
