package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

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

	timestamp := uof.CurrentTimestamp() - 5*60*1000
	var ps uof.ProducersChange
	ps.Add(uof.ProducerPrematch, timestamp)
	ps.Add(uof.ProducerLiveOdds, timestamp)

	errc := pipe.Build(
		queue.WithReconnect(sig, conn),
		pipe.Simple(logMessage),
		pipe.FileStore("./tmp/log"),
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
		buf, _ := json.Marshal(m.Producers)
		fmt.Printf("%-3d %s\n", m.Type, buf)
		return nil
	}
	return nil
}

func logMessage(m *uof.Message) error {
	if m.Type == uof.MessageTypeConnection {
		fmt.Printf("%-3d connection status: %s\n", m.Type, m.Connection.Status)
		return nil
	}
	if m.Type == uof.MessageTypeProducersChange {
		return logProducersChange(m)
	}

	b := m.Raw
	// remove xml header
	if i := bytes.Index(b, []byte("?>")); i > 0 {
		b = b[i+2:]
	}
	// show just first x characters
	if len(b) > 128 {
		b = b[:128]
	}
	fmt.Printf("%-3d %s\n", m.Type, b)
	return nil
}
