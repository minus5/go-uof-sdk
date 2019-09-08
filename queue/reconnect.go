package queue

import (
	"context"
	"time"

	"github.com/minus5/svckit/signal"
	"github.com/minus5/uof"
)

var (
	maxInterval    = 16 * time.Second // max interval for exponential backoff
	maxElapsedTime = 1 * time.Hour    // will give up if not connected longer than this
)

// WithReconnect ensuers reconnects with exponential backoff interval
func WithReconnect(ctx context.Context, conn *Connection) <-chan uof.QueueMsg {
	out := make(chan uof.QueueMsg)

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
		return err
	}

	go func() {
		defer close(out)
		for {
			// TODO signal connect
			drain(conn, out)
			if done() {
				return
			}
			// TODO signal connection lost
			if err := signal.WithBackoff(ctx, reconnect, maxInterval, maxElapsedTime); err != nil {
				return
			}
		}
	}()

	return out
}

// drain consumes from connection until msgs chan is closed
func drain(conn *Connection, out chan<- uof.QueueMsg) {
	go func() {
		for _ = range conn.errs {
			// TODO
		}
	}()

	for m := range conn.msgs {
		out <- uof.QueueMsg{
			RoutingKey: m.RoutingKey,
			Body:       m.Body,
			Timestamp:  0, // TODO uniqTimestamp()
		}
	}
}
