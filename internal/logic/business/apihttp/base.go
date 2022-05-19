package apihttp

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// OutData 响应结构体
type OutData struct {
	Code    int         `json:"code"`
	Msg     string      `json:"msg"`
	Success bool        `json:"success"`
	Result  interface{} `json:"result"`
}

func OutJson(c *gin.Context, dst OutData) {
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.JSON(200, dst)
	c.Abort()
	return
}

type Pages struct {
	MinId    int         `json:"from,omitempty"` // 最小id，默认0，降序分页时传入
	MaxId    int         `json:"to,omitempty"`   // 最大id，默认0，升序分页时传入
	LastPage int         `json:"last_page,omitempty"`
	Total    int64       `json:"total,omitempty"` // total
	Limit    int         `json:"limit,omitempty"` // 每页20条
	Page     int         `json:"page,omitempty"`  // 当前页
	List     interface{} `json:"list"`
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

func OutPageJson(c *gin.Context, data interface{}, p *Pages) {
	c.Header("Content-Type", "application/json; charset=utf-8")
	p.List = data
	dst := &OutData{
		Success: true,
		Code:    200,
		Msg:     "success",
		Result:  p,
	}
	c.JSON(200, dst)
	c.Abort()
}
