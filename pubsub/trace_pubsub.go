package pubsub

import (
	"context"
	"encoding/json"
	"github.com/nats-io/stan.go"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
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

func NewPubSub() TracePubSub {
	return &pubSub{}
}

type PubHandler func() []byte

type MsgMeta struct {
	stan.Msg
	propagation.HeaderCarrier
}

func (ps *pubSub) Pub(ctx context.Context, topic string, message []byte) PubHandler {
	return func() []byte {
		data := &MsgMeta{}
		data.HeaderCarrier = make(map[string][]string)
		b, err := json.Marshal(message)
		if err != nil {
			log.Warn().Msgf("Publish marshal data %v error :%v", data, err.Error())
		}
		data.Data = b
		otel.GetTextMapPropagator().Inject(ctx, data)

		passB, err := json.Marshal(data)
		if err != nil {
			log.Warn().Msgf("Publish marshal data %v error :%v", data, err.Error())
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
		log.Warn().Msgf("Sub unmarshal data %v error :%v", data, err.Error())
	}
	spanCtx := otel.GetTextMapPropagator().Extract(ctx, msgMeta)
	spanCtx, span := opentelemetry.StartSpan(spanCtx, "Subscribe:"+topic, trace.WithSpanKind(trace.SpanKindConsumer))
	defer span.End()
	handler(data)
}
