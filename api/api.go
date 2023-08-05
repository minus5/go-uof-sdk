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
	"github.com/pvotal-tech/go-uof-sdk"
)

const (
	stagingServer          = "stgapi.betradar.com"
	productionServer       = "api.betradar.com"
	productionServerGlobal = "global.api.betradar.com"
)

var RequestTimeout = 32 * time.Second

type API struct {
	server  string
	token   string
	exitSig context.Context
	client  *retryablehttp.Client
}

// Dial connect to the staging or production api environment
func Dial(ctx context.Context, env uof.Environment, token string) (*API, error) {
	switch env {
	case uof.Replay:
		return Staging(ctx, token)
	case uof.Staging:
		return Staging(ctx, token)
	case uof.Production:
		return Production(ctx, token)
	case uof.ProductionGlobal:
		return ProductionGlobal(ctx, token)
	default:
		return nil, uof.Notice("queue dial", fmt.Errorf("unknown environment %d", env))
	}
}

// Staging connects to the staging system
func Staging(exitSig context.Context, token string) (*API, error) {
	a := &API{
		server:  stagingServer,
		token:   token,
		exitSig: exitSig,
		client:  client(),
	}
	return a, a.Ping()
}

// Production connects to the production system
func Production(exitSig context.Context, token string) (*API, error) {
	a := &API{
		server:  productionServer,
		token:   token,
		exitSig: exitSig,
		client:  client(),
	}
	return a, a.Ping()
}

// Production connects to the production system
func ProductionGlobal(exitSig context.Context, token string) (*API, error) {
	a := &API{
		server:  productionServerGlobal,
		token:   token,
		exitSig: exitSig,
		client:  client(),
	}
	return a, a.Ping()
}

func client() *retryablehttp.Client {
	c := retryablehttp.NewClient()
	c.Logger = nil
	c.RetryWaitMin = 1 * time.Second
	c.RetryWaitMax = 16 * time.Second
	c.RetryMax = 4
	return c
}

const (
	recovery     = "/v1/{{.Producer}}/recovery/initiate_request?after={{.Timestamp}}&request_id={{.RequestID}}&node_id={{.NodeID}}"
	fullRecovery = "/v1/{{.Producer}}/recovery/initiate_request?request_id={{.RequestID}}&node_id={{.NodeID}}"
	ping         = "/v1/users/whoami.xml"
)

func (a *API) RequestRecovery(producer uof.Producer, timestamp, requestID, nodeID int) error {
	if timestamp <= 0 {
		return a.RequestFullOddsRecovery(producer, requestID, nodeID)
	}
	return a.RequestRecoverySinceTimestamp(producer, timestamp, requestID, nodeID)
}

// RequestRecoverySinceTimestamp does recovery of odds and stateful messages
// over the feed since after timestamp. Subscribes client to feed messages.
func (a *API) RequestRecoverySinceTimestamp(producer uof.Producer, timestamp, requestID, nodeID int) error {
	return a.post(recovery, &params{Producer: producer, Timestamp: timestamp, RequestID: requestID, NodeID: nodeID})
}

// RequestFullOddsRecovery does recovery of odds over the feed. Subscribes
// client to feed messages.
func (a *API) RequestFullOddsRecovery(producer uof.Producer, requestID, nodeID int) error {
	return a.post(fullRecovery, &params{Producer: producer, RequestID: requestID, NodeID: nodeID})
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

func (a *API) Ping() error {
	_, err := a.get(ping, nil)
	return err
}

func (a *API) getAs(o interface{}, tpl string, p *params) error {
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
func (a *API) get(tpl string, p *params) ([]byte, error) {
	return a.httpRequest(tpl, p, "GET")
}

// make http put request
func (a *API) put(tpl string, p *params) error {
	_, err := a.httpRequest(tpl, p, "PUT")
	return err
}

// make http post request
func (a *API) post(tpl string, p *params) error {
	_, err := a.httpRequest(tpl, p, "POST")
	return err
}

func (a *API) httpRequest(tpl string, p *params, method string) ([]byte, error) {
	path := runTemplate(tpl, p)
	url := fmt.Sprintf("https://%s%s", a.server, path)

	req, err := retryablehttp.NewRequest(method, url, nil)
	if err != nil {
		return nil, uof.E("http.NewRequest", uof.APIError{URL: url, Inner: err})
	}
	if a.exitSig != nil {
		ctx, cancel := context.WithTimeout(a.exitSig, RequestTimeout)
		defer cancel()
		req = req.WithContext(ctx)
	}

	req.Header.Set("x-access-token", a.token)
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, uof.E("client.Do", uof.APIError{URL: url, Inner: err})
	}

	defer resp.Body.Close()
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, uof.E("http.Body", uof.APIError{URL: url, Inner: err})
	}

	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		return nil, uof.E("http.StatusCode", uof.APIError{URL: url, StatusCode: resp.StatusCode, Response: string(buf)})
	}

	return buf, nil
}

type params struct {
	EventURN           uof.URN
	ScenarioID         int
	Speed              int
	MaxDelay           int
	PlayerID           int
	MarketID           int
	Variant            string
	Timestamp          int
	RequestID          int
	Start              int
	Limit              int
	IncludeMappings    bool
	UseReplayTimestamp bool
	Lang               uof.Lang
	Producer           uof.Producer
	NodeID             int
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
	return buf.String()
}
