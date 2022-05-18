package apihttp

import (
	"fmt"
	"goim-demo/pkg/logger"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"
)

func (s *Router) apiUpload(c *gin.Context) {
	var (
		newPath string // 暂时只处理1个文件上传
	)
	if r.Method != "POST" {
		return
	}
	reader, err := r.MultipartReader()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	basedir, _ := os.Getwd() //获取当前目录路径 /webser/go_wepapp/goim-demo
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}

		logger.Logger.Debug("apiUpload", zap.String("FileName", part.FileName()),
			zap.String("FormName", part.FormName()))

		if part.FileName() == "" { // this is FormData
			data, _ := ioutil.ReadAll(part)
			logger.Logger.Debug("apiUpload", zap.String("data", string(data)))
		} else { // This is FileData
			newPath = fmt.Sprintf("%s/dist/upload/%d_%s", basedir, time.Now().Unix(), part.FileName())
			dst, _ := os.Create(newPath) // 写入时需要dist 访问路径上不能带有 /dist
			defer dst.Close()
			io.Copy(dst, part)
		}
	}

	if newPath == "" {
		OutJson(c, OutData{Code: -1, Success: false, Msg: "上传失败"})
		return
	}
	OutJson(c, OutData{Code: 200, Success: true, Result: newPath})
}
