// Package queue implements connection to the Betradar amqp queue

// You cannot create your own queues. Instead you have to request a server-named
// queue (empty queue name in the request). Passive, Exclusive, Non-durable.
// Reference: https://docs.betradar.com/display/BD/UOF+-+Messages
package queue

import (
	"context"
	"crypto/tls"
	"fmt"

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

func (c *Connection) Listen() (<-chan *uof.Message, <-chan error) {
	out := make(chan *uof.Message)
	errc := make(chan error)
	go func() {
		defer close(out)
		defer close(errc)
		c.drain(out, errc)
	}()
	return out, errc

}

// drain consumes from connection until msgs chan is closed
func (c *Connection) drain(out chan<- *uof.Message, errc chan<- error) {
	errsDone := make(chan struct{})
	go func() {
		for err := range c.errs {
			errc <- errors.Wrap(err, "amqp error")
		}
		close(errsDone)
	}()

	for m := range c.msgs {
		m, err := uof.NewQueueMessage(m.RoutingKey, m.Body)
		if err != nil {
			errc <- errors.Wrap(err, "fail to parse delivery")
			continue
		}
		out <- m
	}
	<-errsDone
}

func dial(ctx context.Context, server, bookmakerID, token string) (*Connection, error) {
	addr := fmt.Sprintf("amqps://%s:@%s//unifiedfeed/%s", token, server, bookmakerID)

	tls := &tls.Config{
		ServerName:         server,
		InsecureSkipVerify: true,
	}
	conn, err := amqp.DialTLS(addr, tls)
	if err != nil {
		fmt.Println(addr)
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
