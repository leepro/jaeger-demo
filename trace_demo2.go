package main

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"net/http"
	"time"

	//	"github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	//	"github.com/mwitkow/go-grpc-middleware"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"

	svc "./svc"
)

const (
	httpAddr = ":6000"
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
		"service2",
		config.Logger(jaeger.StdLogger),
	)
	opentracing.SetGlobalTracer(tracer)
	defer closer.Close()

	http.Handle("/xxx", http.HandlerFunc(XXXHandler))

	mux := nethttp.Middleware(
		opentracing.GlobalTracer(),
		http.DefaultServeMux)

	log.Println(httpAddr)
	go http.ListenAndServe(httpAddr, mux)

	select {}
}

func XXXHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	age := r.URL.Query().Get("age")
	infoCtx(ctx, "calling XXXHandler")
	resp, err := xxxResponse(ctx, age)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, _ = w.Write([]byte(resp))
}

func xxxResponse(ctx context.Context, name string) (string, error) {
	infoCtx(ctx, "this is xxxRequest start")
	ret := ClientRPC(ctx)
	infoCtx(ctx, "this is xxxRequest end")
	return ret, nil
}

func infoCtx(ctx context.Context, msg string) {
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		span.LogFields(otlog.String("info", msg))
	}

	log.Println(msg)
}

func ClientRPC(ctx context.Context) string {
	infoCtx(ctx, "this is ClientRPC")
	defer infoCtx(ctx, "ends ClientRPCRPC")

	log.Printf("%#v\n", ctx)

	tracer := opentracing.GlobalTracer()
	opt := grpc.WithUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(tracer))
	conn, err := grpc.Dial("localhost:7000", opt, grpc.WithInsecure())
	if err != nil {
		log.Printf("Dial Error=%s\n", err.Error())
		return err.Error()
	}
	client := svc.NewSVCClient(conn)
	res, err := client.Ping(ctx, &svc.PingRequest{Status: "hey"})
	if err != nil {
		log.Println("err>>>>", err.Error())
		return err.Error()
	}
	log.Println("status=", res.Status)
	return res.Status
}
