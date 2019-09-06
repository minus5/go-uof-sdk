package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/minus5/svckit/file"
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
	val, ok := os.LookupEnv(name)
	if !ok {
		log.Errorf("env %s not found", name)
	}
	return val
}

var (
	bookmakerID  string
	token        string
	scenarioID   int
	eventID      int
	sample       bool
	speed        int
	maxDelay     int
	outputFolder string
)

func init() {
	var show bool
	flag.IntVar(&speed, "speed", 100, "replay speed, speed times faster than in reality")
	flag.IntVar(&maxDelay, "max-delay", 10, "maximum delay between messages in milliseconds (this is helpful especially in pre-match odds where delay can be even a few hours or more)")
	flag.IntVar(&scenarioID, "scenario", 0, "scenario (1,2 or 3) to replay")
	flag.IntVar(&eventID, "event", 0, "event to replay")
	flag.BoolVar(&sample, "sample", false, "replay sample events")
	flag.BoolVar(&show, "show", false, "show interesting sample events and exit")
	flag.StringVar(&outputFolder, "out", "./tmp", "output fodler location")
	flag.Parse()

	if show {
		showSampleEvents()
		os.Exit(0)
	}
	token = env(EnvToken)
	bookmakerID = env(EnvBookmakerID)
}

func main() {
	conn, err := queue.DialReplay(signal.InteruptContext(), bookmakerID, token)
	must(err)
	log.Debug("connected")

	languages := uof.Languages("en,de,hr")
	stg := api.Staging(token)

	startReplay()

	done(
		pipe.FileStore(outputFolder,
			pipe.VariantMarket(stg, languages,
				pipe.Player(stg, languages,
					pipe.Fixture(stg, languages,
						pipe.Markets(stg, languages,
							pipe.ToMessage(
								conn.Listen())))))))
}

func startReplay() {
	rpl := api.Replay(token)
	if eventID > 0 {
		must(rpl.StartEvent(uof.NewEventURN(eventID), speed, maxDelay))
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

func done(in <-chan *uof.Message) {
	for m := range in {
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
		}()
		fmt.Printf("%s", v)
	}
}

func saveMsgs(in <-chan uof.QueueMsg) <-chan uof.QueueMsg {
	out := make(chan uof.QueueMsg, 128)
	go func() {
		defer close(out)
		for m := range in {
			out <- m
			saveMsg(m)
		}
	}()
	return out
}

func saveMsg(m uof.QueueMsg) {
	fn := fmt.Sprintf("%s/%011d-%s", outputFolder, m.Timestamp, m.RoutingKey)
	if err := file.Save(fn, m.Body); err != nil {
		log.Fatal(err)
	}
}
