package recover

import (
	"testing"
	"web"
)

func TestMiddlewareRecoverBuilder_Build(t *testing.T) {
	builder := NewRecoverBuilder().Build()
	server := web.NewHTTPServer(builder)
	server.Get("/user", func(ctx *web.Context) {
		panic("发生panic 了")
	})
	server.Start(":8081")
}
