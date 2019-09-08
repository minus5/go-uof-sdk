// Package api connects to the Unified Feed API interface
package api

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"text/template"

	"github.com/minus5/uof"
	"github.com/pkg/errors"
)

const (
	stagingServer    = "stgapi.betradar.com"
	productionServer = "api.betradar.com"
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

const (
	recovery     = "/v1/{{.Producer}}/recovery/initiate_request?after={{.Timestamp}}&request_id={{.RequestID}}"
	fullRecovery = "/v1/{{.Producer}}/recovery/initiate_request&request_id={{.RequestID}}"
)

func (a *Api) RequestRecovery(producer uof.Producer, timestamp int64, requestID int) error {
	if timestamp <= 0 {
		return a.RequestFullOddsRecovery(producer, requestID)
	}
	return a.RequestRecoverySinceTimestamp(producer, timestamp, requestID)
}

// RequestRecoverySinceTimestamp does recovery of odds and stateful messages
// over the feed since after timestamp. Subscribes client to feed messages.
func (a *Api) RequestRecoverySinceTimestamp(producer uof.Producer, timestamp int64, requestID int) error {
	return a.post(recovery, &params{Producer: producer, Timestamp: timestamp, RequestID: requestID})
}

// RequestFullOddsRecovery does recovery of odds over the feed. Subscribes
// client to feed messages.
func (a *Api) RequestFullOddsRecovery(producer uof.Producer, requestID int) error {
	return a.post(fullRecovery, &params{Producer: producer, RequestID: requestID})
}

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
func (a *Api) get(tpl string, p *params) ([]byte, error) {
	path := runTemplate(tpl, p)
	url := fmt.Sprintf("https://%s%s", a.server, path)
	client := http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	req.Header.Set("x-access-token", a.token)
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer resp.Body.Close()
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if resp.StatusCode != 200 {
		err := fmt.Errorf("status code: %d\npath: %s\nresponse: %s", resp.StatusCode, path, buf)
		return nil, errors.WithStack(err)
	}
	return buf, nil
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
		return errors.WithStack(err)
	}

	req.Header.Set("x-access-token", a.token)
	resp, err := client.Do(req)
	if err != nil {
		return errors.WithStack(err)
	}
	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		buf, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("status code %d, url: %s, rsp: %s", resp.StatusCode, url, buf)
	}

	return nil
}

type params struct {
	EventURN           uof.URN
	ScenarioID         int
	Speed              int
	MaxDelay           int
	UseReplayTimestamp bool
	Lang               uof.Lang
	PlayerID           int
	MarketID           int
	Variant            string
	IncludeMappings    bool
	Producer           uof.Producer
	Timestamp          int64
	RequestID          int
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
