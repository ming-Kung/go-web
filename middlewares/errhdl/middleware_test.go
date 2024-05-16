package errhdl

import (
	"net/http"
	"testing"
	"web"
)

func TestMiddlewareErrHdlBuilder_Build(t *testing.T) {
	builder := NewErrHdlBuilder(map[int][]byte{}).
		AddCode(http.StatusNotFound, []byte(`
<html>
	<body>
		<h1>链接找不到</h1>
	</body>
</html>
`)).
		AddCode(http.StatusBadRequest, []byte(`
<html>
	<body>
		<h1>请求不对</h1>
	</body>
</html>
`)).Build()
	server := web.NewHTTPServer(builder)

	server.Get("/user", func(ctx *web.Context) {
		ctx.RespStatusCode = 200
	})
	server.Start(":8081")
}
