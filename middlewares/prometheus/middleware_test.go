package prometheus

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"math/rand"
	"net/http"
	"testing"
	"time"
	"web"
)

func TestMiddlewareMetricsBuilder_Build(t *testing.T) {
	builder := NewMetricBuilder("gm_metric", "web", "http_response", "").Build()
	server := web.NewHTTPServer(builder)
	server.Get("/user", func(ctx *web.Context) {
		val := rand.Intn(1000) + 1
		time.Sleep(time.Duration(val) * time.Millisecond)
		ctx.RespStatusCode = 202
		ctx.RespData = []byte("hello，metrics")
	})

	go func() {
		//通过访问localhost:8082/metrics就能看到上传到prometheus的指标数据
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":8082", nil)
	}()

	server.Start(":8081")
}
