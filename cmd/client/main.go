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
		uof.ProducerPrematch: 1567859558616,
		uof.ProducerLiveOdds: 1567859558616,
	}

	done(
		pipe.Recovery(stg, producers,
			pipe.ToMessage(
				queue.WithReconnect(sig, conn))))
	//conn.Listen()))
}

func done(in <-chan *uof.Message) {
	for m := range in {
		b := m.Body
		// remove xml header
		if i := bytes.Index(b, []byte("?>")); i > 0 {
			b = b[i+2:]
		}

		// show just first x characters
		if len(b) > 128 {
			b = b[:128]
		}
		fmt.Printf("%s\n", b)
	}
}
