package tracers_test

import (
	"aurora/internal/opentracing/tracers"
	"context"
	"testing"
	"time"

	"github.com/opentracing/opentracing-go"
	jaeger "github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
)

func TestJaeger(t *testing.T) {
	closer, err := tracers.SetupTracer("test-svc", "http://127.0.0.1:14268/api/traces", true)
	if err != nil {
		t.Fatal(err)
	}
	defer closer()

	span, _ := opentracing.StartSpanFromContext(context.Background(), "test")
	defer span.Finish()
}

func TestJaegerDirect(t *testing.T) {

	cfg := jaegercfg.Configuration{
		ServiceName: "client test",
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			QueueSize:           1,
			BufferFlushInterval: time.Duration(1) * time.Second,
			// LocalAgentHostPort:  "localhost:6831", // "localhost:6831" Based on UDP is not reliable
			CollectorEndpoint: "http://127.0.0.1:14268/api/traces",
			LogSpans:          true,
		},
	}

	jLogger := jaegerlog.StdLogger
	tracer, closer, err := cfg.NewTracer(
		jaegercfg.Logger(jLogger),
	)

	defer closer.Close()
	if err != nil {
	}

	// 创建第一个 span A
	parentSpan := tracer.StartSpan("A")
	defer parentSpan.Finish()

	B(tracer, parentSpan)
}

func B(tracer opentracing.Tracer, parentSpan opentracing.Span) {
	// 继承上下文关系，创建子 span
	childSpan := tracer.StartSpan(
		"B",
		opentracing.ChildOf(parentSpan.Context()),
	)
	defer childSpan.Finish()
}
