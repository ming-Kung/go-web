package web

import (
	"fmt"
	"net/http"
	"testing"
)

func TestServer(t *testing.T) {
	var s = NewHTTPServer()

	s.Get("/usr/detail", func(ctx *Context) {
		ctx.RespData = []byte("hello,/usr/detail")
	})
	s.Get("/usr/*", func(ctx *Context) {
		ctx.RespData = []byte("hello,/usr/*")
	})
	//框架中不支持路由回溯，这种路径不支持。如果浏览器输入http://localhost:8081/usr/home/lis，就会not found
	s.Get("/usr/*/*", func(ctx *Context) {
		ctx.RespData = []byte("hello,/usr/*/*")
	})
	s.Get("/usr/home/list", func(ctx *Context) {
		ctx.RespData = []byte("hello,/usr/home/list")
	})
	s.Get("/*/detail", func(ctx *Context) {
		ctx.RespData = []byte("hello,/*/detail")
	})
	s.Get("/usr/detail/:id", func(ctx *Context) {
		ctx.RespData = []byte(fmt.Sprintf("hello,/usr/detail/:id,id:%s", ctx.PathParams["id"]))
	})

	//因为路径同一个位置不能注册不同的路由参数，此处会panic
	/*s.Get("/usr/detail/:name", func(ctx *Context) {
		ctx.Resp.Write([]byte(fmt.Sprintf("hello,/usr/:name,name:%s", ctx.PathParams["name"])))
	})*/

	s.Get("/gm/:id([0-9a-zA-Z]+)", func(ctx *Context) {
		ctx.RespData = []byte(fmt.Sprintf("hello,%s,id:%s", "/gm/:id([0-9a-zA-Z]+)", ctx.PathParams["id"]))
	})

	s.Post("/form", func(ctx *Context) {
		ctx.RespData = []byte("hello,form")
	})

	//启动服务器监听
	//用法一：完全委托给http包管
	//这个handler就是我们跟http包的结合点
	//http.ListenAndServe(":8081", s)
	//http.ListenAndServeTLS(":443", "", "", s)

	//用法二：自己手动管
	s.Start(":8081")

}

func TestHTTPServer_ServeHTTP(t *testing.T) {
	server := NewHTTPServer()
	server.mdls = []Middleware{
		func(next HandleFunc) HandleFunc {
			return func(ctx *Context) {
				fmt.Println("第一个before")
				next(ctx)
				fmt.Println("第一个after")
			}
		},
		func(next HandleFunc) HandleFunc {
			return func(ctx *Context) {
				fmt.Println("第二个before")
				next(ctx)
				fmt.Println("第二个after")
			}
		},
		func(next HandleFunc) HandleFunc {
			return func(ctx *Context) {
				fmt.Println("第三个中断")
			}
		},
		func(next HandleFunc) HandleFunc {
			return func(ctx *Context) {
				fmt.Println("第四个，你看不到这句话")
			}
		},
	}
	server.ServeHTTP(nil, &http.Request{})
}
