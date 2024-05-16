package web

// Middleware 函数式的责任链模式、函数式的洋葱模式
type Middleware func(next HandleFunc) HandleFunc

//AOP方案在不同的框架，不同的语言里面都有不同的叫法
//Middleware、Handler、Chain、Filter、Filter-Chain、Interceptor、Wrapper
// 拦截器设置
/*type MiddlewareV1 interface {
	Invoke(next HandleFunc) HandleFunc
}
type Interceptor interface {
	Before(ctx *Context)
	After(ctx *Context)
	Surround(ctx *Context)
}*/

/*type HandleFuncV2 func(ctx *Context) (next bool)
type ChainV2 struct {
	handlers []HandleFuncV2
}
func (c ChainV2) Run(ctx *Context) {
	for _, h := range c.handlers {
		next := h(ctx)
		//中断执行
		if !next {
			return
		}
	}
}*/
