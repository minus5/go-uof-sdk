package main

import (
	"fmt"

	"github.com/minus5/go-uof-sdk"
)

type sampleEvent struct {
	Description string
	URN         uof.URN
}

func sampleEvents() []sampleEvent {
	// stolen from the ExampleReplayEvents.cs in the C# SDK
	return []sampleEvent{
		sampleEvent{"Soccer Match - English Premier League 2017 (Watford vs Westham)", "sr:match:11830662"},
		sampleEvent{"Soccer Match w Overtime - Primavera Cup", "sr:match:12865222"},
		sampleEvent{
			"Soccer Match w Overtime & Penalty Shootout - KNVB beker 17/18 - FC Twente Enschede vs Ajax Amsterdam",
			"sr:match:12873164"},
		sampleEvent{"Soccer Match with Rollback Betsettlement from Prematch Producer", "sr:match:11958226"},
		sampleEvent{
			"Soccer Match aborted mid-game - new match played later (first match considered cancelled according to betting rules}",
			"sr:match:11971876"},
		sampleEvent{"Soccer Match w PlayerProps (prematch odds only)", "sr:match:12055466"},
		sampleEvent{"Tennis Match - ATP Paris Final 2017", "sr:match:12927908"},
		sampleEvent{"Tennis Match where one of the players retired", "sr:match:12675240"},
		sampleEvent{"Tennis Match with bet_cancel adjustments using rollback_bet_cancel", "sr:match:13616027"},
		sampleEvent{
			"Tennis Match w voided markets due to temporary loss of coverage - no ability to verify results",
			"sr:match:13600533"},
		sampleEvent{"Basketball Match - NBA Final 2017 - (Golden State Warriors vs Cleveland Cavaliers)",
			"sr:match:11733773"},
		sampleEvent{"Basketball Match w voided DrawNoBet (2nd half draw)", "sr:match:12953638"},
		sampleEvent{"Basketball Match w PlayerProps", "sr:match:12233896"},
		sampleEvent{"Icehockey Match - NHL Final 2017 (6th match - Nashville Predators vs Pittsburg Penguins)",
			"sr:match:11784628"},
		sampleEvent{"Icehockey Match with Rollback BetCancel", "sr:match:11878140"},
		sampleEvent{"Icehockey Match with overtime + rollback_bet_cancel + match_status=\"aet\"",
			"sr:match:11878386"},
		sampleEvent{"American Football Game - NFL 2018/2018 (Chicago Bears vs Atlanta Falcons)",
			"sr:match:11538563"},
		sampleEvent{"American Football Game w PlayerProps", "sr:match:13552497"},
		sampleEvent{"Handball Match - DHB Pokal 17/18 (SG Flensburg-Handewitt vs Fuchse Berlin)",
			"sr:match:12362564"},
		sampleEvent{"Baseball Game - MLB 2017 (Final Los Angeles Dodgers vs Houston Astros)",
			"sr:match:12906380"},
		sampleEvent{"Badminton Game - Indonesia Masters 2018", "sr:match:13600687"},
		sampleEvent{"Snooker - International Championship 2017 (Final Best-of-19 frames)", "sr:match:12927314"},
		sampleEvent{"Darts - PDC World Championship 17/18 - (Final)", "sr:match:13451765"},
		sampleEvent{"CS:GO (ESL Pro League 2018)", "sr:match:13497893"},
		sampleEvent{"Dota2 (The International 2017 - Final)", "sr:match:12209528"},
		sampleEvent{"League of Legends Match (LCK Spring 2018)", "sr:match:13516251"},
		sampleEvent{"Cricket Match [Premium Cricket] - The Ashes 2017 (Australia vs England)",
			"sr:match:11836360"},
		sampleEvent{
			"Cricket Match (rain affected) [Premium Cricket] - ODI Series New Zealand vs. Pakistan 2018",
			"sr:match:13073610"},
		sampleEvent{"Volleyball Match (includes bet_cancels)", "sr:match:12716714"},
		sampleEvent{"Volleyball match where Betradar loses coverage mid-match - no ability to verify results",
			"sr:match:13582831"},
		sampleEvent{"Aussie Rules Match (AFL 2017 Final)", "sr:match:12587650"},
		sampleEvent{"Table Tennis Match (World Cup 2017 Final)", "sr:match:12820410"},
		sampleEvent{"Squash Match (Qatar Classic 2017)", "sr:match:12841530"},
		sampleEvent{"Beach Volleyball", "sr:match:13682571"},
		sampleEvent{"Badminton", "sr:match:13600687"},
		sampleEvent{"Bowls", "sr:match:13530237"},
		sampleEvent{"Rugby League", "sr:match:12979908"},
		sampleEvent{"Rugby Union", "sr:match:12420636"},
		sampleEvent{"Rugby Union 7s", "sr:match:13673067"},
		sampleEvent{"Handball", "sr:match:12362564"},
		sampleEvent{"Futsal", "sr:match:12363102"},
		sampleEvent{"Golf Winner Events + Three Balls - South African Open (Winner events + Three balls)",
			"sr:simple_tournament:66820"},
		sampleEvent{"Season Outrights (Long-term Outrights) - NFL 2017/18", "sr:season:40175"},
		sampleEvent{"Race Outrights (Short-term Outrights) - Cycling Tour Down Under 2018", "sr:stage:329361"},
	}
}

func showSampleEvents() {
	fmt.Printf("%-27s %s\n", "URN", "DESCRIPTION")
	for _, e := range sampleEvents() {
		fmt.Printf("%-27s %s\n", e.URN, e.Description)
	}
}
