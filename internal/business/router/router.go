package router

import (
	"strconv"

	"goim-example/internal/business/service"
	"goim-example/internal/logic"
	"goim-example/internal/logic/conf"

	"github.com/gin-gonic/gin"
)

var contextErrCodeKey = "context/err/code"

type Router struct {
	c     *conf.Config
	logic *logic.Logic
	svc   *service.Service
}

func New(c *conf.Config, s *service.Service, l *logic.Logic) *Router {
	r := &Router{
		c:     c,
		logic: l,
		svc:   s,
	}
	return r
}

// ///////////////////////////////////////////////////////
// OutData 响应结构体
type OutData struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func (s *Router) OutJson(c *gin.Context, code int, msg string, data interface{}) {
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.Set(contextErrCodeKey, code)
	c.JSON(200, OutData{
		Code: code,
		Msg:  msg,
		Data: data,
	})
	c.Abort()
	return
}

// ///////////////////////////////////////////////////////
type Pages struct {
	MinId    int         `json:"from,omitempty"` // 最小id，默认0，降序分页时传入
	MaxId    int         `json:"to,omitempty"`   // 最大id，默认0，升序分页时传入
	LastPage int         `json:"last_page,omitempty"`
	Total    int64       `json:"total,omitempty"` // total
	Limit    int         `json:"limit,omitempty"` // 每页20条
	Page     int         `json:"page,omitempty"`  // 当前页
	Data     interface{} `json:"data,omitempty"`
}

func NewPage(c *gin.Context) *Pages {
	param := new(Pages)
	minId, _ := strconv.Atoi(c.Query("from"))
	maxId, _ := strconv.Atoi(c.Query("to"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	page, _ := strconv.Atoi(c.Query("page"))

	param.MinId = minId
	param.MaxId = maxId
	param.Limit = limit
	param.Page = page
	//param.Total, _ := strconv.ParseInt(c.Query("total"), 10, 64)

	if param.Limit <= 0 || param.Limit >= 100 {
		param.Limit = 15
	}
	if param.Page <= 0 || param.Page >= 10000 {
		param.Page = 1
	}

	return param
}

// ///////////////////////////////////////////////////////
type OutDataPage struct {
	//［结构体变量名 ｜ 变量类型 ｜ json 数据 对应字段名]
	Code int         `json:"code"` //接口响应状态码
	Msg  string      `json:"msg"`  //接口响应信息
	Data interface{} `json:"data"`

	MinId    int   `json:"from,omitempty"` // 最小id，默认0，降序分页时传入
	MaxId    int   `json:"to,omitempty"`   // 最大id，默认0，升序分页时传入
	LastPage int   `json:"last_page,omitempty"`
	Total    int64 `json:"total"`                  // total
	Limit    int   `json:"per_page,omitempty"`     // 每页20条
	Page     int   `json:"current_page,omitempty"` // 当前页
}

func (s *Router) OutPageJson(c *gin.Context, data interface{}, p *Pages) {
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.Set(contextErrCodeKey, 200)
	dst := &OutDataPage{
		Code:     200,
		Msg:      "success",
		Data:     data,
		MinId:    p.MinId,
		MaxId:    p.MaxId,
		LastPage: p.LastPage,
		Total:    p.Total,
		Limit:    p.Limit,
		Page:     p.Page,
	}
	c.JSON(200, dst)
	c.Abort()
}
