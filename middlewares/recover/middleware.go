package recover

import (
	"fmt"
	"log"
	"web"
)

type MiddlewareRecoverBuilder struct {
	StatusCode int
	Data       []byte
	Log        func(ctx *web.Context, str any)
}

func NewRecoverBuilder() *MiddlewareRecoverBuilder {
	return &MiddlewareRecoverBuilder{
		StatusCode: 500,
		Data:       []byte("你 panic 了"),
		Log: func(ctx *web.Context, str any) {
			log.Println(fmt.Sprintf("panic 路径：%s，panic 内容:%v", ctx.Req.URL.String(), str))
		},
	}
}

func (b *MiddlewareRecoverBuilder) Build() web.Middleware {
	return func(next web.HandleFunc) web.HandleFunc {
		return func(ctx *web.Context) {
			defer func() {
				if err := recover(); err != nil {
					ctx.RespData = b.Data
					ctx.RespStatusCode = b.StatusCode
					b.Log(ctx, err)
				}
			}()
			next(ctx)
		}
	}
}
