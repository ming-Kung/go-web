package web

import (
	"bytes"
	"html/template"
)

type TemplateEngine interface {
	//Render 渲染页面
	//tplName:模版名字，按名索引；data:渲染页面所需要的数据
	Render(tplName string, data any) ([]byte, error)
}

type GoTemplateEngine struct {
	T *template.Template
}

func (g *GoTemplateEngine) Render(tplName string, data any) ([]byte, error) {
	bs := &bytes.Buffer{}
	err := g.T.ExecuteTemplate(bs, tplName, data)
	return bs.Bytes(), err
}

func (g *GoTemplateEngine) ParseGlobal(pattern string) error {
	tpl, err := template.ParseGlob(pattern)
	g.T = tpl
	return err
}
