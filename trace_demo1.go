package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
)

const (
	httpAddr = ":8000"
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
		"service1",
		config.Logger(jaeger.StdLogger),
	)
	opentracing.SetGlobalTracer(tracer)
	defer closer.Close()

	http.Handle("/hello", http.HandlerFunc(HelloHandler))
	http.Handle("/xxx", http.HandlerFunc(XXXHandler))

	mux := nethttp.Middleware(
		opentracing.GlobalTracer(),
		http.DefaultServeMux)

	log.Println(httpAddr)
	go http.ListenAndServe(httpAddr, mux)

	select {}
}

func HelloHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	name := r.URL.Query().Get("name")
	infoCtx(ctx, "calling helloResponse")
	resp, err := helloResponse(ctx, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, _ = w.Write([]byte(resp))
}

func XXXHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	age := r.URL.Query().Get("age")
	infoCtx(ctx, "calling xxxHandler")
	resp, err := xxxResponse(ctx, age)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, _ = w.Write([]byte(resp))
}

func helloResponse(ctx context.Context, name string) (string, error) {
	infoCtx(ctx, "serving helloResponse")
	// tag(ctx, "tag", name)
	// errorCtx(ctx, fmt.Erroorf("this is error %s", name))

	client := &http.Client{Transport: &nethttp.Transport{}}
	req, err := http.NewRequest("GET", "http://localhost:9000/xxx?age=1975", nil)
	if err != nil {
		return "", err
	}
	req = req.WithContext(ctx)
	tracer := opentracing.GlobalTracer()
	req, ht := nethttp.TraceRequest(tracer, req)
	defer ht.Finish()

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	res.Body.Close()

	return "ok", nil
}

func xxxResponse(ctx context.Context, name string) (string, error) {
	infoCtx(ctx, "this is xxxRequest")
	return "", nil
}

func infoCtx(ctx context.Context, msg string) {
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		span.LogFields(otlog.String("info", msg))
	}

	log.Println(msg)
}