package main

import (
	"context"
	"log"
	"net"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	pb "github.com/louisloechel/cloudservicebenchmarking/pb"
	otelgrpc "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	// log.Printf("Received: %v", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

// loggingInterceptor is a simple interceptor that logs each request.
func loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	log.Printf("Request - Method:%s; %v", info.FullMethod, req)

	// Uncomment to activate client side timeout.UnaryClientInterceptor(500*time.Millisecond)
	// log.Printf("Sleeping for 5 seconds")
	// time.Sleep(5 * time.Second) // Sleep for 5 seconds

	resp, err := handler(ctx, req)
	log.Printf("Response - Method:%s; %v", info.FullMethod, resp)
	return resp, err
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Create a server option with the interceptors
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			loggingInterceptor,                // Your custom logging interceptor
			otelgrpc.UnaryServerInterceptor(), // OpenTelemetry interceptor
		)),
	}

	s := grpc.NewServer(opts...)
	pb.RegisterGreeterServer(s, &server{})
	log.Printf("Server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
