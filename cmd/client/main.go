package main

import (
	"bytes"
	"encoding/json"
	"fmt"
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

func main() {
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
	preloadTo := time.Now().Add(24 * time.Hour)
	//preloadTo := time.Now()
	timestamp := int64(0)
	//timestamp := uof.CurrentTimestamp() - 5*60*1000

	var ps uof.ProducersChange
	ps.Add(uof.ProducerPrematch, timestamp)
	ps.Add(uof.ProducerLiveOdds, timestamp)

	errc := pipe.Build(
		queue.WithReconnect(sig, conn),
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
	// if m.Type == uof.MessageTypeConnection {
	// 	fmt.Printf("%-25s connection status: %s\n", m.Type, m.Connection.Status)
	// 	return nil
	// }
	// if m.Type == uof.MessageTypeProducersChange {
	// 	return logProducersChange(m)
	// }

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
	return nil
}
