package web

import (
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	log4plus "github.com/nGPU/common/log4go"
	"github.com/nGPU/discordBot/db"
	"github.com/nGPU/discordBot/header"
	"github.com/nGPU/discordBot/implementation"
)

const (
	RemoveBGError = header.RemoveBase + 1
)

type RemoveBGWeb struct {
	roots   *x509.CertPool
	rootPEM []byte
}

var gRemoveBGWeb *RemoveBGWeb

func (w *RemoveBGWeb) removeBG(c *gin.Context) {
	funName := "removeBG"
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

	request := &header.RequestRemoveBG{}
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

	err, taskId := implementation.SingtonRemoveBG().RemoveBG(funName, apiKey, data)
	if err != nil {
		errString := fmt.Sprintf("%s RemoveBG failed err=[%s] taskId=[%s]", funName, err.Error(), taskId)
		log4plus.Error(errString)
		sendError(c, BlackWhite2ColorError, errString)
		return
	}
	response := header.ResponseBlackWhite2Color{
		ResultCode: 200,
		Message:    "success",
		TaskId:     taskId,
	}
	c.JSON(http.StatusOK, response)
}

func (w *RemoveBGWeb) replaceBG(c *gin.Context) {
	funName := "replaceBG"
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

	request := &header.RequestReplaceBG{}
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

	err, taskId := implementation.SingtonRemoveBG().ReplaceBG(funName, apiKey, data)
	if err != nil {
		errString := fmt.Sprintf("%s ReplaceBG failed err=[%s] taskId=[%s]", funName, err.Error(), taskId)
		log4plus.Error(errString)
		sendError(c, BlackWhite2ColorError, errString)
		return
	}
	response := header.ResponseBlackWhite2Color{
		ResultCode: 200,
		Message:    "success",
		TaskId:     taskId,
	}
	c.JSON(http.StatusOK, response)
}

func (w *RemoveBGWeb) Start(userGroup *gin.RouterGroup) {
	userGroup.POST("/removeBG", w.removeBG)
	userGroup.POST("/replaceBG", w.replaceBG)
}

func SingtonRemoveWeb() *RemoveBGWeb {
	if gRemoveBGWeb == nil {
		gRemoveBGWeb = &RemoveBGWeb{}
	}
	return gRemoveBGWeb
}
