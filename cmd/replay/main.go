package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"net/http"
	_ "net/http/pprof"

	"github.com/minus5/go-uof-sdk"
	"github.com/minus5/go-uof-sdk/api"
	"github.com/minus5/go-uof-sdk/pipe"
	"github.com/minus5/go-uof-sdk/sdk"
)

const (
	EnvBookmakerID = "UOF_BOOKMAKER_ID"
	EnvToken       = "UOF_TOKEN"
)

func env(name string) string {
	val, ok := os.LookupEnv(name)
	if !ok {
		log.Printf("env %s not found", name)
	}
	return val
}

var (
	bookmakerID  string
	token        string
	scenarioID   int
	eventURN     uof.URN
	speed        int
	maxDelay     int
	outputFolder string
)

func init() {
	var event string

	flag.IntVar(&speed, "speed", 100, "replay speed, speed times faster than in reality")
	flag.IntVar(&maxDelay, "max-delay", 10, "maximum delay between messages in milliseconds (this is helpful especially in pre-match odds where delay can be even a few hours or more)")
	flag.IntVar(&scenarioID, "scenario", 0, "scenario (1,2 or 3) to replay")
	flag.StringVar(&event, "event", "", "event urn to replay")
	flag.StringVar(&outputFolder, "out", "./tmp", "output fodler location")
	flag.Parse()

	if scenarioID == 0 && event == "" {
		event = "sr:match:11830662"
		log.Printf("no event or scenario found, will replay sample event %s", event)
	}
	eventURN.Parse(event)

	token = env(EnvToken)
	bookmakerID = env(EnvBookmakerID)
}

func debugHTTP() {
	if err := http.ListenAndServe("localhost:8123", nil); err != nil {
		log.Fatal(err)
	}
}

func exitSignal() context.Context {
	ctx, stop := context.WithCancel(context.Background())
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c
		stop()
	}()
	return ctx
}

// UOF - Example replays:
// https://docs.betradar.com/display/BD/UOF+-+Example+replays
func main() {
	go debugHTTP()

	err := sdk.Run(exitSignal(),
		sdk.Credentials(bookmakerID, token),
		sdk.Languages(uof.Languages("en,de,hr")),
		sdk.BufferedConsumer(pipe.FileStore(outputFolder), 1024),
		sdk.Callback(progress),
		sdk.Replay(startReplay),
	)
	if err != nil {
		log.Fatal(err)
	}
}

func startReplay(rpl *api.ReplayApi) error {
	if !eventURN.Empty() {
		return rpl.StartEvent(eventURN, speed, maxDelay)
	}
	if scenarioID > 0 {
		return rpl.StartScenario(scenarioID, speed, maxDelay)
	}
	return nil
}

func progress(m *uof.Message) error {
	v := func() string {
		switch m.Type {
		case uof.MessageTypeAlive:
			return "a"
		case uof.MessageTypeMarkets:
			return "m"
		case uof.MessageTypePlayer:
			return "p"
		case uof.MessageTypeFixture:
			return "f"
		default:
			return "."
		}
	}
	fmt.Printf("%s", v())
	return nil
}
