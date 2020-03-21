package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/minus5/go-uof-sdk"
	"github.com/minus5/go-uof-sdk/pipe"
	"github.com/minus5/go-uof-sdk/sdk"
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
	if err := http.ListenAndServe("localhost:8125", nil); err != nil {
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

	var preloadTo time.Time                         // zero
	timestamp := uof.CurrentTimestamp() - 5*60*1000 // -5 minutes
	var pc uof.ProducersChange
	pc.AddAll(uof.VirtualProducers(), timestamp)

	err := sdk.Run(exitSignal(),
		sdk.Credentials(bookmakerID, token),
		sdk.Staging(),
		sdk.BindVirtuals(),
		sdk.Recovery(pc),
		sdk.Fixtures(preloadTo),
		sdk.Languages(uof.Languages("en,hr")),
		sdk.BufferedConsumer(pipe.FileStore("./tmp"), 1024),
		sdk.Consumer(logMessages),
	)
	if err != nil {
		log.Fatal(err)
	}
}

// consumer of incomming messages
func logMessages(in <-chan *uof.Message) error {
	for m := range in {
		logMessage(m)
	}
	return nil
}

func logMessage(m *uof.Message) {
	switch m.Type {
	case uof.MessageTypeConnection:
		fmt.Printf("%-25s status: %s\n", m.Type, m.Connection.Status)
	case uof.MessageTypeFixture:
		fmt.Printf("%-25s lang: %s, urn: %s raw: %d\n", m.Type, m.Lang, m.Fixture.URN, len(m.Raw))
	case uof.MessageTypeMarkets:
		fmt.Printf("%-25s lang: %s, count: %d\n", m.Type, m.Lang, len(m.Markets))
	case uof.MessageTypeAlive:
		if m.Alive.Subscribed != 0 {
			fmt.Printf("%-25s producer: %s, timestamp: %d\n", m.Type, m.Alive.Producer, m.Alive.Timestamp)
		}
	case uof.MessageTypeOddsChange:
		fmt.Printf("%-25s event: %s, markets: %d\n", m.Type, m.EventURN, len(m.OddsChange.Markets))
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
