package accesslog

import (
	"encoding/json"
	"log"
	"web"
)

type MiddlewareAccessLogBuilder struct {
	logFunc func(accessLog string)
}

// NewAccessLogBuilder 初始化结构体MiddlewareAccessLogBuilder中函数logFunc为打印一行字符串
func NewAccessLogBuilder() *MiddlewareAccessLogBuilder {
	return &MiddlewareAccessLogBuilder{
		logFunc: func(accessLog string) {
			log.Println(accessLog)
		},
	}
}

// LogFunc 自己通过传参来定义b.logFunc为啥具体函数
func (b *MiddlewareAccessLogBuilder) LogFunc(fn func(log string)) *MiddlewareAccessLogBuilder {
	b.logFunc = fn
	return b
}

// Build 将自定义结构体MiddlewareAccessLogBuilder转为web.Middleware
func (b *MiddlewareAccessLogBuilder) Build() web.Middleware {
	return func(next web.HandleFunc) web.HandleFunc {
		return func(ctx *web.Context) {
			//记录请求
			//在defer里输出日志原因：
			//1、确保即便next里面发生了panic，也能将请求记录下来
			//2、获得MatchedRoute，它只有在执行了next之后才能获得，因为依赖于最终的路由树匹配
			defer func() {
				l := NewAccessLog(ctx)
				data, _ := json.Marshal(l)
				b.logFunc(string(data))
			}()
			next(ctx)
		}
	}
}

type accessLog struct {
	//host
	Host string `json:"host,omitempty"`
	//命中的路由
	Route string `json:"route,omitempty"`
	//http方法
	HTTPMethod string `json:"http_method,omitempty"`
	//实际路径
	Path string `json:"path,omitempty"`
	//响应码
	RespStatusCode int `json:"resp_status_code"`
	//响应数据
	RespData string `json:"resp_data"`
}

func NewAccessLog(ctx *web.Context) accessLog {
	return accessLog{
		Host:           ctx.Req.Host,
		Route:          ctx.MatchedRoute,
		HTTPMethod:     ctx.Req.Method,
		Path:           ctx.Req.URL.Path,
		RespStatusCode: ctx.RespStatusCode,
		RespData:       string(ctx.RespData),
	}
}
