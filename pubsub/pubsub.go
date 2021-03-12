package pubsub

import (
	"context"
	"encoding/json"
)

type SubHandler func(json.RawMessage) error

type Puber interface {
	Pub(ctx context.Context, topic string, message []byte) error
}

type Suber interface {
	Sub(ctx context.Context, topic string, handler SubHandler)
}
type PubSuber interface {
	Puber
	Suber
}
