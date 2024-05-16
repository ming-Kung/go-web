package opentelemetry

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"log"
	"os"
	"testing"
	"time"
	"web"
)

func TestMiddlewareTraceBuilder_Build(t *testing.T) {
	//生成tracer的middleware
	tracer := otel.GetTracerProvider().Tracer(instrumentationName)
	builder := NewTraceBuilder(tracer).Build()
	//将tracer的middleware注入进server中
	server := web.NewHTTPServer(builder)

	server.Get("/user", func(ctx *web.Context) {
		c, span := tracer.Start(ctx.Req.Context(), "first_layer1")
		defer span.End()

		cs, second := tracer.Start(c, "second_layer")
		time.Sleep(time.Second)
		c, third1 := tracer.Start(cs, "third_layer_1")
		third1.End()
		time.Sleep(100 * time.Millisecond)
		c, third2 := tracer.Start(cs, "third_layer_2")
		time.Sleep(300 * time.Millisecond)
		third2.End()
		second.End()

		c, first := tracer.Start(ctx.Req.Context(), "first_layer_2")
		defer first.End()

		ctx.RespStatusCode = 200
		ctx.RespData = []byte("hello,world")
	})

	//docker启动zipkin服务后,访问localhost:9411就能进入zipkin界面
	initZipkin(t)
	server.Start(":8080")
}

func initZipkin(t *testing.T) {
	// 创建 Zipkin 导出器
	exporter, err := zipkin.New(
		"http://localhost:9411/api/v2/spans",
		zipkin.WithLogger(log.New(os.Stderr, "opentelemetry-demo", log.Ldate)),
	)
	if err != nil {
		t.Fatal(err)
	}

	batcher := sdktrace.NewBatchSpanProcessor(exporter)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(batcher),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("opentelemetry-demo"),
		)),
	)
	otel.SetTracerProvider(tp)
}
