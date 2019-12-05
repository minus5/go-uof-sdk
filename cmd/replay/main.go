package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"net/http"
	_ "net/http/pprof"

	uof "github.com/minus5/go-uof-sdk"
	"github.com/minus5/go-uof-sdk/api"
	"github.com/minus5/go-uof-sdk/pipe"
	"github.com/minus5/go-uof-sdk/queue"
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
	sample       bool
	speed        int
	maxDelay     int
	outputFolder string
)

func init() {
	var show bool
	var event string

	flag.IntVar(&speed, "speed", 100, "replay speed, speed times faster than in reality")
	flag.IntVar(&maxDelay, "max-delay", 10, "maximum delay between messages in milliseconds (this is helpful especially in pre-match odds where delay can be even a few hours or more)")
	flag.IntVar(&scenarioID, "scenario", 0, "scenario (1,2 or 3) to replay")
	flag.StringVar(&event, "event", "", "event to replay")
	flag.BoolVar(&sample, "sample", false, "replay sample events")
	flag.BoolVar(&show, "show", false, "show interesting sample events and exit")
	flag.StringVar(&outputFolder, "out", "./tmp", "output fodler location")
	flag.Parse()

	if show {
		showSampleEvents()
		os.Exit(0)
	}
	if event != "" {
		eventURN = uof.URN(event)
		if id, err := strconv.Atoi(event); err == nil {
			eventURN = uof.NewEventURN(id)
		}
	}
	token = env(EnvToken)
	bookmakerID = env(EnvBookmakerID)
}

func debugHTTP() {
	if err := http.ListenAndServe("localhost:8123", nil); err != nil {
		log.Fatal(err)
	}
}

func interuptContext() context.Context {
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

// UOF - Example replays:
// https://docs.betradar.com/display/BD/UOF+-+Example+replays
func main() {
	go debugHTTP()

	sig := interuptContext()
	conn, err := queue.DialReplay(sig, bookmakerID, token)
	must(err)

	languages := uof.Languages("en,de,hr")
	stg, err := api.Staging(sig, token)
	must(err)
	startReplay(sig)
	var zero time.Time

	fmt.Printf("▶️")
	errc := pipe.Build(
		queue.WithReconnect(sig, conn),
		pipe.Markets(stg, languages),
		pipe.Fixture(stg, languages, zero),
		pipe.Player(stg, languages),
		pipe.BetStop(),
		pipe.FileStore("./tmp"),
		pipe.Simple(show),
	)

	for err := range errc {
		var ue uof.Error

		fmt.Printf("%s ", time.Now().Format("2006-01-02 15:04:05"))
		if errors.As(err, &ue) {
			fmt.Println(ue.Error())
		} else {
			fmt.Printf("unknown error %s\n", err)
		}
	}
}

func startReplay(sig context.Context) {
	rpl, err := api.Replay(sig, token)
	must(err)
	if string(eventURN) != "" {
		must(rpl.StartEvent(eventURN, speed, maxDelay))
	}
	if scenarioID > 0 {
		must(rpl.StartScenario(scenarioID, speed, maxDelay))
	}
	if sample {
		must(rpl.Reset())
		for _, s := range sampleEvents() {
			must(rpl.Add(s.URN))
		}
		must(rpl.Play(speed, maxDelay))
	}
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func show(m *uof.Message) error {
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
