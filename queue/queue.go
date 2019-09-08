// Package queue implements connection to the Betradar amqp queue

// You cannot create your own queues. Instead you have to request a server-named
// queue (empty queue name in the request). Passive, Exclusive, Non-durable.
// Reference: https://docs.betradar.com/display/BD/UOF+-+Messages
package queue

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/minus5/uof"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

const (
	replayServer     = "replaymq.betradar.com:5671"
	stagingServer    = "stgmq.betradar.com:5671"
	productionServer = "mq.betradar.com:5671"
	queueExchange    = "unifiedfeed"
	bindingKeyAll    = "#"
)

// Dial connects to the production queue
func Dial(ctx context.Context, bookmakerID, token string) (*Connection, error) {
	return dial(ctx, productionServer, bookmakerID, token)
}

// DialStaging connects to the staging queue
func DialStaging(ctx context.Context, bookmakerID, token string) (*Connection, error) {
	return dial(ctx, stagingServer, bookmakerID, token)
}

// DialReplay connects to the replay server
func DialReplay(ctx context.Context, bookmakerID, token string) (*Connection, error) {
	return dial(ctx, replayServer, bookmakerID, token)
}

type Connection struct {
	msgs   <-chan amqp.Delivery
	errs   <-chan *amqp.Error
	reDial func() (*Connection, error)
}

func (c *Connection) Listen() <-chan uof.QueueMsg {
	lastTs := CurrentTimestamp()
	uniqTimestamp := func() int64 {
		ts := CurrentTimestamp()
		if ts <= lastTs {
			ts += 1
		}
		lastTs = ts
		return ts
	}

	out := make(chan uof.QueueMsg)
	go func() {
		defer close(out)
		for m := range c.msgs {
			out <- uof.QueueMsg{
				RoutingKey: m.RoutingKey,
				Body:       m.Body,
				Timestamp:  uniqTimestamp(),
			}
		}
	}()
	go func() {
		for e := range c.errs {
			fmt.Printf("error: %s\n, code: %d, reason: %s, server: %v, recover: %v\n", e, e.Code, e.Reason, e.Server, e.Recover)
		}
		fmt.Printf("errs chan closed\n")
	}()

	return out
}

// CurrentTimestamp in milliseconds
func CurrentTimestamp() int64 {
	return timeToTimestamp(time.Now())
}
func timeToTimestamp(t time.Time) int64 {
	return t.UnixNano() / 1e6
}

func dial(ctx context.Context, server, bookmakerID, token string) (*Connection, error) {
	addr := fmt.Sprintf("amqps://%s:@%s//unifiedfeed/%s", token, server, bookmakerID)

	tls := &tls.Config{
		ServerName:         server,
		InsecureSkipVerify: true,
	}
	conn, err := amqp.DialTLS(addr, tls)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	chnl, err := conn.Channel()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	qee, err := chnl.QueueDeclare(
		"",    // name, leave empty to generate a unique name
		false, // durable
		true,  // delete when unused
		true,  // exclusive
		false, // noWait
		nil,   // arguments
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	err = chnl.QueueBind(
		qee.Name,      // name of the queue
		bindingKeyAll, // bindingKey
		queueExchange, // sourceExchange
		false,         // noWait
		nil,           // arguments
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	consumerTag := ""
	msgs, err := chnl.Consume(
		qee.Name,    // queue
		consumerTag, // consumerTag
		true,        // auto-ack
		true,        // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	errs := make(chan *amqp.Error)
	chnl.NotifyClose(errs)

	c := &Connection{
		msgs: msgs,
		errs: errs,
		reDial: func() (*Connection, error) {
			return dial(ctx, server, bookmakerID, token)
		},
	}

	go func() {
		<-ctx.Done()
		// cleanup on exit
		chnl.Cancel(consumerTag, true)
		conn.Close()
	}()

	return c, nil
}
