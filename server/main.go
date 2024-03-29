/*
DISCLAIMER: Parts of this code are referencing following sources:
- https://github.com/grpc-ecosystem/go-grpc-middleware/tree/main/examples
*/

package main

import (
	"context"
	"fmt"
	stdlog "log"
	"net"
	"os"

	kitlog "github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/go-yaml/yaml"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/selector"
	pb "github.com/louisloechel/cloudservicebenchmarking/pb"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

const (
	component = "grpc-component"
	port      = "0.0.0.0:50051"
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

// interceptorLogger adapts go-kit logger to interceptor logger.
// This code is simple enough to be copied and not imported.
func interceptorLogger(l kitlog.Logger) logging.Logger {
	return logging.LoggerFunc(func(_ context.Context, lvl logging.Level, msg string, fields ...any) {
		largs := append([]any{"msg", msg}, fields...)
		switch lvl {
		case logging.LevelDebug:
			_ = level.Debug(l).Log(largs...)
		case logging.LevelInfo:
			_ = level.Info(l).Log(largs...)
		case logging.LevelWarn:
			_ = level.Warn(l).Log(largs...)
		case logging.LevelError:
			_ = level.Error(l).Log(largs...)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}

// No-Op Interceptor is needed in case you want to chain interceptors but don't want to add any additional logic.
// Fixes returning nil as an interceptor which causes the chain to break.
func noOpInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Just call the handler directly without any additional logic
	return handler(ctx, req)
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		stdlog.Fatalf("failed to listen: %v", err)
	}

	// Set up logging
	logger := kitlog.NewLogfmtLogger(os.Stderr)
	rpcLogger := kitlog.With(logger, "service", "gRPC/server", "component", component)
	logTraceID := func(ctx context.Context) logging.Fields {
		if span := trace.SpanContextFromContext(ctx); span.IsSampled() {
			return logging.Fields{"traceID", span.TraceID().String()}
		}
		return nil
	}

	// Prometheus metrics
	srvMetrics := grpcprom.NewServerMetrics(
		grpcprom.WithServerHandlingTimeHistogram(
			grpcprom.WithHistogramBuckets([]float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120}),
		),
	)
	reg := prometheus.NewRegistry()
	reg.MustRegister(srvMetrics)
	exemplarFromContext := func(ctx context.Context) prometheus.Labels {
		if span := trace.SpanContextFromContext(ctx); span.IsSampled() {
			return prometheus.Labels{"traceID": span.TraceID().String()}
		}
		return nil
	}

	// Custom auth
	authFn := func(ctx context.Context) (context.Context, error) {
		token, err := auth.AuthFromMD(ctx, "bearer")
		if err != nil {
			return nil, err
		}
		// TODO: This is example only, perform proper Oauth/OIDC verification!
		if token != "yolo" {
			return nil, status.Error(codes.Unauthenticated, "invalid auth token")
		}
		// NOTE: You can also pass the token in the context for further interceptors or gRPC service code.
		return ctx, nil
	}

	// Setup auth matcher.
	allButHealthZ := func(ctx context.Context, callMeta interceptors.CallMeta) bool {
		return healthpb.Health_ServiceDesc.ServiceName != callMeta.Service
	}

	/* ---------------------------------------------------------
	| This is the core of the gRPC interceptor chain experiment |
	-----------------------------------------------------------*/

	type Config struct {
		Logging struct {
			Enabled bool `yaml:"enabled"`
		} `yaml:"logging"`
		Auth struct {
			Enabled bool `yaml:"enabled"`
		} `yaml:"auth"`
		Metrics struct {
			Enabled bool `yaml:"enabled"`
		} `yaml:"metrics"`
		Telemetry struct {
			Enabled bool `yaml:"enabled"`
		} `yaml:"telemetry"`
	}

	// Load the config.yml file
	configFile := "./config.yml"
	configData, err := os.ReadFile(configFile)
	if err != nil {
		stdlog.Fatalf("failed to read config file: %v", err)
	}
	stdlog.Printf("Loaded config file: %v", configFile)
	stdlog.Printf("Config data: %v", string(configData))

	// Parse the YAML data into the Config struct
	var config Config
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		stdlog.Fatalf("failed to parse config file: %v", err)
	}

	// Apply the boolean values from the config
	enableMetrics := config.Metrics.Enabled
	enableLogging := config.Logging.Enabled
	enableAuth := config.Auth.Enabled
	enableTelemetry := config.Telemetry.Enabled

	stdlog.Printf("Metrics enabled: %v", enableMetrics)
	stdlog.Printf("Logging enabled: %v", enableLogging)
	stdlog.Printf("Auth enabled: %v", enableAuth)

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			// Conditionally include the metrics interceptor
			func() grpc.UnaryServerInterceptor {
				if enableMetrics {
					return srvMetrics.UnaryServerInterceptor(grpcprom.WithExemplarFromContext(exemplarFromContext))
				}
				return noOpInterceptor
			}(),
			// Conditionally include the logging interceptor
			func() grpc.UnaryServerInterceptor {
				if enableLogging {
					return logging.UnaryServerInterceptor(interceptorLogger(rpcLogger), logging.WithFieldsFromContext(logTraceID))
				}
				return noOpInterceptor
			}(),
			// Conditionally include the auth interceptor
			func() grpc.UnaryServerInterceptor {
				if enableAuth {
					return selector.UnaryServerInterceptor(auth.UnaryServerInterceptor(authFn), selector.MatchFunc(allButHealthZ))
				}
				return noOpInterceptor
			}(),
			// Conditionally include the telemetry interceptor
			func() grpc.UnaryServerInterceptor {
				if enableTelemetry {
					return otelgrpc.UnaryServerInterceptor()
				}
				return noOpInterceptor
			}(),
		)),
	}

	s := grpc.NewServer(opts...)
	pb.RegisterGreeterServer(s, &server{})
	stdlog.Printf("Server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		stdlog.Fatalf("failed to serve: %v", err)
	}
}
