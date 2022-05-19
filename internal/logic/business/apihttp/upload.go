package apihttp

import (
	"path"

	"github.com/gin-gonic/gin"
)

func (s *Router) Upload(c *gin.Context) {
	//从请求中读取文件
	file, err := c.FormFile("f1") //请求中获取携带的参数,就是html文件中的name="f1"
	if err != nil {               //读取失败，将错误报出来
		OutJson(c, OutData{Code: -1, Success: false, Msg: "上传失败"})
		return
	} else { //读取成功，就保存到服务端本地
		fileDest := path.Join("./", file.Filename)
		c.SaveUploadedFile(file, fileDest)
		OutJson(c, OutData{Code: 200, Success: true, Result: fileDest})
	}
}
