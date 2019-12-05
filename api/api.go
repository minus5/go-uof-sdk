// Package api connects to the Unified Feed API interface
package api

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"text/template"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	uof "github.com/minus5/go-uof-sdk"
)

const (
	stagingServer    = "stgapi.betradar.com"
	productionServer = "api.betradar.com"
)

var RequestTimeout = 32 * time.Second

type Api struct {
	server  string
	token   string
	exitSig context.Context
}

// Staging connects to the staging system
func Staging(exitSig context.Context, token string) (*Api, error) {
	a := &Api{
		server:  stagingServer,
		token:   token,
		exitSig: exitSig,
	}
	return a, a.Ping()
}

// Production connects to the production system
func Production(exitSig context.Context, token string) (*Api, error) {
	a := &Api{
		server:  productionServer,
		token:   token,
		exitSig: exitSig,
	}
	return a, a.Ping()
}

const (
	recovery     = "/v1/{{.Producer}}/recovery/initiate_request?after={{.Timestamp}}&request_id={{.RequestID}}"
	fullRecovery = "/v1/{{.Producer}}/recovery/initiate_request?request_id={{.RequestID}}"
	ping         = "/v1/users/whoami.xml"
)

func (a *Api) RequestRecovery(producer uof.Producer, timestamp int, requestID int) error {
	if timestamp <= 0 {
		return a.RequestFullOddsRecovery(producer, requestID)
	}
	return a.RequestRecoverySinceTimestamp(producer, timestamp, requestID)
}

// RequestRecoverySinceTimestamp does recovery of odds and stateful messages
// over the feed since after timestamp. Subscribes client to feed messages.
func (a *Api) RequestRecoverySinceTimestamp(producer uof.Producer, timestamp int, requestID int) error {
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

func (a *Api) Ping() error {
	_, err := a.get(ping, nil)
	return err
}

func (a *Api) getAs(o interface{}, tpl string, p *params) error {
	buf, err := a.get(tpl, p)
	if err != nil {
		return err
	}
	if err := xml.Unmarshal(buf, o); err != nil {
		return uof.Notice("unmarshal", err)
	}
	return nil
}

// make http get request
func (a *Api) get(tpl string, p *params) ([]byte, error) {
	return a.httpRequest(tpl, p, "GET")
}

// make http put request
func (a *Api) put(tpl string, p *params) error {
	_, err := a.httpRequest(tpl, p, "PUT")
	return err
}

// make http post request
func (a *Api) post(tpl string, p *params) error {
	_, err := a.httpRequest(tpl, p, "POST")
	return err
}

func (a *Api) httpRequest(tpl string, p *params, method string) ([]byte, error) {
	path := runTemplate(tpl, p)
	url := fmt.Sprintf("https://%s%s", a.server, path)

	req, err := retryablehttp.NewRequest(method, url, nil)
	if err != nil {
		return nil, uof.E("http.NewRequest", uof.ApiError{URL: url, Inner: err})
	}
	if a.exitSig != nil {
		ctx, _ := context.WithTimeout(a.exitSig, RequestTimeout)
		req = req.WithContext(ctx)
	}

	client := retryablehttp.NewClient()
	client.Logger = nil
	client.RetryWaitMin = 1 * time.Second
	client.RetryWaitMax = 16 * time.Second
	client.RetryMax = 4

	req.Header.Set("x-access-token", a.token)
	resp, err := client.Do(req)
	if err != nil {
		return nil, uof.E("client.Do", uof.ApiError{URL: url, Inner: err})
	}

	defer resp.Body.Close()
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, uof.E("http.Body", uof.ApiError{URL: url, Inner: err})
	}

	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		return nil, uof.E("http.StatusCode", uof.ApiError{URL: url, StatusCode: resp.StatusCode, Response: string(buf)})
	}

	return buf, nil
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
	Timestamp          int
	RequestID          int
	Start              int
	Limit              int
}

func runTemplate(def string, p *params) string {
	if p == nil {
		return def
	}
	tpl := template.Must(template.New("").Parse(def))
	buf := bytes.NewBuffer(nil)
	if err := tpl.Execute(buf, p); err != nil {
		panic(err)
	}
	return string(buf.Bytes())
}
