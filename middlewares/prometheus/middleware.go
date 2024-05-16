package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
	"time"
	"web"
)

type MiddlewareMetricsBuilder struct {
	Namespace string
	Subsystem string
	Name      string
	Help      string
}

func NewMetricBuilder(namespace, subsystem, name, help string) *MiddlewareMetricsBuilder {
	return &MiddlewareMetricsBuilder{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      name,
		Help:      help,
	}
}

func (b *MiddlewareMetricsBuilder) Build() web.Middleware {
	vector := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: b.Namespace,
		Subsystem: b.Subsystem,
		Name:      b.Name,
		Help:      b.Help,
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.90:  0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	}, []string{"pattern", "method", "status"})
	//将指标向量参数设置，注册进普罗米修斯
	prometheus.MustRegister(vector)

	return func(next web.HandleFunc) web.HandleFunc {
		return func(ctx *web.Context) {
			startTime := time.Now()
			defer func() {
				duration := time.Now().Sub(startTime).Milliseconds()
				pattern := ctx.MatchedRoute
				if pattern == "" {
					pattern = "unknow"
				}
				//将该条请求的label值与采样值，设置进普罗米修斯
				vector.WithLabelValues(pattern, ctx.Req.Method, strconv.Itoa(ctx.RespStatusCode)).Observe(float64(duration))
			}()
			next(ctx)
		}
	}
}
