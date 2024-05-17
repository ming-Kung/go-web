package web

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"sync"
)

type Context struct {
	Req  *http.Request
	Resp http.ResponseWriter

	//正则路由匹配和参数路由匹配对应的参数存这
	PathParams map[string]string

	//url里查询参数缓存
	queryValues url.Values

	//匹配的路由
	MatchedRoute string

	//模版引擎
	tplEngine TemplateEngine

	//业务方法的响应状态码与响应数据。这个主要是为了给middleware读写用的
	RespData       []byte
	RespStatusCode int
}

// RespJSON 将结构体转为json输出
func (c *Context) RespJSON(code int, val any) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	c.RespData = data
	c.RespStatusCode = code
	//c.Resp.WriteHeader(code)
	//_, err = c.Resp.Write(data)
	return err
}

func (c *Context) RespJSONOK(val any) error {
	return c.RespJSON(http.StatusOK, val)
}

// Render 模版渲染
func (c *Context) Render(tplName string, data any) error {
	var err error
	c.RespData, err = c.tplEngine.Render(c.Req.Context(), tplName, data)
	if err != nil {
		c.RespStatusCode = http.StatusInternalServerError
		return err
	}
	return nil
}

func (c *Context) SetCookie(ck *http.Cookie) {
	http.SetCookie(c.Resp, ck)
}

// BindJSON 将输入body里的json转化绑定到结构体上
func (c *Context) BindJSON(val any) error {
	if val == nil {
		return errors.New("web：输入不能为nil")
	}
	if c.Req.Body == nil {
		return errors.New("web：body 为 nil")
	}

	decoder := json.NewDecoder(c.Req.Body)
	//结构体里数字用Number来表示，否则默认是float64
	//decoder.UseNumber()
	//如果json里相对于结构体里多了未知的字段，就会报错
	//decoder.DisallowUnknownFields()
	return decoder.Decode(val)
}

// FormValue 该方法form表单参数、url参数都可获取
func (c *Context) FormValue(key string) StringValue {
	err := c.Req.ParseForm()
	if err != nil {
		return StringValue{
			str: "",
			err: err,
		}
	}
	/*vals, ok := c.Req.Form[key]
	if !ok {
		return "", errors.New("web：key not found")
	}
	return vals[0],nil*/
	return StringValue{
		str: c.Req.FormValue(key),
		err: nil,
	}
}

// QueryValue 获取url参数
func (c *Context) QueryValue(key string) StringValue {
	//这种写法用户无法区别key是否有值，还是值恰好是空字符串
	//return c.Req.URL.Query().Get(key),nil

	if c.queryValues == nil {
		//因为c.Req.URL.Query()每次都要去解析url参数，所以该处加一个url参数缓存
		c.queryValues = c.Req.URL.Query()
	}
	vals, ok := c.queryValues[key]
	if !ok {
		return StringValue{
			str: "",
			err: errors.New("web：找不到这个key"),
		}
	}
	return StringValue{
		str: vals[0],
		err: nil,
	}
}

// PathValue 获取路径参数
func (c *Context) PathValue(key string) StringValue {
	val, ok := c.PathParams[key]
	if !ok {
		return StringValue{
			str: "",
			err: errors.New("web：key 不存在"),
		}
	}
	return StringValue{
		str: val,
		err: nil,
	}
}

type StringValue struct {
	str string
	err error
}

func (s StringValue) Int64() (int64, error) {
	if s.err != nil {
		return 0, s.err
	}
	return strconv.ParseInt(s.str, 10, 64)
}

func (s StringValue) Int32() (int32, error) {
	if s.err != nil {
		return 0, s.err
	}
	int, err := strconv.Atoi(s.str)
	if err != nil {
		return 0, err
	}
	return int32(int), nil
}

// SafeContext 加了锁的线程安全的Context
type SafeContext struct {
	Context
	mutex sync.RWMutex
}

func (s *SafeContext) RespJSON(code int, val any) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	s.RespStatusCode = code
	s.RespData = data
	//s.Resp.WriteHeader(code)
	//_, err = s.Resp.Write(data)
	return err
}
