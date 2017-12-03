package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"time"

	"google.golang.org/grpc"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
	"github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/mwitkow/go-grpc-middleware"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"

	svc "./svc"
)

const (
	httpAddr = ":7000"
)

var JAEGER_AGENT = flag.String("j", "localhost:5775", "jaeger agent")

func main() {
	flag.Parse()
	cfg := config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
			LocalAgentHostPort:  *JAEGER_AGENT,
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

	err := FinalCall(ctx)

	return &svc.PingReply{Status: "ok"}, err
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

func FinalCall(ctx context.Context) error {
	client := &http.Client{Transport: &nethttp.Transport{}}
	req, err := http.NewRequest("GET", "http://localhost:5000/yyy?year=2017", nil)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	tracer := opentracing.GlobalTracer()
	req, ht := nethttp.TraceRequest(tracer, req)
	defer ht.Finish()

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	res.Body.Close()
	return nil
}
