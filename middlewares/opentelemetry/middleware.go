package opentelemetry

import (
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
	"log"
	"os"
	"web"
)

const instrumentationName = "gongmingming_opentelemetry_trace" //最好用当前项目的仓库名

type MiddlewareTraceBuilder struct {
	tracer trace.Tracer
}

func NewTraceBuilder(tracer trace.Tracer) *MiddlewareTraceBuilder {
	return &MiddlewareTraceBuilder{
		tracer: tracer,
	}
}

func (b *MiddlewareTraceBuilder) Build() web.Middleware {
	if b.tracer == nil {
		b.tracer = otel.GetTracerProvider().Tracer(instrumentationName)
	}
	return func(next web.HandleFunc) web.HandleFunc {
		return func(ctx *web.Context) {
			//尝试和客户端的trace结合在一起(通过header)
			reqCtx := ctx.Req.Context()
			reqCtx = otel.GetTextMapPropagator().Extract(reqCtx, propagation.HeaderCarrier(ctx.Req.Header))

			//如果传入的context里面已经有span了，那么新创建的span就是老的span的儿子
			reqCtx, span := b.tracer.Start(reqCtx, "unknow")
			defer span.End()

			//自定义需要记录的
			span.SetAttributes(attribute.String("http.host", ctx.Req.Host))
			span.SetAttributes(attribute.String("http.url", ctx.Req.URL.String()))
			span.SetAttributes(attribute.String("http.method", ctx.Req.Method))
			span.SetAttributes(attribute.String("http.schema", ctx.Req.URL.Scheme))

			//这块非常重要，将调用业务代码前的包含tracer的context更新，串进业务的context中
			ctx.Req = ctx.Req.WithContext(reqCtx)

			next(ctx)

			//将span的名字设置为匹配的路由名，ctx.MatchedRoute只有执行完next才能有值
			span.SetName(ctx.MatchedRoute)

			//把响应码加进去
			span.SetAttributes(attribute.Int("http.status", ctx.RespStatusCode))
		}
	}
}

func InitZipkin() {
	// 创建 Zipkin 导出器
	exporter, err := zipkin.New(
		"http://localhost:9411/api/v2/spans",
		zipkin.WithLogger(log.New(os.Stderr, "opentelemetry-demo", log.Ldate)),
	)
	if err != nil {
		fmt.Println(fmt.Sprintf("tracing zipkin err:%v", err))
		return
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
