package apihttp

import (
	"net/http"
	"os"
)

// StaticServer 静态文件处理
func (s *Router) StaticServer(w http.ResponseWriter, req *http.Request) {
	var (
		basedir string
		indexs  = []string{"index.html", "index.htm"}
	)

	basedir, _ = os.Getwd() // 获取当前目录路径 /webser/go_wepapp/goim-demo
	//filePath := basedir + "/dist" + req.URL.Path
	filePath := basedir + "/../../dist" + req.URL.Path //注意 注意 注意:这里只能处理 dist 目录下的文件
	//fmt.Println(filePath)
	fi, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		http.NotFound(w, req)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if fi.IsDir() {
		if req.URL.Path[len(req.URL.Path)-1] != '/' {
			http.Redirect(w, req, req.URL.Path+"/", 301)
			return
		}
		for _, index := range indexs {
			fi, err = os.Stat(filePath + index)
			if err != nil {
				continue
			}
			http.ServeFile(w, req, filePath+index)
			return
		}
		http.NotFound(w, req)
		return
	}
	http.ServeFile(w, req, filePath)
}
