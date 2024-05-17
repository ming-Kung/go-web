package template_demo

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"html/template"
	"testing"
)

//基本语法分类：
//1、传入变量渲染模版
//* 结构体
//* map
//* 切片
//* 字符串
//2、传入函数渲染模版
//3、range在模版中的使用
//4、if-elseif-else在模版中的使用
//5、pipeline在模版中的使用

// 传入结构体渲染模版
func TestTemplateStruct(t *testing.T) {
	type User struct {
		Name string
	}
	//创建一个模版实例，传入模版名字
	tpl := template.New("hello-world-struct")
	//预编译模版，传入的参数是模版的具体内容
	tpl, err := tpl.Parse(`Hello,{{.Name}}`)
	require.NoError(t, err)
	buffer := &bytes.Buffer{}
	//传入数据(结构体)，渲染模版
	err = tpl.Execute(buffer, User{Name: "Gm"})
	require.NoError(t, err)
	assert.Equal(t, "Hello,Gm", buffer.String())
}

// 传入Map渲染模版
func TestTemplateMap(t *testing.T) {
	tpl := template.New("hello-world-map")
	tpl, err := tpl.Parse(`Hello,{{.Name}}`)
	require.NoError(t, err)
	buffer := &bytes.Buffer{}
	//传入数据（map），渲染模版
	err = tpl.Execute(buffer, map[string]string{"Name": "Gm"})
	require.NoError(t, err)
	assert.Equal(t, "Hello,Gm", buffer.String())
}

// 传入切片渲染模版
func TestTemplateSlice(t *testing.T) {
	tpl := template.New("hello-world-slice")
	tpl, err := tpl.Parse(`Hello,{{index . 0}}`)
	require.NoError(t, err)
	buffer := &bytes.Buffer{}
	//传入数据（切片），渲染模版
	err = tpl.Execute(buffer, []string{"Gm"})
	require.NoError(t, err)
	assert.Equal(t, "Hello,Gm", buffer.String())
}

// 传入字符串渲染模版
func TestTemplateString(t *testing.T) {
	tpl := template.New("hello-world-base")
	tpl, err := tpl.Parse(`Hello,{{.}}`)
	require.NoError(t, err)
	buffer := &bytes.Buffer{}
	//传入数据（字符串），渲染模版
	err = tpl.Execute(buffer, "Gm")
	require.NoError(t, err)
	assert.Equal(t, "Hello,Gm", buffer.String())
}

// 传入结构体，通过结构体里的函数渲染模版
func TestTemplateFuncCall(t *testing.T) {
	tpl := template.New("hello-world")
	tpl, err := tpl.Parse(`
切片长度:{{len .Slice}}
Hello,{{.Hello "Tom" "Jerry"}}
打印数字:{{printf "%.2f" 3.141}}
`)
	require.NoError(t, err)
	buffer := &bytes.Buffer{}
	//传入结构体，通过结构体里的函数渲染模版
	err = tpl.Execute(buffer, FuncCall{
		Slice: []string{"a", "b"},
	})
	require.NoError(t, err)
	assert.Equal(t, `
切片长度:2
Hello,Tom·Jerry
打印数字:3.14
`, buffer.String())
}

type FuncCall struct {
	Slice []string
}

func (f FuncCall) Hello(first, last string) string {
	return fmt.Sprintf("%s·%s", first, last)
}

// range在模版中的用法
// {{- }} 表示去除空行
func TestRangeLoop(t *testing.T) {
	tpl := template.New("hello-world")
	tpl, err := tpl.Parse(`
{{- range $idx,$elem := .Slice}}
{{- .}}
{{$idx}}-{{$elem}}
{{end}}
`)
	require.NoError(t, err)
	buffer := &bytes.Buffer{}
	//传入结构体，通过结构体里的函数渲染模版
	err = tpl.Execute(buffer, FuncCall{
		Slice: []string{"a", "b"},
	})
	require.NoError(t, err)
	assert.Equal(t, `a
0-a
b
1-b

`, buffer.String())
}

// if-else在模版中的用法
// 一样采用if-else 或者if-else if的结构
// 可以使用and:and 条件1 条件2
// 可以使用or:or 条件1 条件2
// 可以使用not:not 条件1
func TestIfElse(t *testing.T) {
	type User struct {
		Age int
	}
	//用一点小技巧来实现 for i 循环
	tpl := template.New("hello-world")
	tpl, err := tpl.Parse(`
{{- if and (gt .Age 0) (le .Age 6)}}
儿童 (0,6]
{{else if and (gt .Age 6) (le .Age 18)}}
少年 (6,18]
{{else}}
成人 > 18
{{end -}}
`)
	assert.Nil(t, err)
	buffer := &bytes.Buffer{}
	err = tpl.Execute(buffer, User{Age: 17})
	assert.Nil(t, err)
	assert.Equal(t, `
少年 (6,18]
`, buffer.String())
}
