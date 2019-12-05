package queue

import (
	"context"
	"time"

	"github.com/cenkalti/backoff/v3"
	"github.com/minus5/go-uof-sdk"
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
				conn.drain(out, errc)
				if done() {
					return
				}
				out <- uof.NewConnnectionMessage(uof.ConnectionStatusDown) // signal connection lost
				if err := withBackoff(ctx, reconnect, maxInterval, maxElapsedTime); err != nil {
					return
				}
			}
		}()

		return out, errc
	}
}

func withBackoff(ctx context.Context, op func() error,
	maxInterval, maxElapsedTime time.Duration) error {
	b := backoff.NewExponentialBackOff()
	b.MaxInterval = maxInterval
	b.MaxElapsedTime = maxElapsedTime
	bc := backoff.WithContext(b, ctx)
	return backoff.Retry(op, bc)
}
