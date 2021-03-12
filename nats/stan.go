package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cenk/backoff"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
	"opentelemetry-util/pubsub"
	"time"
)

type config struct {
	tracer     pubsub.TracePubSub
	stanOption []stan.Option
}

type Options func(*config)

func WithTracer(tracer pubsub.TracePubSub) Options {
	return func(c *config) {
		c.tracer = tracer
	}
}

func WithLogger() Options {
	return func(c *config) {
		//todo
	}
}

func WithStanOptions(o ...stan.Option) Options {
	return func(c *config) {
		c.stanOption = o
	}
}

type QueueSuber interface {
	pubsub.PubSuber
	QueueSub(ctx context.Context)
}

// StanClient 訂閱方
type StanClient struct {
	base pubsub.TracePubSub
	*config
	conn          stan.Conn
	handlerQueues []HandlerQueue
}

type HandlerQueue struct {
	Subject string
	Group   string
	Handler pubsub.SubHandler
}

func NewStanClient(clusterID, clientID, appID string, natsClient *NatsClient) *StanClient {
	return NewStanClientWithOption(clusterID, clientID, appID, natsClient)
}

func NewStanClientWithOption(clusterID, clientID, appID string, natsClient *NatsClient, opts ...Options) *StanClient {
	c := &config{}
	for _, opt := range opts {
		opt(c)
	}
	con, err := NewStanConn(clusterID, clientID, appID, natsClient.conn, c.stanOption...)
	if err != nil {
		//todo print err log
	}
	return &StanClient{
		conn:   con,
		config: c,
	}
}

func NewStanConn(clusterID, clientID, appID string, natsCon *nats.Conn, opt ...stan.Option) (stan.Conn, error) {

	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = 3 * time.Minute
	var con stan.Conn

	natsOpt := append(
		opt,
		stan.NatsConn(natsCon),
		stan.Pings(10, 5),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			//todo print lose connection log
			//log.Error().Msgf("Connection lost, reason: %v", reason)
		}),
	)

	err := backoff.Retry(func() error {
		var err error
		if err != nil {
			//log.Error().Msgf("fail to connect to nats: %s, addr: %v", err.Error(), c.Address)
			//todo log
			return err
		}

		switch {
		case clientID == "":
			clientID = uuid.New().String()
		case appID != "":
			clientID = fmt.Sprintf("%s_%s", appID, clientID)
		default:
			clientID = fmt.Sprintf("%s_%s", clientID, uuid.New().String())
		}

		con, err = stan.Connect(clusterID, clientID, natsOpt...)

		if err != nil {
			//log.Error().Msgf("fail to connect to stan: %s, clusterID: %v, clientID: %v", err.Error(), c.ClusterID, c.ClientID)
			//todo print stan connect error log
			return err
		}

		return nil
	}, bo)
	return con, err
}

func (sc *StanClient) RegisterHandlerHandlerQueueGroup(handlerQueues []HandlerQueue) {
	sc.handlerQueues = handlerQueues
}

func (sc *StanClient) Pub(ctx context.Context, topic string, message []byte) error {
	b, err := json.Marshal(message)
	if err != nil {
		//todo print error
	}

	return sc.conn.Publish(topic, sc.base.Pub(ctx, topic, b)())
}

func (sc *StanClient) Sub(ctx context.Context, topic string, handler pubsub.SubHandler) {
	_, err := sc.conn.Subscribe(topic, func(msg *stan.Msg) {
		sc.base.Sub(ctx, topic, handler, msg.Data)

	})
	if err != nil {
		//todo print error
	}
}

func (sc *StanClient) QueueSub(ctx context.Context) {
	for _, handler := range sc.handlerQueues {
		_, err := sc.conn.QueueSubscribe(
			handler.Subject,
			handler.Group,
			func(msg *stan.Msg) {
				sc.base.Sub(ctx, handler.Subject, handler.Handler, msg.Data)
			})
		if err != nil {
			//todo print error
		}
	}
}

func (sc *StanClient) DrainCon() {
	err := sc.conn.NatsConn().Drain()
	if err != nil {
		//todo print error
	}
}
