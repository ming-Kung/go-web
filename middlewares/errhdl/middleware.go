package errhdl

import "web"

type MiddlewareErrHdlBuilder struct {
	//这种设计只能返回固定的值，不能做到动态渲染
	resp map[int][]byte
}

func NewErrHdlBuilder(errcodeMap map[int][]byte) *MiddlewareErrHdlBuilder {
	return &MiddlewareErrHdlBuilder{
		resp: errcodeMap,
	}
}

func (b *MiddlewareErrHdlBuilder) Build() web.Middleware {
	return func(next web.HandleFunc) web.HandleFunc {
		return func(ctx *web.Context) {
			next(ctx)
			resp, ok := b.resp[ctx.RespStatusCode]
			if ok {
				ctx.RespData = resp
			}
		}
	}
}

func (b *MiddlewareErrHdlBuilder) AddCode(code int, data []byte) *MiddlewareErrHdlBuilder {
	b.resp[code] = data
	return b
}
