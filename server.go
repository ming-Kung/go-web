package web

import (
	"fmt"
	"net"
	"net/http"
	"strings"
)

// 这样写的目的是为了确保一定实现了Server接口，否则编译器会报错
var _ Server = &HTTPServer{}

type HandleFunc func(ctx *Context)
type Server interface {
	http.Handler
	//Start 启动服务器
	Start(addr string) error
	//AddRoute 路由注册功能(method：http方法，path：路由，handleFunc：业务逻辑)
	//不返回error类型的原因是不想业务层再次判断error是否为nil，再次单独去处理，如果有error，内部直接panic
	addRoute(method string, path string, handleFunc HandleFunc)
	//AddRoutes(method string, path string, handleFunc ...HandleFunc)
}

type HTTPServer struct {
	*router

	//在Server层面支持middleware
	mdls []Middleware

	//log
	log func(msg string, args ...any)
}

func NewHTTPServer(mils ...Middleware) *HTTPServer {
	return &HTTPServer{
		router: newRouter(),
		mdls:   mils,
		log: func(msg string, args ...any) {
			fmt.Printf(msg, args...)
		},
	}
}

// ServeHTTP是我们整个WEB框架的核心入口，是处理请求的入口。作为http包与Web框架的关联点。
// 将在整个方法内部完成：context构建、路由匹配、执行业务逻辑
func (h *HTTPServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	//request.Body最重要的特征就是只能读取一次，无法重复读取。本质上，request.Body是一种接近stream的设计
	//request.GetBody原则上可以多次读取，但是在原声的http.Request里面，这个是nil。所以有一些Web框架会在收到请求之后，第一件事就是给GetBody赋值
	//在读取到body之后，我们就可以用于反序列化，比如说将json格式的字符串转化为一个对象等
	//request.URL.Query()
	//request.GetBody
	//request.Body
	//request.Header
	//request.ParseForm()
	//request.Form
	//request.FormValue()

	//框架代码就在这里
	ctx := &Context{
		Req:        request,
		Resp:       writer,
		PathParams: map[string]string{},
	}
	//h.Serve(ctx)

	//这里就是利用最后一个不断往前回溯组装链条。从后往前，把后一个作为前一个的next构造好链条
	root := h.Serve
	for i := len(h.mdls) - 1; i >= 0; i-- {
		root = h.mdls[i](root)
	}

	//把ctx.RespData和RespStatusCode刷新到响应里
	var m Middleware = func(next HandleFunc) HandleFunc {
		return func(ctx *Context) {
			next(ctx)
			h.flashResp(ctx)
		}
	}
	root = m(root)

	//这里执行的时候，就是从前往后了
	root(ctx)
	//h.flashResp(ctx)
}

func (h *HTTPServer) flashResp(ctx *Context) {
	if ctx.RespStatusCode != 0 {
		ctx.Resp.WriteHeader(ctx.RespStatusCode)
	}
	n, err := ctx.Resp.Write(ctx.RespData)
	if err != nil || n != len(ctx.RespData) {
		h.log("响应数据写入失败：%v", err)
	}
}

func (h *HTTPServer) Serve(ctx *Context) {
	//查找路由，并且执行命中的业务逻辑
	n, ok := h.findRoute(ctx.Req.Method, ctx.Req.URL.Path)
	if !ok || n.handler == nil {
		//路由没有命中，就是404
		ctx.RespStatusCode = 404
		ctx.RespData = []byte("NOT FOUND")
		//ctx.Resp.WriteHeader(404)
		//ctx.Resp.Write([]byte("NOT FOUND"))
		return
	}

	//获取参数路由的参数
	if n.pattern[0] == ':' {
		segs := strings.Split(ctx.Req.URL.Path, "/")
		ctx.PathParams[n.paramName] = segs[len(segs)-1]
	}
	//获取命中的路由
	ctx.MatchedRoute = n.path

	n.handler(ctx)
}

// Start 将端口监听和服务器启动分离，比直接http.ListenAndServe()有更大灵活性
func (h *HTTPServer) Start(addr string) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	//在这里，可以让用户注册所谓的after start回调
	//比如说往你的admin注册一下自己这个实例
	//在这里执行一些你业务所需的前置条件

	return http.Serve(l, h)
}

func (h *HTTPServer) Start1(addr string) error {
	return http.ListenAndServe(addr, h)
}

func (h *HTTPServer) Get(path string, handleFunc HandleFunc) {
	h.addRoute(http.MethodGet, path, handleFunc)
}
func (h *HTTPServer) Post(path string, handleFunc HandleFunc) {
	h.addRoute(http.MethodPost, path, handleFunc)
}
