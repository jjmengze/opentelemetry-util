package opentelemetry

import (
	"context"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv"
	"go.opentelemetry.io/otel/trace"
)

// RegisterTraceExporter 註冊 opentelemetry trace exporter
func RegisterTraceExporter(endpoint, serviceName string, isEnable bool) (flush func(), err error) {
	// 先一律開啟
	// 目前只有dev 有設置，有需要時再開唷
	if !isEnable {
		return func() {}, nil
	}
	tp, flush, err := jaeger.NewExportPipeline(
		jaeger.WithAgentEndpoint(endpoint), //for udp
		//jaeger.WithCollectorEndpoint("http://localhost:14268/api/traces"), for tcp
		jaeger.WithProcessFromEnv(),
		jaeger.WithSDKOptions(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithResource(resource.NewWithAttributes(
				semconv.ServiceNameKey.String(serviceName),
			)),
		),
	)
	if err != nil {
		//log.Errorf("Failed to create the Open telemetry exporter: %v", err)
		return flush, err
	}

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return flush, nil
}

func NewContextWithSpan(ctx context.Context) context.Context {
	newCtx := context.WithValue(context.Background(), echo.HeaderXRequestID, ctx.Value(echo.HeaderXRequestID))
	parentSpan := trace.SpanFromContext(ctx)
	return trace.ContextWithSpan(newCtx, parentSpan)
}

func StartSpan(ctx context.Context, spName string, opts ...trace.SpanOption) (context.Context, trace.Span) {
	return otel.GetTracerProvider().Tracer("").Start(
		ctx,
		spName,
		opts...,
	)
}
