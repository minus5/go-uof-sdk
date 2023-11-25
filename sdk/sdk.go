package sdk

import (
	"context"
	"time"

	"github.com/pvotal-tech/go-uof-sdk"
	"github.com/pvotal-tech/go-uof-sdk/api"
	"github.com/pvotal-tech/go-uof-sdk/pipe"
	"github.com/pvotal-tech/go-uof-sdk/queue"
)

var defaultLanguages = uof.Languages("en,de")

// ErrorListenerFunc listens all SDK errors
type ErrorListenerFunc func(err error)

// Config is active SDK configuration
type Config struct {
	CustomAMQPServer   string
	CustomAPIServer    string
	BookmakerID        int
	Token              string
	NodeID             int
	IsAMQPTLS          bool
	IsThrottled        bool
	PrefetchCount      int
	ConcurrentAPIFetch bool
	AutoAckDisabled    bool
	PipelineDisabled   bool
	Fixtures           time.Time
	Recovery           []uof.ProducerChange
	Stages             []pipe.InnerStage
	Replay             func(*api.ReplayAPI) error
	Env                uof.Environment
	Languages          []uof.Lang
	ErrorListener      ErrorListenerFunc
}

// Option sets attributes on the Config.
type Option func(*Config)

// Run starts uof connector.
//
// Call to Run blocks until stopped by context, or error occurred.
// Order in which options are set is not important.
// Credentials and one of Callback or Pipe are functional minimum.
func Run(parentCtx context.Context, options ...Option) error {
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()
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
	stages := make([]pipe.InnerStage, 0)
	if !c.PipelineDisabled {
		stages = append(stages,
			pipe.Markets(apiConn, c.Languages),
			pipe.Fixture(apiConn, c.Languages, c.Fixtures),
			pipe.Player(apiConn, c.Languages, c.ConcurrentAPIFetch),
			pipe.BetStop(),
		)
	}
	if len(c.Recovery) > 0 {
		stages = append(stages, pipe.Recovery(apiConn, c.Recovery, c.NodeID))
	}
	stages = append(stages, c.Stages...)

	errc := pipe.Build(
		qc.Listen,
		stages...,
	)
	return firstErr(errc, c.ErrorListener)
}

func firstErr(errc <-chan error, errorListener ErrorListenerFunc) error {
	var err error
	for e := range errc {
		if err == nil {
			err = e
		}
		if errorListener != nil {
			errorListener(e)
		}
	}
	return err
}

func config(options ...Option) Config {
	// defaults
	c := &Config{
		Languages: defaultLanguages,
		Env:       uof.Production,
	}
	for _, o := range options {
		o(c)
	}
	return *c
}

// connect to the queue and api
func connect(ctx context.Context, c Config) (*queue.Connection, *api.API, error) {
	var conn *queue.Connection
	var amqpErr error
	if c.CustomAMQPServer != "" {
		conn, amqpErr = queue.DialCustom(ctx, c.CustomAMQPServer, c.BookmakerID, c.Token, c.NodeID, c.PrefetchCount, c.IsAMQPTLS, c.IsThrottled, !c.AutoAckDisabled)
	} else {
		conn, amqpErr = queue.Dial(ctx, c.Env, c.BookmakerID, c.Token, c.NodeID, c.PrefetchCount, c.IsAMQPTLS, c.IsThrottled, !c.AutoAckDisabled)
	}
	if amqpErr != nil {
		return nil, nil, amqpErr
	}

	var apiConn *api.API
	var apiErr error
	if c.CustomAPIServer != "" {
		apiConn, apiErr = api.DialCustom(ctx, c.CustomAPIServer, c.Token)
	} else {
		apiConn, apiErr = api.Dial(ctx, c.Env, c.Token)
	}
	if apiErr != nil {
		return nil, nil, apiErr
	}
	return conn, apiConn, nil
}

// Credentials for establishing connection to the uof queue and api.
func Credentials(bookmakerID int, token string, nodeID int) Option {
	return func(c *Config) {
		c.BookmakerID = bookmakerID
		c.Token = token
		c.NodeID = nodeID
	}
}

// CustomServers for establishing custom servers
func CustomServers(customAMQPServer, customAPIServer string) Option {
	return func(c *Config) {
		c.CustomAMQPServer = customAMQPServer
		c.CustomAPIServer = customAPIServer
	}
}

// ConfigTLS for setting tls flag
func ConfigTLS(isAMQPTLS bool) Option {
	return func(c *Config) {
		c.IsAMQPTLS = isAMQPTLS
	}
}

// ConfigThrottle is Throttled uses channel.Get internally. prefetchCount is to be used when isThrottled is false
func ConfigThrottle(isThrottled bool, prefetchCount int) Option {
	return func(c *Config) {
		c.IsThrottled = isThrottled
		c.PrefetchCount = prefetchCount
	}
}

// DisablePipeline disables internal stages for processing markets, fixtures, players
func DisablePipeline() Option {
	return func(c *Config) {
		c.PipelineDisabled = true
	}
}

// ConfigConcurrentAPIFetch for auto-fetching rest api concurrently
func ConfigConcurrentAPIFetch(concurrentAPIFetch bool) Option {
	return func(c *Config) {
		c.ConcurrentAPIFetch = concurrentAPIFetch
	}
}

// DisableAutoAck requires usage of exposed message functions Ack/NackRequeue/NackDiscard
func DisableAutoAck() Option {
	return func(c *Config) {
		c.AutoAckDisabled = true
	}
}

// Languages for api calls.
//
// Statefull messages (markets, players, fixtures) will be served in all this
// languages. Each language requires separate call to api. If not specified
// `defaultLanguages` will be used.
func Languages(langs []uof.Lang) Option {
	return func(c *Config) {
		c.Languages = langs
	}
}

// Global forces use of global production environment.
func Global() Option {
	return func(c *Config) {
		c.Env = uof.ProductionGlobal
	}
}

// Staging forces use of staging environment instead of production.
func Staging() Option {
	return func(c *Config) {
		c.Env = uof.Staging
	}
}

// Replay forces use of replay environment.
// Callback will be called to start replay after establishing connection.
func Replay(cb func(*api.ReplayAPI) error) Option {
	return func(c *Config) {
		c.Env = uof.Replay
		c.Replay = cb
	}
}

// Consumer sets chan consumer of the SDK messages stream.
//
// Consumer should range over `in` chan and handle all messages.
// In chan will be closed on SDK tear down.
// If the consumer returns an error it is handled as fatal. Immediately closes SDK connection.
// Can be called multiple times.
func Consumer(consumer pipe.ConsumerStage) Option {
	return func(c *Config) {
		c.Stages = append(c.Stages, pipe.Consumer(consumer))
	}
}

// BufferedConsumer same as consumer but with buffered `in` chan of size `buffer`.
func BufferedConsumer(consumer pipe.ConsumerStage, buffer int) Option {
	return func(c *Config) {
		c.Stages = append(c.Stages, pipe.BufferedConsumer(consumer, buffer))
	}
}

// Callback sets handler for all messages.
//
// If returns error will break the pipe and force exit from sdk.Run.
// Can be called multiple times.
func Callback(cb func(m *uof.Message) error) Option {
	return func(c *Config) {
		c.Stages = append(c.Stages, pipe.Simple(cb))
	}
}

// Recovery starts recovery for each producer
//
// It is responsibility of SDK consumer to track the last timestamp of the
// successfully consumed message for each producer. On startup this timestamp is
// sent here and SDK will request recovery; get all the messages after that ts.
//
// Ref: https://docs.betradar.com/display/BD/UOF+-+Recovery+using+API
func Recovery(pc []uof.ProducerChange) Option {
	return func(c *Config) {
		c.Recovery = pc
	}
}

// Fixtures gets live and pre-match fixtures at start-up.
//
// It gets fixture for all matches which starts before `to` time.
// There is a special endpoint to get almost all fixtures before initiating
// recovery. This endpoint is designed to significantly reduce the number of API
// calls required during recovery.
//
// Ref: https://docs.betradar.com/display/BD/UOF+-+Fixtures+in+the+API
func Fixtures(to time.Time) Option {
	return func(c *Config) {
		c.Fixtures = to
	}
}

// ListenErrors sets ErrorListener for all SDK errors
func ListenErrors(listener ErrorListenerFunc) Option {
	return func(c *Config) {
		c.ErrorListener = listener
	}
}
