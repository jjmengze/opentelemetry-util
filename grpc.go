package opentelemetry

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

	"google.golang.org/grpc"
)

func GRPCServerOptions(option ...otelgrpc.Option) []grpc.ServerOption {
	return []grpc.ServerOption{grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor(option...)), grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor(option...))}
}

func GRPCClientOptions(option ...otelgrpc.Option) []grpc.DialOption {
	return []grpc.DialOption{grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor(option...)), grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor(option...))}
}
