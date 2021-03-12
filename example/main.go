package main

import (
	"context"
	"github.com/nats-io/stan.go"
	"opentelemetry-util/nats"
)

func main() {
	//var tps pubsub.TracePubSub
	//tps = pubsub.NewPubSub()

	var qSuber nats.QueueSuber
	qSuber = nats.NewStanClient("testClusterID", "test", "", nats.NewNatsClient("0.0.0.0"))
	qSuber.QueueSub(context.Background())

	nats.NewStanClientWithOption("testClusterID", "test", "", nats.NewNatsClient("0.0.0.0"),
		nats.WithStanOptions(stan.MaxPubAcksInflight(1)),
	)
}
