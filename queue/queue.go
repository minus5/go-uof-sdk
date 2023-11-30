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
	"time"

	"github.com/pvotal-tech/go-uof-sdk"
	amqp "github.com/rabbitmq/amqp091-go"
	log "github.com/sirupsen/logrus"
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
func Dial(ctx context.Context, env uof.Environment, bookmakerID int, token string, nodeID, prefetchCount int, isTLS, isThrottled, autoAck bool) (*Connection, error) {
	switch env {
	case uof.Replay:
		return DialReplay(ctx, bookmakerID, token, nodeID, prefetchCount, isTLS, isThrottled, autoAck)
	case uof.Staging:
		return DialStaging(ctx, bookmakerID, token, nodeID, prefetchCount, isTLS, isThrottled, autoAck)
	case uof.Production:
		return DialProduction(ctx, bookmakerID, token, nodeID, prefetchCount, isTLS, isThrottled, autoAck)
	case uof.ProductionGlobal:
		return DialProductionGlobal(ctx, bookmakerID, token, nodeID, prefetchCount, isTLS, isThrottled, autoAck)
	default:
		return nil, uof.Notice("queue dial", fmt.Errorf("unknown environment %d", env))
	}
}

// DialProduction connects to the production queue
func DialProduction(ctx context.Context, bookmakerID int, token string, nodeID, prefetchCount int, isTLS, isThrottled, autoAck bool) (*Connection, error) {
	return dial(ctx, productionServer, bookmakerID, token, nodeID, prefetchCount, isTLS, isThrottled, autoAck)
}

// DialProductionGlobal connects to the production queue
func DialProductionGlobal(ctx context.Context, bookmakerID int, token string, nodeID, prefetchCount int, isTLS, isThrottled, autoAck bool) (*Connection, error) {
	return dial(ctx, productionServerGlobal, bookmakerID, token, nodeID, prefetchCount, isTLS, isThrottled, autoAck)
}

// DialStaging connects to the staging queue
func DialStaging(ctx context.Context, bookmakerID int, token string, nodeID, prefetchCount int, isTLS, isThrottled, autoAck bool) (*Connection, error) {
	return dial(ctx, stagingServer, bookmakerID, token, nodeID, prefetchCount, isTLS, isThrottled, autoAck)
}

// DialReplay connects to the replay server
func DialReplay(ctx context.Context, bookmakerID int, token string, nodeID, prefetchCount int, isTLS, isThrottled, autoAck bool) (*Connection, error) {
	return dial(ctx, replayServer, bookmakerID, token, nodeID, prefetchCount, isTLS, isThrottled, autoAck)
}

// DialCustom connects to a custom server
func DialCustom(ctx context.Context, server string, bookmakerID int, token string, nodeID, prefetchCount int, isTLS, isThrottled, autoAck bool) (*Connection, error) {
	return dial(ctx, server, bookmakerID, token, nodeID, prefetchCount, isTLS, isThrottled, autoAck)
}

type Connection struct {
	isThrottled bool
	autoAck     bool
	msgs        <-chan amqp.Delivery
	errs        <-chan *amqp.Error
	chnl        *amqp.Channel
	queueName   string
	reDial      func() (*Connection, error)
	info        ConnectionInfo
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
		out <- uof.NewDetailedConnnectionMessage(uof.ConnectionStatusUp, c.info.server, c.info.local, c.info.network, c.info.tlsVersion)
		c.drain(out, errc)
	}()
	return out, errc
}

func (c *Connection) drain(out chan<- *uof.Message, errc chan<- error) {
	if c.isThrottled {
		c.drainThrottled(out, errc)
	} else {
		c.drainContinuous(out, errc)
	}
}

// drain consumes from connection until msgs chan is closed. consumes as fast as it can from queue
func (c *Connection) drainContinuous(out chan<- *uof.Message, errc chan<- error) {
	errsDone := make(chan struct{})
	go func() {
		for err := range c.errs {
			errc <- uof.E("conn", err)
		}
		close(errsDone)
	}()

	for d := range c.msgs {
		delivery := d
		readAt := time.Now().UTC()
		m, err := uof.NewQueueMessage(delivery.RoutingKey, delivery.Body)
		if err != nil {
			errc <- uof.Notice("conn.DeliveryParse", err)
			continue
		}
		// ignores messages that are of no interest to the current session
		if m.NodeID != 0 && m.NodeID != c.info.nodeID {
			errc <- uof.Notice("conn.drainContinuous", fmt.Errorf("unknown node ID. ignoring message: %d", m.NodeID))
			continue
		}

		m.EnabledAutoAck = c.autoAck
		m.Delivery = &delivery
		m.ReadAt = readAt
		out <- m
	}
	<-errsDone
}

// drain gets a single message and blocks on next iteration if out has not been consumed
func (c *Connection) drainThrottled(out chan<- *uof.Message, errc chan<- error) {
	errsDone := make(chan struct{})
	go func() {
		for err := range c.errs {
			errc <- uof.E("conn", err)
		}
		close(errsDone)
	}()

	for {
		delivery, hasMsg, err := c.chnl.Get(c.queueName, c.autoAck)
		if err != nil {
			errc <- uof.Notice("conn.channelGet", err)
			break
		}
		if !hasMsg {
			continue
		}
		readAt := time.Now().UTC()
		m, err := uof.NewQueueMessage(delivery.RoutingKey, delivery.Body)
		if err != nil {
			errc <- uof.Notice("conn.DeliveryParse", err)
			continue
		}
		m.EnabledAutoAck = c.autoAck
		m.Delivery = &delivery
		m.PendingMsgCount = int(delivery.MessageCount)
		m.ReadAt = readAt

		// ignores messages that are of no interest to the current session
		if m.NodeID != 0 && m.NodeID != c.info.nodeID {
			errc <- uof.Notice("conn.drainThrottled", fmt.Errorf("unknown node ID. ignoring message: %d", m.NodeID))
			continue
		}

		out <- m
	}
	<-errsDone
}

func dial(ctx context.Context, server string, bookmakerID int, token string, nodeID, prefetchCount int, isTLS, isThrottled, autoAck bool) (*Connection, error) {
	addr := fmt.Sprintf("amqps://%s:@%s//unifiedfeed/%d", token, server, bookmakerID)

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
		return nil, uof.Notice("conn.Dial", fmt.Errorf("%w (%s)", err, addr))
	}

	chnl, err := conn.Channel()
	if err != nil {
		if err := conn.Close(); err != nil {
			return nil, uof.Notice("conn.Close", err)
		}
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
		if err := conn.Close(); err != nil {
			return nil, uof.Notice("conn.Close", err)
		}
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
		if err := conn.Close(); err != nil {
			return nil, uof.Notice("conn.Close", err)
		}
		return nil, uof.Notice("conn.QueueBind", err)
	}

	consumerTag := ""
	var msgs <-chan amqp.Delivery
	if !isThrottled {
		err := chnl.Qos(prefetchCount, 0, true)
		if err != nil {
			if err := conn.Close(); err != nil {
				return nil, uof.Notice("conn.Close", err)
			}
			return nil, uof.Notice("conn.Qos", err)
		}
		msgs, err = chnl.Consume(
			qee.Name,    // queue
			consumerTag, // consumerTag
			autoAck,     // auto-ack
			false,       // exclusive
			false,       // no-local
			false,       // no-wait
			nil,         // args
		)
		if err != nil {
			if err := conn.Close(); err != nil {
				return nil, uof.Notice("conn.Close", err)
			}
			return nil, uof.Notice("conn.Consume", err)
		}
	}

	errs := make(chan *amqp.Error)
	chnl.NotifyClose(errs)

	c := &Connection{
		isThrottled: isThrottled,
		autoAck:     autoAck,
		msgs:        msgs,
		errs:        errs,
		chnl:        chnl,
		queueName:   qee.Name,
		reDial: func() (*Connection, error) {
			return dial(ctx, server, bookmakerID, token, nodeID, prefetchCount, isTLS, isThrottled, autoAck)
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
		if err := chnl.Cancel(consumerTag, true); err != nil {
			log.WithError(err).Error("error while closing channel")
		} else {
			log.Info("channel closed OK")
		}
		if err := conn.Close(); err != nil {
			log.WithError(err).Error("error while closing connection")
		} else {
			log.Info("connection closed OK")
		}
	}()

	return c, nil
}
