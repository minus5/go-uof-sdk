package queue

import (
	"context"
	"time"

	"github.com/minus5/svckit/signal"
	"github.com/minus5/uof"
	"github.com/pkg/errors"
)

var (
	maxInterval    = 16 * time.Second // max interval for exponential backoff
	maxElapsedTime = 1 * time.Hour    // will give up if not connected longer than this
)

// WithReconnect ensuers reconnects with exponential backoff interval
func WithReconnect(ctx context.Context, conn *Connection) func() (<-chan *uof.Message, <-chan error) {
	return func() (<-chan *uof.Message, <-chan error) {
		out := make(chan *uof.Message)
		errc := make(chan error)

		done := func() bool {
			select {
			case <-ctx.Done():
				return true
			default:
				return false
			}
		}

		reconnect := func() error {
			nc, err := conn.reDial()
			if err == nil {
				conn = nc // replace existing with new connection
			}
			if err != nil {
				errc <- errors.Wrap(err, "reconnect failed")
			}
			return err
		}

		go func() {
			defer close(out)
			defer close(errc)
			for {
				out <- uof.NewConnnectionMessage(uof.ConnectionStatusUp) // signal connect
				drain(conn, out, errc)
				if done() {
					return
				}
				out <- uof.NewConnnectionMessage(uof.ConnectionStatusDown) // signal connection lost
				if err := signal.WithBackoff(ctx, reconnect, maxInterval, maxElapsedTime); err != nil {
					return
				}
			}
		}()

		return out, errc
	}
}

// drain consumes from connection until msgs chan is closed
func drain(conn *Connection, out chan<- *uof.Message, errc chan<- error) {
	go func() {
		for err := range conn.errs {
			errc <- errors.Wrap(err, "amqp error")
		}
	}()

	for m := range conn.msgs {
		m, err := uof.NewQueueMessage(m.RoutingKey, m.Body)
		if err != nil {
			errc <- errors.Wrap(err, "fail to parse delivery")
			continue
		}
		out <- m
	}
}
