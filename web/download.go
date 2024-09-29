package web

import (
	"crypto/x509"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nGPU/bot/configure"
	log4plus "github.com/nGPU/common/log4go"
)

const (
	DownloadBase      = iota + 310
	FileNotFoundError = DownloadBase + 1
)

type DownloadWeb struct {
	roots   *x509.CertPool
	rootPEM []byte
}

var gDownloadWeb *DownloadWeb

func (w *DownloadWeb) download(c *gin.Context) {
	funName := "download"
	filename := c.Param("filename")
	log4plus.Info("%s filename=[%s]", funName, filename)
	if strings.Trim(filename, " ") == "" {
		errString := fmt.Sprintf("%s DefaultQuery failed filename is empty", funName)
		log4plus.Error(errString)
		sendError(c, FileNotFoundError, errString)
		return
	}
	filepath := fmt.Sprintf("%s/%s", configure.SingtonConfigure().Resource.ResourcePath, filename)
	// 设置HTTP响应头
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/octet-stream")
	c.File(filepath)
}
func (w *DownloadWeb) Start(userGroup *gin.RouterGroup) {
	userGroup.GET("/download/:filename", w.download)
}

func SingtonDownloadWeb() *DownloadWeb {
	if gDownloadWeb == nil {
		gDownloadWeb = &DownloadWeb{}
	}
	return gDownloadWeb
}
