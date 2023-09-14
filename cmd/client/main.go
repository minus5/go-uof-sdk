package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pvotal-tech/go-uof-sdk"
	"github.com/pvotal-tech/go-uof-sdk/pipe"
	"github.com/pvotal-tech/go-uof-sdk/sdk"
)

const (
	EnvBookmakerID = "UOF_BOOKMAKER_ID"
	EnvToken       = "UOF_TOKEN"
)

func env(name string) string {
	e, ok := os.LookupEnv(name)
	if !ok {
		log.Printf("env %s not found", name)
	}
	return e
}

var (
	bookmakerID string
	token       string
)

func init() {
	token = env(EnvToken)
	bookmakerID = env(EnvBookmakerID)
}

func debugHTTP() {
	if err := http.ListenAndServe("localhost:8124", nil); err != nil {
		log.Fatal(err)
	}
}

func exitSignal() context.Context {
	ctx, stop := context.WithCancel(context.Background())
	go func() {
		c := make(chan os.Signal, 1)
		//SIGINT je ctrl-C u shell-u, SIGTERM salje upstart kada se napravi sudo stop ...
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c
		stop()
	}()
	return ctx
}

func main() {
	go debugHTTP()

	preloadTo := time.Now().Add(24 * time.Hour)

	timestamp := uof.CurrentTimestamp() - 5*60*1000 // -5 minutes
	var pc uof.ProducersChange
	pc.Add(uof.ProducerPrematch, timestamp)
	pc.Add(uof.ProducerLiveOdds, timestamp)

	err := sdk.Run(exitSignal(),
		sdk.Credentials(bookmakerID, token, 123),
		sdk.Staging(),
		sdk.Recovery(pc),
		sdk.Fixtures(preloadTo),
		sdk.Languages(uof.Languages("en,de,hr")),
		sdk.BufferedConsumer(pipe.FileStore("./tmp"), 1024),
		sdk.Consumer(logMessages),
		sdk.ListenErrors(listenSDKErrors),
	)
	if err != nil {
		log.Fatal(err)
	}
}

// consumer of incoming messages
func logMessages(in <-chan *uof.Message) error {
	for m := range in {
		logMessage(m)
	}
	return nil
}

func logMessage(m *uof.Message) {
	switch m.Type {
	case uof.MessageTypeConnection:
		fmt.Printf("%-25s status: %s, server: %s, local: %s, network: %s, tls: %s\n", m.Type, m.Connection.Status, m.Connection.ServerName, m.Connection.LocalAddr, m.Connection.Network, m.Connection.TLSVersionToString())
	case uof.MessageTypeFixture:
		fmt.Printf("%-25s lang: %s, urn: %s raw: %d\n", m.Type, m.Lang, m.Fixture.URN, len(m.Raw))
	case uof.MessageTypeMarkets:
		fmt.Printf("%-25s lang: %s, count: %d\n", m.Type, m.Lang, len(m.Markets))
	case uof.MessageTypeAlive:
		if m.Alive.Subscribed != 0 {
			fmt.Printf("%-25s producer: %s, timestamp: %d\n", m.Type, m.Alive.Producer, m.Alive.Timestamp)
		}
	case uof.MessageTypeBetSettlement:
		for _, v := range m.BetSettlement.Markets {
			fmt.Printf("BET SETTLEMENT producer=%v eventID=%d marketID=%v status=%v\n", m.Producer, m.BetSettlement.EventURN.ID(), v.ID, v.Result)
		}
	case uof.MessageTypeBetStop:
		for _, v := range m.BetStop.MarketIDs {
			fmt.Printf("BET STOP producer=%v eventID=%d marketID=%v status=%v\n", m.Producer, m.BetStop.EventURN.ID(), v, m.BetStop.Status)
		}
	case uof.MessageTypeOddsChange:
		fmt.Printf("ODDS CHANGE producer=%v eventID=%d eventStatus=%v\n", m.Producer, m.OddsChange.EventURN.ID(), m.OddsChange.EventStatus)
		for _, v := range m.OddsChange.Markets {
			fmt.Printf("ODDS CHANGE producer=%v eventID=%d marketID=%v status=%v\n", m.Producer, m.OddsChange.EventURN.ID(), v.ID, v.Status)
		}
	default:
		var b []byte
		if false && m.Raw != nil {
			b = m.Raw
			// remove xml header
			if i := bytes.Index(b, []byte("?>")); i > 0 {
				b = b[i+2:]
			}
		} else {
			b, _ = json.Marshal(m.Body)
		}
		// show just first x characters
		x := 186
		if len(b) > x {
			b = b[:x]
		}
		fmt.Printf("%-25s %s\n", m.Type, b)
	}
}

// listenSDKErrors listens all SDK errors for logging or any other pourpose
func listenSDKErrors(err error) {
	// example handling SDK typed errors
	var eu uof.Error
	if errors.As(err, &eu) {
		// use uof.Error attributes to build custom logging
		var logLine string
		if eu.Severity == uof.NoticeSeverity {
			logLine = fmt.Sprintf("NOTICE Operation:%s Details:", eu.Op)
		} else {
			logLine = fmt.Sprintf("ERROR Operation:%s Details:", eu.Op)
		}

		if eu.Inner != nil {
			var ea uof.APIError
			if errors.As(eu.Inner, &ea) {
				// use uof.APIError attributes for custom logging
				logLine = fmt.Sprintf("%s URL:%s", logLine, ea.URL)
				logLine = fmt.Sprintf("%s StatusCode:%d", logLine, ea.StatusCode)
				logLine = fmt.Sprintf("%s Response:%s", logLine, ea.Response)
				if ea.Inner != nil {
					logLine = fmt.Sprintf("%s Inner:%s", logLine, ea.Inner)
				}

				// or just log error as is...
				//log.Print(ea.Error())
			} else {
				// not an uof.APIError
				logLine = fmt.Sprintf("%s %s", logLine, eu.Inner)
			}
		}
		log.Println(logLine)

		// or just log error as is...
		//log.Println(eu.Error())
	} else {
		// any other error not uof.Error
		log.Println(err)
	}
}
