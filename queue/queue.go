// Package queue implements connection to the Betradar amqp queue
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
	stagingServer = "stgmq.betradar.com:5671"
	server        = "mq.betradar.com:5671"
)

func Dial(ctx context.Context, bookmakerID, token string) (*Connection, error) {
	return dial(ctx, server, bookmakerID, token)
}

func DialStaging(ctx context.Context, bookmakerID, token string) (*Connection, error) {
	return dial(ctx, stagingServer, bookmakerID, token)
}

type Connection struct {
	msgs <-chan amqp.Delivery
	errs <-chan *amqp.Error
}

func (c *Connection) Listen(out chan<- uof.QueueMsg) {
	for m := range c.msgs {
		out <- uof.QueueMsg{
			RoutingKey: m.RoutingKey,
			Body:       m.Body,
			Timestamp:  m.Timestamp,
		}
	}
	close(out)
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
		"#",           // bindingKey - bind to all messages
		"unifiedfeed", // sourceExchange
		false,         // noWait
		nil,           // arguments
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	msgs, err := chnl.Consume(
		qee.Name, // queue
		"",       // consumer
		true,     // auto-ack
		true,     // exclusive
		false,    // no-local
		false,    // no-wait
		nil,      // args
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	errs := make(chan *amqp.Error)
	chnl.NotifyClose(errs)

	c := &Connection{
		msgs: msgs,
		errs: errs,
	}

	go func() {
		<-ctx.Done()
		chnl.Cancel("", false)
		conn.Close()
	}()

	return c, nil
}
