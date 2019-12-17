package api

import (
	"context"

	"github.com/minus5/go-uof-sdk"
)

// replay api paths
const (
	startScenario = "/v1/replay/scenario/play/{{.ScenarioID}}?speed={{.Speed}}&max_delay={{.MaxDelay}}&use_replay_timestamp={{.UseReplayTimestamp}}"
	replayStop    = "/v1/replay/stop"
	replayReset   = "/v1/replay/reset"
	replayAdd     = "/v1/replay/events/{{.EventURN}}"
	replayPlay    = "/v1/replay/play?speed={{.Speed}}&max_delay={{.MaxDelay}}&use_replay_timestamp={{.UseReplayTimestamp}}"
)

// Replay service for unified feed methods
func Replay(exitSig context.Context, token string) (*ReplayAPI, error) {
	r := &ReplayAPI{
		api: &API{
			server:  productionServer,
			token:   token,
			exitSig: exitSig,
		},
	}
	return r, r.Reset()
}

type ReplayAPI struct {
	api *API
}

// Start replay of the scenario from replay queue. Your current playlist will be
// wiped, and populated with events from specified scenario. Events are played
// in the order they were played in reality. Parameters 'speed' and 'max_delay'
// specify the speed of replay and what should be the maximum delay between
// messages. Default values for these are speed = 10 and max_delay = 10000. This
// means that messages will be sent 10x faster than in reality, and that if
// there was some delay between messages that was longer than 10 seconds it will
// be reduced to exactly 10 seconds/10 000 ms (this is helpful especially in
// pre-match odds where delay can be even a few hours or more). If player is
// already in play, nothing will happen.
func (r *ReplayAPI) StartScenario(scenarioID, speed, maxDelay int) error {
	return r.api.post(startScenario, &params{ScenarioID: scenarioID, Speed: speed, MaxDelay: maxDelay})
}

// StartEvent starts replay of a single event.
func (r *ReplayAPI) StartEvent(eventURN uof.URN, speed, maxDelay int) error {
	if err := r.Reset(); err != nil {
		return err
	}
	if err := r.Add(eventURN); err != nil {
		return err
	}
	return r.Play(speed, maxDelay)
}

// Adds to the end of the replay queue.
func (r *ReplayAPI) Add(eventURN uof.URN) error {
	return r.api.put(replayAdd, &params{EventURN: eventURN})
}

// Start replay the events from replay queue. Events are played in the order
// they were played in reality. Parameters 'speed' and 'max_delay' specify the
// speed of replay and what should be the maximum delay between messages.
// Default values for these are speed = 10 and max_delay = 10000. This means
// that messages will be sent 10x faster than in reality, and that if there was
// some delay between messages that was longer than 10 seconds it will be
// reduced to exactly 10 seconds/10 000 ms (this is helpful especially in
// pre-match odds where delay can be even a few hours or more). If player is
// already in play, nothing will happen.
func (r *ReplayAPI) Play(speed, maxDelay int) error {
	return r.api.post(replayPlay, &params{Speed: speed, MaxDelay: maxDelay})
}

// Stop the player if it is currently playing. If player is already stopped,
// nothing will happen.
func (r *ReplayAPI) Stop() error {
	return r.api.post(replayStop, nil)
}

// Stop the player if it is currently playing and clear the replay queue. If
// player is already stopped, the queue is cleared.
func (r *ReplayAPI) Reset() error {
	return r.api.post(replayReset, nil)
}
