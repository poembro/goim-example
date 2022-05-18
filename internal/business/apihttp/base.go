package apihttp

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

// OutData 响应结构体
type OutData struct {
	Code    int         `json:"code"`
	Msg     string      `json:"msg"`
	Success bool        `json:"success"`
	Result  interface{} `json:"result"`
}

func OutJson(w http.ResponseWriter, dst OutData) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	err := json.NewEncoder(w).Encode(dst)
	if err != nil {
		w.Write([]byte(`{"code":-1, "success":false, "msg":"解析JSON出错"}`))
	}
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

func NewPage(r *http.Request) *Pages {
	param := new(Pages)
	query := r.URL.Query()

	minId, _ := strconv.Atoi(query.Get("from"))
	maxId, _ := strconv.Atoi(query.Get("to"))
	limit, _ := strconv.Atoi(query.Get("limit"))
	page, _ := strconv.Atoi(query.Get("page"))

	param.MinId = minId
	param.MaxId = maxId
	param.Limit = limit
	param.Page = page
	//param.Total, _ := strconv.ParseInt(query.Get("total"), 10, 64)

	if param.Limit <= 0 || param.Limit >= 100 {
		param.Limit = 15
	}
	if param.Page <= 0 || param.Page >= 10000 {
		param.Page = 1
	}

	return param
}

func OutPageJson(w http.ResponseWriter, data interface{}, p *Pages) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	p.List = data

	dst := &OutData{
		Success: true,
		Code:    200,
		Msg:     "success",
		Result:  p,
	}
	err := json.NewEncoder(w).Encode(dst)
	if err != nil {
		log.Println(err)
	}

	return err
}
