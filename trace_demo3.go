package main

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"
	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/mwitkow/go-grpc-middleware"
	//"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"

	svc "./svc"
)

const (
	httpAddr = ":7000"
)

func main() {
	cfg := config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
		},
	}
	tracer, closer, _ := cfg.New(
		"service3",
		config.Logger(jaeger.StdLogger),
	)
	opentracing.SetGlobalTracer(tracer)
	defer closer.Close()

	RPCServer()
}

type Server struct {
}

func (s *Server) Ping(ctx context.Context, req *svc.PingRequest) (*svc.PingReply, error) {
	log.Printf("%#v\n", ctx)
	infoCtx(ctx, "This is actual Ping")
	return &svc.PingReply{Status: "ok"}, nil
}

func RPCServer() {
	listen, err := net.Listen("tcp", httpAddr)
	if err != nil {
		log.Fatalf("[server] failed to listen: %v", err)
	}

	t := grpc_opentracing.UnaryServerInterceptor()
	i := grpc_middleware.WithUnaryServerChain([]grpc.UnaryServerInterceptor{t}...)
	// tracer := opentracing.GlobalTracer()
	// i := grpc.UnaryInterceptor(otgrpc.OpenTracingServerInterceptor(tracer))
	srv := grpc.NewServer(i)
	bs := &Server{}
	svc.RegisterSVCServer(srv, bs)
	srv.Serve(listen)
}

func infoCtx(ctx context.Context, msg string) {
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		span.LogFields(otlog.String("info", msg))
	}

	log.Println(msg)
}
