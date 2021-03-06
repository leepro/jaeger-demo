package main

import (
	"context"
	"flag"
	"fmt"
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
	httpAddr = ":5000"
)

var JAEGER_AGENT = flag.String("j", "10.0.0.33:5775", "jaeger agent")

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
		"service1",
		config.Logger(jaeger.StdLogger),
	)
	opentracing.SetGlobalTracer(tracer)
	defer closer.Close()

	http.Handle("/hello", http.HandlerFunc(HelloHandler))
	http.Handle("/yyy", http.HandlerFunc(YYYHandler))

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

func YYYHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	infoCtx(ctx, "calling helloResponse")
	tagCtx(ctx, "func", "YYYHandler")
	errorCtx(ctx, fmt.Errorf("this is error"))
	w.Write([]byte("ok"))
}

func helloResponse(ctx context.Context, name string) (string, error) {
	infoCtx(ctx, "serving helloResponse")
	tagCtx(ctx, "func", "helloResponse")
	// errorCtx(ctx, fmt.Erroorf("this is error %s", name))

	client := &http.Client{Transport: &nethttp.Transport{}}
	req, err := http.NewRequest("GET", "http://localhost:6000/xxx?age=20", nil)
	if err != nil {
		return "", err
	}
	req = req.WithContext(ctx)
	// ct:= req.Context()
	// tagCtx(ct, "func", "good!")
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

func tagCtx(ctx context.Context, key, tag string) {
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		span.SetTag(key, tag)
	}
}

func errorCtx(ctx context.Context, err error) {
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		span.LogFields(otlog.Error(err))
	}
	log.Println(err.Error())
}

func infoCtx(ctx context.Context, msg string) {
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		span.LogFields(otlog.String("info", msg))
	}
	log.Println(msg)
}
