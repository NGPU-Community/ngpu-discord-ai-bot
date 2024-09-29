package web

import (
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	StableDiffustionPostDataError = header.StableDiffustionBase + 1
)

type StableDiffustionWeb struct {
	roots   *x509.CertPool
	rootPEM []byte
}

var gStableDiffustionWeb *StableDiffustionWeb

func (w *StableDiffustionWeb) img2img(c *gin.Context) {
	funName := "img2img"
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

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		errString := fmt.Sprintf("The duration of the API key has expired. Please renew it")
		log4plus.Error(errString)
		sendError(c, header.ApiKeyTimeoutError, errString)
		return
	}

	taskId := header.GetTaskId()
	err, data := implementation.SingtonStableDiffusion().PostData(funName, apiKey, taskId, body)
	if err != nil {
		errString := fmt.Sprintf("%s GenerateAudio failed err=[%s]", funName, err.Error())
		log4plus.Error(errString)
		sendError(c, StableDiffustionPostDataError, errString)
		return
	}
	var result map[string]interface{}
	result = make(map[string]interface{})
	result["result_code"] = 101
	result["msg"] = "success"
	result["result_size"] = len(data)
	result["task_dursion"] = time.Now().Unix() - now

	var innerData map[string]interface{}
	innerData = make(map[string]interface{})
	err = json.Unmarshal(data, &innerData)
	result["data"] = innerData
	c.JSON(http.StatusOK, result)
}

func (w *StableDiffustionWeb) Start(userGroup *gin.RouterGroup) {
	userGroup.POST("/img2img", w.img2img)
}

func SingtonStableDiffustionWeb() *StableDiffustionWeb {
	if gStableDiffustionWeb == nil {
		gStableDiffustionWeb = &StableDiffustionWeb{}
	}
	return gStableDiffustionWeb
}
