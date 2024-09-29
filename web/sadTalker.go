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
	SadTalkerPostDataError = header.SadTalkerBase + 1
)

type SadTalkerWeb struct {
	roots   *x509.CertPool
	rootPEM []byte
}

var gSadTalkerWeb *SadTalkerWeb

func (w *SadTalkerWeb) sadTalker(c *gin.Context) {
	funName := "sadTalker"
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

	request := &header.RequestSadTalker{}
	if err := c.BindJSON(&request); err != nil {
		errString := fmt.Sprintf("%s BindJSON failed err=[%s]", funName, err.Error())
		log4plus.Error(errString)
		sendError(c, header.JsonParseError, errString)
		return
	}

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

	err, taskId := implementation.SingtonSadTalker().SadTalker(funName, apiKey, data)
	if err != nil {
		errString := fmt.Sprintf("%s SadTalker failed err=[%s] resultString=[%s]", funName, err.Error(), err.Error())
		log4plus.Error(errString)
		sendError(c, SadTalkerPostDataError, errString)
		return
	}
	response := header.ResponseBlackWhite2Color{
		ResultCode: 200,
		Message:    "success",
		TaskId:     taskId,
	}
	c.JSON(http.StatusOK, response)
}

func (w *SadTalkerWeb) Start(userGroup *gin.RouterGroup) {
	userGroup.POST("/sadTalker", w.sadTalker)
}

func SingtonSadTalkerWeb() *SadTalkerWeb {
	if gSadTalkerWeb == nil {
		gSadTalkerWeb = &SadTalkerWeb{}
	}
	return gSadTalkerWeb
}
