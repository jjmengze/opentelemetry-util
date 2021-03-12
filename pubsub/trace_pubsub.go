package pubsub

import (
	"context"
	"encoding/json"
	"github.com/nats-io/stan.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	opentelemetry "opentelemetry-util"
)

type TracePuber interface {
	Pub(ctx context.Context, topic string, message []byte) PubHandler
}

type TraceSuber interface {
	Sub(ctx context.Context, topic string, handler SubHandler, data json.RawMessage)
}

type TracePubSub interface {
	TracePuber
	TraceSuber
}

type pubSub struct {
}

type PubHandler func() []byte

type MsgMeta struct {
	stan.Msg
	http.Header
}

func (ps *pubSub) Pub(ctx context.Context, topic string, message []byte) PubHandler {
	return func() []byte {
		data := &MsgMeta{}
		b, err := json.Marshal(message)
		if err != nil {
			//todo err log
		}
		data.Data = b
		otel.GetTextMapPropagator().Inject(ctx, data)

		passB, err := json.Marshal(data)
		if err != nil {
			//todo err log
		}
		_, span := opentelemetry.StartSpan(ctx, "Publish:"+topic, trace.WithSpanKind(trace.SpanKindProducer))
		defer span.End()
		return passB
	}
}

func (ps *pubSub) Sub(ctx context.Context, topic string, handler SubHandler, data json.RawMessage) {
	msgMeta := &MsgMeta{}
	err := json.Unmarshal(data, msgMeta)
	if err != nil {
		//todo err log
	}
	spanCtx := otel.GetTextMapPropagator().Extract(ctx, msgMeta)
	spanCtx, span := opentelemetry.StartSpan(spanCtx, "Subscribe:"+topic, trace.WithSpanKind(trace.SpanKindConsumer))
	defer span.End()
	handler(data)
}
