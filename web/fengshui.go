package web

import (
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nGPU/bot/aiModule"
	"github.com/nGPU/bot/configure"
	"github.com/nGPU/bot/db"
	"github.com/nGPU/bot/header"
	log4plus "github.com/nGPU/common/log4go"
)

const (
	FengshuiPostDataError = header.FengshuiBase + 1
)

type FengshuiWeb struct {
	roots   *x509.CertPool
	rootPEM []byte
}

var gFengshuiWeb *FengshuiWeb

func getUrlExtension(urlString string) string {
	funName := "getUrlExtension"
	parsedURL, err := url.Parse(urlString)
	if err != nil {
		log4plus.Error("%s Failed err=%s", funName, err.Error())
		return ""
	}
	fileName := path.Base(parsedURL.Path)
	fileExtension := strings.TrimPrefix(path.Ext(fileName), ".")
	log4plus.Info("%s fileExtension=%s", funName, fileExtension)
	return fileExtension
}

func downloadFile(url string) (err error, localPath, imageType string) {
	funName := "downloadFile"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s url=%s consumption time=%d(s)", funName, url, time.Now().Unix()-now)
	}()

	response, err := http.Get(url)
	if err != nil {
		log4plus.Error("%s Get Failed err=%s", funName, err.Error())
		return err, "", ""
	}
	defer response.Body.Close()

	imageType = response.Header.Get("Content-Type")
	fileExt := getUrlExtension(url)
	fileName := fmt.Sprintf("%s%s.%s", time.Now().Format("20060102150405"), fmt.Sprintf("%06d", time.Now().Nanosecond()/1e3), fileExt)
	localPath = fmt.Sprintf("%s/%s", configure.SingtonConfigure().Resource.ResourcePath, fileName)
	file, err := os.Create(localPath)
	if err != nil {
		log4plus.Error("%s Create Failed err=%s", funName, err.Error())
		return err, "", ""
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		log4plus.Error("%s Copy Failed err=%s", funName, err.Error())
		return err, "", ""
	}
	return nil, localPath, imageType
}

func (w *FengshuiWeb) fengshui(c *gin.Context) {
	funName := "fengshui"
	now := time.Now().Unix()
	clientIp := getClientIP(c)
	defer func() {
		log4plus.Info("%s clientIp=[%s] consumption time=%d(s)", funName, clientIp, time.Now().Unix()-now)
	}()

	apiKey := c.GetHeader("x-api-key")
	if strings.Trim(apiKey, " ") == "" {
		errString := fmt.Sprintf("Please include your API key in the header")
		log4plus.Error(errString)
		sendError(c, header.CheckApiKeyError, errString)
		return
	}
	log4plus.Info("%s x-api-key=[%s]", funName, apiKey)

	if strings.ToLower(apiKey) != "123456" {
		err, spaceTime := db.SingtonUserDB().CheckApiKey(apiKey)
		if err != nil {
			errString := fmt.Sprintf("%s CheckApiKey failed err=[%s]", funName, err.Error())
			log4plus.Error(errString)
			sendError(c, header.CheckApiKeyError, errString)
			return
		}
		log4plus.Info("%s CheckApiKey spaceTime=[%d]", funName, spaceTime)

		if spaceTime <= 0 {
			errString := fmt.Sprintf("The duration of the API key has expired. Please renew it")
			log4plus.Error(errString)
			sendError(c, header.ApiKeyTimeoutError, errString)
			return
		}
	}

	request := &header.RequestFengshui{}
	if err := c.BindJSON(&request); err != nil {
		errString := fmt.Sprintf("%s BindJSON failed err=[%s]", funName, err.Error())
		log4plus.Error(errString)
		sendError(c, header.JsonParseError, errString)
		return
	}
	log4plus.Info("%s request=[%v]", funName, request)

	err, localFilePath, imageType := downloadFile(request.ImageUrl)
	if err != nil {
		errString := fmt.Sprintf("%s downloadFile localFilePath=[%s] imageType=[%s]", funName, localFilePath, imageType)
		log4plus.Info("%s errString=[%s]", funName, errString)
		sendError(c, header.ApiKeyTimeoutError, errString)
		return
	}
	log4plus.Info("%s downloadFile localFilePath=[%s] imageType=[%s]", funName, localFilePath, imageType)
	err, txt := aiModule.SingtonAnthropic().FengShui(localFilePath, imageType, request.Prompt)

	var result map[string]interface{}
	result = make(map[string]interface{})
	result["result_code"] = 101
	result["msg"] = "success"
	result["result_size"] = len(txt)
	result["task_dursion"] = time.Now().Unix() - now
	result["data"] = txt
	c.JSON(http.StatusOK, result)
}

func (w *FengshuiWeb) Start(userGroup *gin.RouterGroup) {
	userGroup.POST("/fengshui", w.fengshui)
}

func SingtonFengshuiWeb() *FengshuiWeb {
	if gFengshuiWeb == nil {
		gFengshuiWeb = &FengshuiWeb{}
	}
	return gFengshuiWeb
}
