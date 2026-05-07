package grpc

import (
	"context"
	"log"

	"google.golang.org/grpc"
)

func UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	log.Printf("gRPC request: %s", info.FullMethod)
	return handler(ctx, req)
}
