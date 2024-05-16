package web

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"reflect"
	"testing"
)

// TDD（驱动测试），编写具体方法前，先写测试用例。
// Test_router_addRoute 测试构建路由树
func Test_router_addRoute(t *testing.T) {
	//构造路由树
	//验证路由树
	testRoutes := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodGet,
			path:   "/user/home",
		},
		{
			method: http.MethodGet,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/user",
		},
		{
			method: http.MethodGet,
			path:   "/order/detail",
		},
		{
			method: http.MethodGet,
			path:   "/order/*",
		},
		{
			method: http.MethodGet,
			path:   "/order/detail",
		},
		{
			method: http.MethodPost,
			path:   "/order/create",
		},
		{
			method: http.MethodPost,
			path:   "login",
		},
	}
	mockHandler := func(ctx *Context) {}

	r := newRouter()
	for _, tr := range testRoutes {
		r.addRoute(tr.method, tr.path, mockHandler)
	}

	//在这里断言路由树和预期的一摸一样
	wantRouter := &router{
		trees: map[string]*node{
			http.MethodGet: &node{
				pattern: "/",
				handler: mockHandler,
				children: map[string]*node{
					"user": &node{
						pattern: "user",
						children: map[string]*node{
							"home": &node{
								pattern: "home",
								handler: mockHandler,
							},
						},
						handler: mockHandler,
					},
					"order": &node{
						pattern: "order",
						children: map[string]*node{
							"detail": &node{
								pattern: "detail",
								handler: mockHandler,
							},
						},
						startChild: &node{
							pattern: "*",
							handler: mockHandler,
						},
					},
				},
			},
			http.MethodPost: &node{
				pattern: "/",
				children: map[string]*node{
					"order": &node{
						pattern: "order",
						children: map[string]*node{
							"create": &node{
								pattern: "create",
								handler: mockHandler,
							},
						},
					},
					"login": &node{
						pattern: "login",
						handler: mockHandler,
					},
				},
			},
		},
	}
	//判断两者是否相等
	msg, ok := wantRouter.equal(r)
	assert.True(t, ok, msg)
}

// Test_router_findRoute 测试查找路由树
func Test_router_findRoute(t *testing.T) {
	testRoute := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodPost,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/order/detail",
		},
	}

	r := newRouter()
	var mockHandler HandleFunc = func(ctx *Context) {}
	for _, route := range testRoute {
		r.addRoute(route.method, route.path, mockHandler)
		/*n, isFind := r.findRoute(route.method, route.path)
		if isFind {
			fmt.Printf("找到路由：%s，方法：%v\n", route.path, n.handler)
		} else {
			fmt.Printf("未找到路由：%s", route.path)
		}*/
	}

	testCases := []struct {
		name      string
		method    string
		path      string
		wantFound bool
		wantNode  *node
	}{
		{
			//根节点
			name:      "root",
			method:    http.MethodPost,
			path:      "/",
			wantFound: true,
			wantNode: &node{
				pattern: "/",
				handler: mockHandler,
			},
		},
		{
			//完全命中
			name:      "order detail",
			method:    http.MethodGet,
			path:      "/order/detail",
			wantFound: true,
			wantNode: &node{
				handler: mockHandler,
				pattern: "detail",
			},
		},
	}
	for _, tc := range testCases {
		//在一个测试函数中动态的运行其他测试函数
		t.Run(tc.name, func(t *testing.T) {
			n, found := r.findRoute(tc.method, tc.path)
			assert.Equal(t, tc.wantFound, found)
			if !found {
				return
			}
			assert.Equal(t, tc.wantNode.pattern, n.pattern)
			assert.Equal(t, tc.wantNode.children, n.children)
			nHandler := reflect.ValueOf(n.handler)
			yHandler := reflect.ValueOf(tc.wantNode.handler)
			assert.True(t, nHandler == yHandler)
		})
	}

}

// 返回一个错误信息，帮助我们排查问题
// bool 是否代表真的相等
func (r *router) equal(y *router) (string, bool) {
	for k, v := range r.trees {
		dst, ok := y.trees[k]
		if !ok {
			return fmt.Sprintf("找不到对应的 http mothed"), false
		}
		//判断v和dst是否相等
		msg, equal := v.equal(dst)
		if !equal {
			return msg, false
		}
	}
	return "", true
}

func (n *node) equal(y *node) (string, bool) {
	if y.pattern != n.pattern {
		return fmt.Sprintf("节点路径不匹配"), false
	}
	if len(y.children) != len(n.children) {
		return fmt.Sprintf("子节点数量不相等"), false
	}

	if n.startChild != nil {
		msg, ok := n.startChild.equal(y.startChild)
		if !ok {
			return msg, ok
		}
	}

	if n.paramChild != nil {
		msg, ok := n.paramChild.equal(y.paramChild)
		if !ok {
			return msg, ok
		}
	}

	//比较handler
	nHandler := reflect.ValueOf(n.handler)
	yHandler := reflect.ValueOf(y.handler)
	if nHandler != yHandler {
		return fmt.Sprintf("handler 不相等"), false
	}

	for path, c := range n.children {
		dst, ok := y.children[path]
		if !ok {
			return fmt.Sprintf("子节点 %s 不存在", path), false
		}
		msg, ok := c.equal(dst)
		if !ok {
			return msg, false
		}
	}

	return "", true
}
