package accesslog

import (
	"testing"
	"web"
)

func TestMiddlewareBuilder(t *testing.T) {
	//builder := MiddlewareBuilder{}
	//mdl := builder.LogFunc(func(log string) {
	//	fmt.Println(log)
	//}).Build()

	mdl := NewAccessLogBuilder().Build()
	server := web.NewHTTPServer(mdl)
	server.Get("/a/b/*", func(ctx *web.Context) {
		ctx.RespStatusCode = 200
		ctx.RespData = []byte("hello,it's me")
	})

	//模拟http get请求
	//req, err := http.NewRequest(http.MethodGet, "/a/b/c", nil)
	//req.Host = "localhost:8081"
	//if err != nil {
	//	t.Fatal(err)
	//}
	//server.ServeHTTP(nil, req)

	server.Start(":8081")
}
