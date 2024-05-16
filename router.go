package web

import (
	"fmt"
	"regexp"
	"strings"
)

type nodeType int

const (
	//静态路由
	nodeTypeStatic = iota
	//正则路由
	nodeTypeReg
	//路径参数路由
	nodeTypeParam
	//通配符路由
	nodeTypeAny
)

// 用来支持对路由树的操作
// 路由树（森林）
type router struct {
	//Beego Gin HTTP method对应一棵树
	//GET有一棵树，POST也有一棵树

	//http method => 路由树根节点
	trees map[string]*node
}
type node struct {
	//节点名
	pattern string

	//节点类型
	typ int

	//静态匹配节点
	children map[string]*node //子 path 到子节点的映射

	//通配符匹配节点
	startChild *node

	//参数路由
	paramChild *node

	//正则表达式
	regChild *node
	regExpr  *regexp.Regexp

	//正则路由和参数路由都会使用这个字段
	paramName string

	//具体业务方法
	handler HandleFunc

	//完整路由
	path string
}

func (n *node) childOrCreate(seg string) *node {
	//以 : 开头，需要进一步解析，判断是参数路由还是正则路由
	if seg[0] == ':' {
		paramName, expr, isReg := n.parseParam(seg)
		if isReg {
			//获取或创建正则匹配路由下一个节点
			return n.childOrCreateReg(seg, expr, paramName)
		}
		//获取或创建参数匹配路由下一个节点
		return n.childOrCreateParam(seg, paramName)
	}
	if seg == "*" {
		//获取或创建通配符匹配路由下一个节点
		return n.childOrCreateAny(seg)
	}

	//获取或创建静态路由下一个节点
	if n.children == nil {
		n.children = map[string]*node{}
	}
	res, ok := n.children[seg]
	if !ok {
		//要新建节点
		res = &node{
			pattern: seg,
			typ:     nodeTypeStatic,
		}
		n.children[seg] = res
	}
	return res
}

// childOf 返回子节点
// 匹配优先级：静态匹配 > 正则路由匹配 > 参数路径匹配 > 通配符匹配
func (n *node) childOf(seg string) (*node, bool) {
	if seg == "" {
		return n, true
	}
	if n.children == nil {
		return n.childOfNonStatic(seg)
	}
	children, ok := n.children[seg]
	if !ok {
		return n.childOfNonStatic(seg)
	}
	return children, true
}

// parseParam 用于解析是不是正则表达式(要求正则表达式的格式为： :paramName(xxx))
// 第一个返回值是参数名字
// 第二个返回值是正则表达式
// 第三个返回值为true则说明是正则路由
func (n *node) parseParam(seg string) (string, string, bool) { //假设传参seg为 :paramName(xxx)
	//去除: ，seg变为paramName(xxx)
	seg = seg[1:]
	//将seg切割为最多两段，paramName 、 xxx)
	segs := strings.SplitN(seg, "(", 2)
	if len(segs) == 2 {
		expr := segs[1]
		if strings.HasSuffix(expr, ")") { //检查字符串expr是否是以')'结尾
			return segs[0], expr[:len(expr)-1], true
		}
	}
	return seg, "", false
}

// childOrCreateReg 获取或创建正则匹配子节点
func (n *node) childOrCreateReg(seg string, expr string, paramName string) *node {
	if n.paramChild != nil {
		panic("web：非法路由，已有路径参数。不允许同时注册路径参数路由和正则匹配路由")
	}
	if n.startChild != nil {
		panic("web：非法路由，已有通配符匹配。不允许同时注册通配符匹配路由和正则匹配路由")
	}
	if n.regChild != nil {
		//相同位置不允许重复注册正则路由
		if n.regChild.regExpr == nil || n.regChild.regExpr.String() != expr || n.regChild.paramName != paramName {
			panic(fmt.Sprintf("web：路由冲突，正则路由冲突，已有 %s，新注册 %s", n.regChild.pattern, seg))
		}
	} else {
		regExpr, err := regexp.Compile(expr)
		if err != nil {
			panic(fmt.Errorf("web：正则表达式错误 %v", err))
		}
		n.regChild = &node{
			pattern:   seg,
			paramName: paramName,
			regExpr:   regExpr,
			typ:       nodeTypeReg,
		}
	}
	return n.regChild
}

// childOrCreateParam 获取或创建参数匹配子节点
func (n *node) childOrCreateParam(seg string, paramName string) *node {
	if n.regChild != nil {
		panic("web：非法路由，已有正则匹配。不允许同时注册路径参数和正则匹配")
	}
	if n.startChild != nil {
		panic("web：非法路由，已有通配符匹配。不允许同时注册路径参数和通配符匹配")
	}
	if n.paramChild == nil {
		n.paramChild = &node{
			pattern:   seg,
			paramName: paramName,
			typ:       nodeTypeParam,
		}
	} else {
		if n.paramChild.pattern != seg {
			panic(fmt.Sprintf("web：路由冲突，路径参数冲突，不允许注册相同位置的路径参数，已有路径参数%s，不允许再注册路径参数%s", n.paramChild.pattern, seg))
		}
	}
	return n.paramChild
}

// childOrCreateAny 获取或创建通配符匹配子节点
func (n *node) childOrCreateAny(seg string) *node {
	if n.paramChild != nil {
		panic("web：非法路由，已有路径参数。不允许同时注册路径参数和通配符匹配")
	}
	if n.regChild != nil {
		panic("web：非法路由，已有正则匹配。不允许同时注册通配符匹配和正则匹配")
	}
	if n.startChild == nil {
		n.startChild = &node{
			pattern: seg,
			typ:     nodeTypeAny,
		}
	}
	return n.startChild
}

// childOfNonStatic 查找非静态匹配子节点
func (n *node) childOfNonStatic(seg string) (*node, bool) {
	//正则匹配
	if n.regChild != nil {
		if n.regChild.regExpr.Match([]byte(seg)) {
			return n.regChild, true
		}
	}
	//参数匹配
	if n.paramChild != nil {
		return n.paramChild, true
	}
	//通配符匹配
	//通配符*在末尾时，能匹配后续所有的路径
	if n.pattern == "*" && n.startChild == nil {
		return n, true
	}
	return n.startChild, n.startChild != nil
}

func newRouter() *router {
	return &router{
		trees: map[string]*node{},
	}
}

// AddRoute 新增路由
// * 已经注册了的路由，无法被覆盖，例如/user/home注册两次，会冲突
// * path不可以为"",可以是"/"开头，也可以不是"/"开头
// * 支持通配符*在中间时，只能匹配一段，但是在末尾时，则能匹配后续所有的路径。即/a/*/c能匹配/a/b/c，不能匹配/a/b1/b2/c，/a/b/*能匹配/a/b/c/d
// * 不能在同一个位置注册不同的参数路由，例如/user/:id 和 /user/:name冲突
// * 同一个位置只能注册路径参数，通配符路由和正则路由中的一个，三者是互斥的，例如/user/:id 和 /user/*冲突
// * 同名路径参数，在路由匹配的时候，值会被覆盖，例如/user/:id/abc/:id，那么/user/123/abc/456 最终id=456
// * 路由树可以不用线程安全，因为程序启动监听之前，路由已经依次注册号，不存在并发写路由树的场景，并发读不影响
func (r *router) addRoute(method string, path string, handleFunc HandleFunc) {
	if path == "" {
		panic("web:路径不能为空字符串")
	}
	root, ok := r.trees[method]
	if !ok {
		//说明没有根节点，创建根节点
		root = &node{
			pattern: "/",
		}
		r.trees[method] = root
	}
	// "/user/home"如果不去掉第一个"/"，就会被切割成三段：""、"user"、"home"
	//path = path[1:]
	//切割path
	segs := strings.Split(path, "/")
	for _, seg := range segs {
		if seg == "" {
			continue
		}
		//递归下去，找准位置
		//如果中途有节点不存在，就要创建出来
		children := root.childOrCreate(seg)
		root = children
	}
	if root.handler != nil {
		panic(fmt.Sprintf("web:路由冲突，重复注册[%s]", path))
	}
	root.handler = handleFunc
	root.path = path
}

// findRoute 路由查找
// 这里不支持路由查找回溯
func (r *router) findRoute(method string, path string) (*node, bool) {
	if path == "" {
		return nil, false
	}
	root, ok := r.trees[method]
	if !ok {
		return nil, false
	}
	segs := strings.Split(path, "/")
	for _, seg := range segs {
		child, found := root.childOf(seg)
		if !found {
			return nil, false
		}
		root = child
	}
	return root, true
}
