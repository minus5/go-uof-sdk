// Package queue implements connection to the Betradar amqp queue

// You cannot create your own queues. Instead you have to request a server-named
// queue (empty queue name in the request). Passive, Exclusive, Non-durable.
// Reference: https://docs.betradar.com/display/BD/UOF+-+Messages
package queue

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/pvotal-tech/go-uof-sdk"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	replayServer           = "replaymq.betradar.com:5671"
	stagingServer          = "stgmq.betradar.com:5671"
	productionServer       = "mq.betradar.com:5671"
	productionServerGlobal = "global.mq.betradar.com:5671"
	queueExchange          = "unifiedfeed"
	bindingKeyAll          = "#"
	amqpDefaultHeartbeat   = 10 * time.Second
	amqpDefaultLocale      = "en_US"
)

// Dial connects to the queue chosen by environment
func Dial(ctx context.Context, env uof.Environment, bookmakerID, token string, nodeID int, isTLS bool) (*Connection, error) {
	switch env {
	case uof.Replay:
		return DialReplay(ctx, bookmakerID, token, nodeID, isTLS)
	case uof.Staging:
		return DialStaging(ctx, bookmakerID, token, nodeID, isTLS)
	case uof.Production:
		return DialProduction(ctx, bookmakerID, token, nodeID, isTLS)
	case uof.ProductionGlobal:
		return DialProductionGlobal(ctx, bookmakerID, token, nodeID, isTLS)
	default:
		return nil, uof.Notice("queue dial", fmt.Errorf("unknown environment %d", env))
	}
}

// Dial connects to the production queue
func DialProduction(ctx context.Context, bookmakerID, token string, nodeID int, isTLS bool) (*Connection, error) {
	return dial(ctx, productionServer, bookmakerID, token, nodeID, isTLS)
}

// Dial connects to the production queue
func DialProductionGlobal(ctx context.Context, bookmakerID, token string, nodeID int, isTLS bool) (*Connection, error) {
	return dial(ctx, productionServerGlobal, bookmakerID, token, nodeID, isTLS)
}

// DialStaging connects to the staging queue
func DialStaging(ctx context.Context, bookmakerID, token string, nodeID int, isTLS bool) (*Connection, error) {
	return dial(ctx, stagingServer, bookmakerID, token, nodeID, isTLS)
}

// DialReplay connects to the replay server
func DialReplay(ctx context.Context, bookmakerID, token string, nodeID int, isTLS bool) (*Connection, error) {
	return dial(ctx, replayServer, bookmakerID, token, nodeID, isTLS)
}

// DialCustom connects to a custom server
func DialCustom(ctx context.Context, server, bookmakerID, token string, nodeID int, isTLS bool) (*Connection, error) {
	return dial(ctx, server, bookmakerID, token, nodeID, isTLS)
}

type Connection struct {
	msgs   <-chan amqp.Delivery
	errs   <-chan *amqp.Error
	reDial func() (*Connection, error)
	info   ConnectionInfo
}

type ConnectionInfo struct {
	server     string
	local      string
	network    string
	tlsVersion uint16
	nodeID     int
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
			errc <- uof.E("conn", err)
		}
		close(errsDone)
	}()

	for m := range c.msgs {
		m, err := uof.NewQueueMessage(m.RoutingKey, m.Body)
		if err != nil {
			errc <- uof.Notice("conn.DeliveryParse", err)
			continue
		}
		// ignores messages that are of no interest to the current session
		if m.NodeID != 0 && m.NodeID != c.info.nodeID {
			return
		}
		out <- m
	}
	<-errsDone
}

func dial(ctx context.Context, server, bookmakerID, token string, nodeID int, isTLS bool) (*Connection, error) {
	addr := fmt.Sprintf("amqps://%s:@%s//unifiedfeed/%s", token, server, bookmakerID)

	tlsConfig := &tls.Config{
		ServerName:         server,
		InsecureSkipVerify: true,
	}
	var dialer func(network string, addr string) (net.Conn, error)
	if isTLS {
		dialer = func(network string, addr string) (net.Conn, error) {
			return tls.Dial(network, addr, tlsConfig)
		}
	}
	config := amqp.Config{
		Heartbeat:       amqpDefaultHeartbeat,
		TLSClientConfig: tlsConfig,
		Locale:          amqpDefaultLocale,
		Dial:            dialer,
	}
	conn, err := amqp.DialConfig(addr, config)
	if err != nil {
		fmt.Println(strings.ReplaceAll(addr, token, "<token>"))
		return nil, uof.Notice("conn.Dial", err)
	}

	chnl, err := conn.Channel()
	if err != nil {
		return nil, uof.Notice("conn.Channel", err)
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
		return nil, uof.Notice("conn.QueueDeclare", err)
	}

	err = chnl.QueueBind(
		qee.Name,      // name of the queue
		bindingKeyAll, // bindingKey
		queueExchange, // sourceExchange
		false,         // noWait
		nil,           // arguments
	)
	if err != nil {
		return nil, uof.Notice("conn.QueueBind", err)
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
		return nil, uof.Notice("conn.Consume", err)
	}

	errs := make(chan *amqp.Error)
	chnl.NotifyClose(errs)

	c := &Connection{
		msgs: msgs,
		errs: errs,
		reDial: func() (*Connection, error) {
			return dial(ctx, server, bookmakerID, token, nodeID, isTLS)
		},
		info: ConnectionInfo{
			server:     server,
			local:      conn.LocalAddr().String(),
			network:    conn.LocalAddr().Network(),
			tlsVersion: conn.ConnectionState().Version,
			nodeID:     nodeID,
		},
	}

	go func() {
		<-ctx.Done()
		// cleanup on exit
		_ = chnl.Cancel(consumerTag, true)
		conn.Close()
	}()

	return c, nil
}
