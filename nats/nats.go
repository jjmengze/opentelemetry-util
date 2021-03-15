package nats

import (
	"github.com/cenk/backoff"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
	"time"
)

// NatsClient Client 訂閱方
type NatsClient struct {
	conn *nats.Conn
}

func NewNatsClient(endpoint string) *NatsClient {
	con := NewNatsClientWithOption(endpoint)

	return &NatsClient{
		conn: con,
	}
}

func NewNatsClientWithOption(endpoint string, opt ...nats.Option) *nats.Conn {
	con, err := NewNatsConn(endpoint, opt...)
	if err != nil {
		log.Warn().Msgf("New nats connection error :%v", err.Error())
	}

	return con
}

func NewNatsConn(endpoint string, natsOpt ...nats.Option) (*nats.Conn, error) {
	natsOpt = append(natsOpt, nats.ClosedHandler(func(_ *nats.Conn) {
		//if nats close connection it ,the function will be triggered
		//waitGroup.Done()
	}))

	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = 3 * time.Minute
	var con *nats.Conn
	err := backoff.Retry(func() error {
		var err error
		con, err = nats.Connect(endpoint, natsOpt...)
		if err != nil {
			log.Warn().Msgf("Nats client connect nats server %v error :%v", endpoint, err.Error())
			return err
		}
		return nil
	}, bo)

	return con, err
}

func (sc *NatsClient) Pub(topic string, data interface{}) {
	log.Warn().Msgf("Nats publish not implement")
}

func (sc *NatsClient) Sub(topic string) {
	log.Warn().Msgf("Nats sub not implement")
}
