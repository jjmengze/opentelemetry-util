module opentelemetry-util

go 1.15

require (
	github.com/cenk/backoff v2.2.1+incompatible
	github.com/google/uuid v1.1.2
	github.com/jinzhu/gorm v1.9.16
	github.com/jjmengze/otgorm v0.19.0
	github.com/labstack/echo/v4 v4.2.1
	github.com/nats-io/nats-streaming-server v0.21.1 // indirect
	github.com/nats-io/nats.go v1.10.0
	github.com/nats-io/stan.go v0.8.3
	github.com/rs/zerolog v1.20.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.18.0
	go.opentelemetry.io/otel v0.19.0
	go.opentelemetry.io/otel/exporters/trace/jaeger v0.19.0
	go.opentelemetry.io/otel/sdk v0.19.0
	go.opentelemetry.io/otel/trace v0.19.0
)
