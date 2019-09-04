// Package api connects to the Unified Feed API interface
package api

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"text/template"
)

const (
	stagingServer    = "stgapi.betradar.com"
	productionServer = "api.betradar.com"
)

const (
	startScenario = "/v1/replay/scenario/play/{{.ScenarioID}}?speed={{.Speed}}&max_delay={{.MaxDelay}}&use_replay_timestamp={{.UseReplayTimestamp}}"
	replayStop    = "/v1/replay/stop"
	replayReset   = "/v1/replay/reset"
	replayAdd     = "/v1/replay/events/{{.EventURN}}"
	replayPlay    = "/v1/replay/play?speed={{.Speed}}&max_delay={{.MaxDelay}}&use_replay_timestamp={{.UseReplayTimestamp}}"
)

type Api struct {
	server string
	token  string
}

// Staging connects to the staging system
func Staging(token string) *Api {
	return &Api{
		server: stagingServer,
		token:  token,
	}
}

// Production connects to the production system
func Production(token string) *Api {
	return &Api{
		server: productionServer,
		token:  token,
	}
}

// Replay service for unified feed methods
func  Replay(token string) *ReplayApi {
	return &ReplayApi{
		api: &Api{
			server: productionServer,
			token:  token,
		},
	}
}

type ReplayApi struct {
	api *Api
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
func (r *ReplayApi) StartScenario(scenarioID, speed, maxDelay int) error {
	return r.api.post(startScenario, &params{ScenarioID: scenarioID, Speed: speed, MaxDelay: maxDelay})
}

// StartEvent starts replay of a single event.
func (r *ReplayApi) StartEvent(eventURN string, speed, maxDelay int) error {
	if err := r.Reset(); err != nil {
		return err
	}
	if err := r.Add(eventURN); err != nil {
		return err
	}
	return r.Play(speed, maxDelay)
}

// Adds to the end of the replay queue.
func (r *ReplayApi) Add(eventURN string) error {
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
func (r *ReplayApi) Play(speed, maxDelay int) error {
	return r.api.post(replayPlay, &params{Speed: speed, MaxDelay: maxDelay})
}

// Stop the player if it is currently playing. If player is already stopped,
// nothing will happen.
func (r *ReplayApi) Stop() error {
	return r.api.post(replayStop, nil)
}

// Stop the player if it is currently playing and clear the replay queue. If
// player is already stopped, the queue is cleared.
func (r *ReplayApi) Reset() error {
	return r.api.post(replayReset, nil)
}

// // RequestRecoverySinceTimestamp does recovery of odds and stateful messages
// // over the feed since after timestamp. Subscribes client to feed messages.
// func (a *Api) RequestRecoverySinceTimestamp(product string, timestamp int64) error {
// 	return a.post(fmt.Sprintf("/v1/%s/recovery/initiate_request?after=%d", product, timestamp))
// }

// // RequestFullOddsRecovery does recovery of odds over the feed. Subscribes
// // client to feed messages.
// func (a *Api) RequestFullOddsRecovery(product string) error {
// 	return a.post(fmt.Sprintf("/v1/%s/recovery/initiate_request", product))
// }

// // RecoverSportEvent requests to resend all odds for all markets for a sport
// // event.
// func (a *Api) RecoverSportEvent(product, eventID string) error {
// 	return a.post(fmt.Sprintf("/v1/%s/events/%s/initiate_request", product, eventID))
// }

// // RecoverStatefulForSportEvent requests to resend all stateful-messages
// // (BetSettlement, RollbackBetSettlement, BetCancel, UndoBetCancel) for a sport
// // event.
// func (a *Api) RecoverStatefulForSportEvent(product, eventID string) error {
// 	return a.post(fmt.Sprintf("/v1/%s/stateful_messages/events/%s/initiate_request", product, eventID))
// }

// http get request
func (a *Api) get(path string) ([]byte, int, error) {
	url := fmt.Sprintf("https://%s/%s", a.server, path)
	client := http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("x-access-token", a.token)
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}
	if resp.StatusCode != 200 {
		return buf, resp.StatusCode, fmt.Errorf("status code: %d, response: %s", resp.StatusCode, buf)
	}
	return buf, resp.StatusCode, nil
}

func (a *Api) put(tpl string, p *params) error {
	return a.httpRequest(tpl, p, "PUT")
}

// http post request
func (a *Api) post(tpl string, p *params) error {
	return a.httpRequest(tpl, p, "POST")
}

func (a *Api) httpRequest(tpl string, p *params, method string) error {
	path := runTemplate(tpl, p)
	url := fmt.Sprintf("https://%s%s", a.server, path)
	client := http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("x-access-token", a.token)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		buf, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("status code %d, url: %s, rsp: %s", resp.StatusCode, url, buf)
	}

	return nil
}

type params struct {
	EventURN           string
	ScenarioID         int
	Speed              int
	MaxDelay           int
	UseReplayTimestamp bool
}

func runTemplate(def string, p *params) string {
	if p == nil {
		return def
	}
	tpl := template.Must(template.New("").Parse(def))
	buf := bytes.NewBuffer(nil)
	if err := tpl.Execute(buf, p); err != nil {
		log.Fatal(err)
	}
	return string(buf.Bytes())
}
