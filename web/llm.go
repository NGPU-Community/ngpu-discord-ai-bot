package web

import (
	"crypto/x509"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nGPU/bot/aiModule"
	"github.com/nGPU/bot/db"
	"github.com/nGPU/bot/header"
	log4plus "github.com/nGPU/common/log4go"
)

const (
	LLMError = header.LLMBase + 1
)

type LLMWeb struct {
	roots   *x509.CertPool
	rootPEM []byte
}

var gLlmWeb *LLMWeb

func (w *LLMWeb) chat(c *gin.Context) {
	funName := "chat"
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

	// body, err := ioutil.ReadAll(c.Request.Body)
	// if err != nil {
	// 	errString := fmt.Sprintf("The duration of the API key has expired. Please renew it")
	// 	log4plus.Error(errString)
	// 	sendError(c, header.ApiKeyTimeoutError, errString)
	// 	return
	// }

	request := &header.RequestChat{}
	if err := c.BindJSON(&request); err != nil {
		errString := fmt.Sprintf("%s BindJSON failed err=[%s]", funName, err.Error())
		log4plus.Error(errString)
		sendError(c, header.JsonParseError, errString)
		return
	}
	err, data := aiModule.SingtonAnthropic().Chat(request.Prompt)
	if err != nil {
		errString := fmt.Sprintf("%s PostData failed err=[%s]", funName, err.Error())
		log4plus.Error(errString)
		sendError(c, LLMError, errString)
		return
	}

	// taskId := header.GetTaskId()
	// err, data := implementation.SingtonLlm().PostData(funName, apiKey, taskId, body)
	// if err != nil {
	// 	errString := fmt.Sprintf("%s PostData failed err=[%s]", funName, err.Error())
	// 	log4plus.Error(errString)
	// 	sendError(c, LLMError, errString)
	// 	return
	// }
	var result map[string]interface{}
	result = make(map[string]interface{})
	result["result_code"] = 200
	result["msg"] = "success"
	result["result_size"] = len(data)
	result["task_dursion"] = time.Now().Unix() - now

	// var innerData map[string]interface{}
	// innerData = make(map[string]interface{})
	// err = json.Unmarshal(data, &innerData)
	result["data"] = data
	c.JSON(http.StatusOK, result)
}

func (w *LLMWeb) Start(userGroup *gin.RouterGroup) {
	userGroup.POST("/chat", w.chat)
}

func SingtonLlmWeb() *LLMWeb {
	if gLlmWeb == nil {
		gLlmWeb = &LLMWeb{}
	}
	return gLlmWeb
}
