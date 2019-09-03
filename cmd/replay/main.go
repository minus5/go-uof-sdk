package main

import (
	"fmt"
	"os"

	"github.com/minus5/svckit/log"
	"github.com/minus5/svckit/signal"

	"github.com/minus5/uof/queue"
)

const (
	EnvBookmakerID = "UOF_BOOKMAKER_ID"
	EnvToken       = "UOF_TOKEN"
)

func env(name string) string {
	val, ok := os.LookupEnv(name)
	if !ok {
		log.Errorf("env %s not found", name)
	}
	return val
}

func main() {
	ctx := signal.InteruptContext()
	conn, err := queue.DialReplay(ctx, env(EnvBookmakerID), env(EnvToken))
	if err != nil {
		log.Fatal(err)
	}
	log.Debug("connected")

	for m := range conn.Listen() {
		fmt.Printf("%s\n", m.Body)
	}
}
