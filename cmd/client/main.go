package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/minus5/svckit/log"
	"github.com/minus5/svckit/signal"
	"github.com/minus5/uof"
	"github.com/minus5/uof/api"
	"github.com/minus5/uof/pipe"
	"github.com/minus5/uof/queue"
)

const (
	EnvBookmakerID = "UOF_BOOKMAKER_ID"
	EnvToken       = "UOF_TOKEN"
)

func env(name string) string {
	e, ok := os.LookupEnv(name)
	if !ok {
		log.Errorf("env %s not found", name)
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

func main() {
	go debugHTTP()

	sig := signal.InteruptContext()
	conn, err := queue.DialStaging(sig, bookmakerID, token)
	if err != nil {
		log.Fatal(err)
	}
	stg, err := api.Staging(token)
	if err != nil {
		log.Fatal(err)
	}
	log.Debug("connected")

	languages := uof.Languages("en,de,hr")

	//preloadTo := time.Now().Add(24 * time.Hour)
	// timestamp := int64(0)
	preloadTo := time.Now()
	timestamp := uof.CurrentTimestamp() - 5*60*1000

	var ps uof.ProducersChange
	ps.Add(uof.ProducerPrematch, timestamp)
	ps.Add(uof.ProducerLiveOdds, timestamp)

	errc := pipe.Build(
		queue.WithReconnect(sig, conn),
		pipe.Markets(stg, languages),
		pipe.Fixture(stg, languages, preloadTo),
		pipe.Simple(logMessage),
		pipe.FileStore("./tmp"),
		pipe.Recovery(stg, ps),
		pipe.Simple(logProducersChange),
	)

	for err := range errc {
		//fmt.Printf("error: %s\n", err.Error())
		log.Error(err)
	}
}

func logProducersChange(m *uof.Message) error {
	if m.Type == uof.MessageTypeProducersChange {
		return logMessage(m)
	}
	return nil
}

func logMessage(m *uof.Message) error {
	switch m.Type {
	case uof.MessageTypeConnection:
		fmt.Printf("%-25s status: %s\n", m.Type, m.Connection.Status)
	case uof.MessageTypeFixture:
		fmt.Printf("%-25s lang: %s, urn: %s\n", m.Type, m.Lang, m.Fixture.URN)
		// case uof.MessageTypeOddsChange:
		// 	fmt.Printf("%-25s urn: %s\n", m.Type, m.Lang, m.Fixture.URN)
		// 	return nil
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
	return nil
}
