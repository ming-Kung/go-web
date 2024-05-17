package web

import (
	"log"
	"testing"
)

func TestLoginPage(t *testing.T) {
	engine := &GoTemplateEngine{}
	err := engine.ParseGlobal("testdata/tpls/*.gohtml")
	if err != nil {
		log.Fatal(err)
	}
	s := NewHTTPServer()
	s.SetTemplateEngine(engine)
	s.Get("/login", func(ctx *Context) {
		err := ctx.Render("login.gohtml", nil)
		if err != nil {
			log.Println(err)
		}
	})

	s.Start(":8081")
}
