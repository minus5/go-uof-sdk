package sdk

import (
	"context"
	"time"

	"github.com/minus5/go-uof-sdk"
	"github.com/minus5/go-uof-sdk/api"
	"github.com/minus5/go-uof-sdk/pipe"
	"github.com/minus5/go-uof-sdk/queue"
)

var defaultLanuages = uof.Languages("en,de")

type Config struct {
	BookmakerID       string
	Token             string
	Staging           bool
	Languages         []uof.Lang
	FixturesPreloadTo time.Time
	RecoveryFrom      int
	Stages            []pipe.StageHandler
	Replay            func(*api.ReplayApi) error
}

// Option is a function on the options for a connection.
type Option func(*Config) error

// Run starts uof connector
// Call to run blocks until stopped by context, or error occured.
// Order in wich options are set is not important.
// Credentials and one Callback or Pipe are functional minimum.
func Run(ctx context.Context, options ...Option) error {
	c := config(options...)
	qc, apiConn, err := connect(ctx, c)
	if err != nil {
		return err
	}
	if c.Replay != nil {
		rpl, err := api.Replay(ctx, c.Token)
		if err != nil {
			return err
		}
		if err := c.Replay(rpl); err != nil {
			return err
		}
	}

	stages := []pipe.StageHandler{
		pipe.Markets(apiConn, c.Languages),
		pipe.Fixture(apiConn, c.Languages, c.FixturesPreloadTo),
		pipe.Player(apiConn, c.Languages),
		pipe.BetStop(),
	}
	if c.RecoveryFrom > 0 {
		// TODO: what producers
		var ps uof.ProducersChange
		ps.Add(uof.ProducerPrematch, c.RecoveryFrom)
		ps.Add(uof.ProducerLiveOdds, c.RecoveryFrom)
		stages = append(stages, pipe.Recovery(apiConn, ps))
	}
	stages = append(stages, c.Stages...)

	errc := pipe.Build(
		queue.WithReconnect(ctx, qc),
		stages...,
	)
	return firstErr(errc)
}

func firstErr(errc <-chan error) error {
	var err error
	for e := range errc {
		if err == nil {
			err = e
		}
	}
	return err
}

func config(options ...Option) Config {
	c := &Config{}
	for _, o := range options {
		o(c)
	}
	if len(c.Languages) == 0 {
		c.Languages = defaultLanuages
	}
	return *c
}

func connect(ctx context.Context, c Config) (*queue.Connection, *api.Api, error) {
	if c.Replay != nil {
		conn, err := queue.DialReplay(ctx, c.BookmakerID, c.Token)
		if err != nil {
			return nil, nil, err
		}
		stg, err := api.Staging(ctx, c.Token)
		if err != nil {
			return nil, nil, err
		}
		return conn, stg, nil
	}

	if c.Staging {
		conn, err := queue.DialStaging(ctx, c.BookmakerID, c.Token)
		if err != nil {
			return nil, nil, err
		}
		stg, err := api.Staging(ctx, c.Token)
		if err != nil {
			return nil, nil, err
		}
		return conn, stg, nil
	}

	conn, err := queue.Dial(ctx, c.BookmakerID, c.Token)
	if err != nil {
		return nil, nil, err
	}
	stg, err := api.Production(ctx, c.Token)
	if err != nil {
		return nil, nil, err
	}
	return conn, stg, nil
}

// Credentials for establishing connection to the uof queue and api.
func Credentials(bookmakerID, token string) Option {
	return func(c *Config) error {
		c.BookmakerID = bookmakerID
		c.Token = token
		return nil
	}
}

// Staging forces use of staging environment instead of production.
func Staging() Option {
	return func(c *Config) error {
		c.Staging = true
		return nil
	}
}

// Languages for api calls.
// All api calls will be made for all the languages specified here.
// If not specified defaultLanguages will be used.
func Languages(langs []uof.Lang) Option {
	return func(c *Config) error {
		c.Languages = langs
		return nil
	}
}

// Pipe sets chan handler for all messages.
// Can be called multiple times.
func Pipe(s pipe.StageHandler) Option {
	return func(c *Config) error {
		c.Stages = append(c.Stages, s)
		return nil
	}
}

// Callback sets handler for all messages.
// If returns error will break the pipe and force exit from sdk.Run.
// Can be called multiple times.
func Callback(cb func(m *uof.Message) error) Option {
	return func(c *Config) error {
		c.Stages = append(c.Stages, pipe.Simple(cb))
		return nil
	}
}

// Replay forces use of replay environment.
// Callback will be called to start replay after establishing connection.
func Replay(cb func(*api.ReplayApi) error) Option {
	return func(c *Config) error {
		c.Replay = cb
		return nil
	}
}

func RecoveryFrom(ts int) Option {
	return func(c *Config) error {
		c.RecoveryFrom = ts
		return nil
	}
}
func PreloadFixturesTo(to time.Time) Option {
	return func(c *Config) error {
		c.FixturesPreloadTo = to
		return nil
	}
}
