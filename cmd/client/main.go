package main

import (
	"bytes"
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
	//go debugHTTP()

	sig := signal.InteruptContext()
	conn, err := queue.DialStaging(sig, bookmakerID, token)
	if err != nil {
		log.Fatal(err)
	}
	log.Debug("connected")

	//languages := uof.Languages("en,de,hr")
	stg := api.Staging(token)
	// TODO ping na startu

	producers := map[uof.Producer]int64{
		uof.ProducerPrematch: 1567960553490,
		uof.ProducerLiveOdds: 1567960553490,
	}

	errc := pipe.Build(
		queue.WithReconnect(sig, conn),
		pipe.Recovery(stg, producers),
		pipe.Simple(logMessage),
	)

	for err := range errc {
		//fmt.Printf("error: %s\n", err.Error())
		log.Error(err)
	}
}

func logMessage(m *uof.Message) error {
	if m.Type == uof.MessageTypeConnection {
		fmt.Printf("%-3d connection status: %s\n", m.Type, m.Connection.Status)
		return nil
	}

	b := m.Body
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
