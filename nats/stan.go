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
	"github.com/rs/zerolog/log"
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
		//todo maybe we can add log interface,Easier to replace the log system
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
	*config
	conn          stan.Conn
	handlerQueues []HandlerQueue
}

type HandlerQueue struct {
	Subject string
	Group   string
	Handler pubsub.SubHandler
	Options []stan.SubscriptionOption
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
		log.Warn().Msgf("New stan connection error :%v", err.Error())
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
			log.Warn().Msgf("stan client connect stan lose  connection  :%v", reason.Error())
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
			log.Warn().Msgf("stan client connect stan server %v cluster id client id %v  error :%v", clientID, clientID, err.Error())
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
		log.Warn().Msgf("stan publish marshal data %v error :%v", string(message), err.Error())
	}
	err = sc.conn.Publish(topic, sc.tracer.Pub(ctx, topic, b)())
	if err != nil {
		log.Warn().Msgf("stan client publish %v topic occurred error :%v", topic, err.Error())
	}
	return err
}

func (sc *StanClient) Sub(ctx context.Context, topic string, handler pubsub.SubHandler) {

	_, err := sc.conn.Subscribe(topic, func(msg *stan.Msg) {
		sc.tracer.Sub(ctx, topic, handler, msg.Data)
	})
	if err != nil {
		log.Warn().Msgf("stan client sub %v topic occurred error :%v", topic, err.Error())
	}
}

func (sc *StanClient) QueueSub(ctx context.Context) {
	for _, handler := range sc.handlerQueues {
		_, err := sc.conn.QueueSubscribe(
			handler.Subject,
			handler.Group,
			func(msg *stan.Msg) {
				sc.tracer.Sub(ctx, handler.Subject, handler.Handler, msg.Data)
			},
			handler.Options...,
		)
		if err != nil {
			log.Warn().Msgf("stan client queue subscribe %v topic %v group error :%v", handler.Subject, handler.Subject, err.Error())
		}
	}
}

func (sc *StanClient) DrainCon() {
	err := sc.conn.NatsConn().Drain()
	if err != nil {
		log.Warn().Msgf("stan client drain connect error :%v", err.Error())
	}
}
