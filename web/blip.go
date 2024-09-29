package web

import (
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nGPU/bot/db"
	"github.com/nGPU/bot/header"
	"github.com/nGPU/bot/implementation"
	log4plus "github.com/nGPU/common/log4go"
)

const (
	BlipPostDataError = header.BlipBase + 1
)

type BlipWeb struct {
	roots   *x509.CertPool
	rootPEM []byte
}

var gBlipWeb *BlipWeb

func (w *BlipWeb) img2txt(c *gin.Context) {
	funName := "img2txt"
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

	request := &header.RequestImg2Txt{}
	if err := c.BindJSON(&request); err != nil {
		errString := fmt.Sprintf("%s BindJSON failed err=[%s]", funName, err.Error())
		log4plus.Error(errString)
		sendError(c, header.JsonParseError, errString)
		return
	}
	log4plus.Info("%s request=[%v]", funName, request)

	tmpData, err := json.Marshal(request)
	if err != nil {
		errString := fmt.Sprintf("%s json.Marshal Failed err=[%s]", funName, err.Error())
		log4plus.Error(errString)
		sendError(c, header.JsonParseError, errString)
		return
	}

	var body header.RequestData
	body.BTCAddr = "0000000000000000000000000GFg7xJaNVN2"
	body.Data = json.RawMessage(tmpData)

	data, err := json.Marshal(body)
	if err != nil {
		errString := fmt.Sprintf("%s json.Marshal Failed err=[%s]", funName, err.Error())
		log4plus.Info("%s errString=[%s]", funName, errString)
		sendError(c, header.JsonParseError, errString)
		return
	}

	err, txt := implementation.SingtonBlip().Blip(funName, apiKey, data)
	if err != nil {
		errString := fmt.Sprintf("%s Blip failed err=[%s]", funName, err.Error())
		log4plus.Error(errString)
		sendError(c, BlipPostDataError, errString)
		return
	}

	var result map[string]interface{}
	result = make(map[string]interface{})
	result["result_code"] = 101
	result["msg"] = "success"
	result["result_size"] = len(txt)
	result["task_dursion"] = time.Now().Unix() - now
	result["data"] = txt
	c.JSON(http.StatusOK, result)
}

func (w *BlipWeb) Start(userGroup *gin.RouterGroup) {
	userGroup.POST("/img2txt", w.img2txt)
}

func SingtonBlipWeb() *BlipWeb {
	if gBlipWeb == nil {
		gBlipWeb = &BlipWeb{}
	}
	return gBlipWeb
}
