package uof

type SportEventStatus struct {
	Clock              *Clock        `xml:"clock,omitempty" json:"clock,omitempty"`
	PeriodScores       []PeriodScore `xml:"period_scores>period_score,omitempty" json:"periodScores,omitempty"`
	Results            []Result      `xml:"results>result,omitempty" json:"results,omitempty"`
	Statistics         *Statistics   `xml:"statistics,omitempty" json:"statistics,omitempty"`
	Status             *uint8        `xml:"status,attr" json:"status"`                           // May be one of 0, 1, 2, 3, 4
	Reporting          *int8         `xml:"reporting,attr,omitempty" json:"reporting,omitempty"` // May be one of 0, 1, -1
	MatchStatus        *int32        `xml:"match_status,attr" json:"matchStatus"`
	HomeScore          *float64      `xml:"home_score,attr,omitempty" json:"homeScore,omitempty"`
	AwayScore          *float64      `xml:"away_score,attr,omitempty" json:"awayScore,omitempty"`
	HomePenaltyScore   *int32        `xml:"home_penalty_score,attr,omitempty" json:"homePenaltyScore,omitempty"`
	AwayPenaltyScore   *int32        `xml:"away_penalty_score,attr,omitempty" json:"awayPenaltyScore,omitempty"`
	HomeGamescore      *int32        `xml:"home_gamescore,attr,omitempty" json:"homeGamescore,omitempty"`
	AwayGamescore      *int32        `xml:"away_gamescore,attr,omitempty" json:"awayGamescore,omitempty"`
	HomeLegscore       *int32        `xml:"home_legscore,attr,omitempty" json:"homeLegscore,omitempty"`
	AwayLegscore       *int32        `xml:"away_legscore,attr,omitempty" json:"awayLegscore,omitempty"`
	CurrentServer      *uint8        `xml:"current_server,attr,omitempty" json:"currentServer,omitempty"` // May be one of 1, 2
	ExpediteMode       *bool         `xml:"expedite_mode,attr,omitempty" json:"expediteMode,omitempty"`
	Tiebreak           *bool         `xml:"tiebreak,attr,omitempty" json:"tiebreak,omitempty"`
	HomeSuspend        *int32        `xml:"home_suspend,attr,omitempty" json:"homeSuspend,omitempty"`
	AwaySuspend        *int32        `xml:"away_suspend,attr,omitempty" json:"awaySuspend,omitempty"`
	Balls              *int32        `xml:"balls,attr,omitempty" json:"balls,omitempty"`
	Strikes            *int32        `xml:"strikes,attr,omitempty" json:"strikes,omitempty"`
	Outs               *int32        `xml:"outs,attr,omitempty" json:"outs,omitempty"`
	Bases              *string       `xml:"bases,attr,omitempty" json:"bases,omitempty"`
	HomeBatter         *int32        `xml:"home_batter,attr,omitempty" json:"homeBatter,omitempty"`
	AwayBatter         *int32        `xml:"away_batter,attr,omitempty" json:"awayBatter,omitempty"`
	Possession         *int32        `xml:"possession,attr,omitempty" json:"possession,omitempty"`
	Position           *int32        `xml:"position,attr,omitempty" json:"position,omitempty"`
	Try                *int32        `xml:"try,attr,omitempty" json:"try,omitempty"`
	Yards              *int32        `xml:"yards,attr,omitempty" json:"yards,omitempty"`
	Throw              *int32        `xml:"throw,attr,omitempty" json:"throw,omitempty"`
	Visit              *int32        `xml:"visit,attr,omitempty" json:"visit,omitempty"`
	RemainingReds      *int32        `xml:"remaining_reds,attr,omitempty" json:"remainingReds,omitempty"`
	Delivery           *int32        `xml:"delivery,attr,omitempty" json:"delivery,omitempty"`
	HomeRemainingBowls *int32        `xml:"home_remaining_bowls,attr,omitempty" json:"homeRemainingBowls,omitempty"`
	AwayRemainingBowls *int32        `xml:"away_remaining_bowls,attr,omitempty" json:"awayRemainingBowls,omitempty"`
	CurrentEnd         *int32        `xml:"current_end,attr,omitempty" json:"currentEnd,omitempty"`
	Innings            *int32        `xml:"innings,attr,omitempty" json:"innings,omitempty"`
	Over               *int32        `xml:"over,attr,omitempty" json:"over,omitempty"`
	HomePenaltyRuns    *int32        `xml:"home_penalty_runs,attr,omitempty" json:"homePenaltyRuns,omitempty"`
	AwayPenaltyRuns    *int32        `xml:"away_penalty_runs,attr,omitempty" json:"awayPenaltyRuns,omitempty"`
	HomeDismissals     *int32        `xml:"home_dismissals,attr,omitempty" json:"homeDismissals,omitempty"`
	AwayDismissals     *int32        `xml:"away_dismissals,attr,omitempty" json:"awayDismissals,omitempty"`
	CurrentCtTeam      *uint8        `xml:"current_ct_team,attr,omitempty" json:"currentCtTeam,omitempty"` // May be one of 1, 2
}

type Clock struct {
	MatchTime             *string `xml:"match_time,attr,omitempty" json:"matchTime,omitempty"`                           // Must match the pattern [0-9]+:[0-9]+|[0-9]+
	StoppageTime          *string `xml:"stoppage_time,attr,omitempty" json:"stoppageTime,omitempty"`                     // Must match the pattern [0-9]+:[0-9]+|[0-9]+
	StoppageTimeAnnounced *string `xml:"stoppage_time_announced,attr,omitempty" json:"stoppageTimeAnnounced,omitempty"`  // Must match the pattern [0-9]+:[0-9]+|[0-9]+
	RemainingTime         *string `xml:"remaining_time,attr,omitempty" json:"remainingTime,omitempty"`                   // Must match the pattern [0-9]+:[0-9]+|[0-9]+
	RemainingTimeInPeriod *string `xml:"remaining_time_in_period,attr,omitempty" json:"remainingTimeInPeriod,omitempty"` // Must match the pattern [0-9]+:[0-9]+|[0-9]+
	Stopped               *bool   `xml:"stopped,attr,omitempty" json:"stopped,omitempty"`
}

type PeriodScore struct {
	MatchStatusCode *int32   `xml:"match_status_code,attr" json:"matchStatusCode"`
	Number          *int32   `xml:"number,attr" json:"number"`
	HomeScore       *float64 `xml:"home_score,attr" json:"homeScore"`
	AwayScore       *float64 `xml:"away_score,attr" json:"awayScore"`
}

type Result struct {
	MatchStatusCode *int32   `xml:"match_status_code,attr" json:"matchStatusCode"`
	HomeScore       *float64 `xml:"home_score,attr" json:"homeScore"`
	AwayScore       *float64 `xml:"away_score,attr" json:"awayScore"`
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
