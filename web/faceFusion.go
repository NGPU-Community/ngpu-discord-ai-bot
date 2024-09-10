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
	BlackWhite2ColorError = header.FaceFusionBase + 1
	LipSyncerError        = header.FaceFusionBase + 2
	FaceSwapError         = header.FaceFusionBase + 3
	FrameEnhanceError     = header.FaceFusionBase + 4
)

type FaceFusionWeb struct {
	roots   *x509.CertPool
	rootPEM []byte
}

var gFaceFusionWeb *FaceFusionWeb

func (w *FaceFusionWeb) blackWhite2Color(c *gin.Context) {
	funName := "blackWhite2Color"
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

	if strings.ToLower(apiKey) != strings.ToLower("123456") {
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

	request := &header.RequestBlackWhite2Color{}
	if err := c.BindJSON(request); err != nil {
		errString := fmt.Sprintf("%s BindJSON failed err=[%s]", funName, err.Error())
		log4plus.Error(errString)
		sendError(c, header.JsonParseError, errString)
		return
	}
	if strings.Trim(request.Fun, " ") == "" {
		request.Fun = "blackwhite2color"
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

	err, taskId := implementation.SingtonFaceFusion().BlackWhite2Color(funName, apiKey, data)
	if err != nil {
		errString := fmt.Sprintf("%s BlackWhite2Color failed err=[%s] taskId=[%s]", funName, err.Error(), taskId)
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

func (w *FaceFusionWeb) lipSyncer(c *gin.Context) {
	funName := "lipSyncer"
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

	request := &header.RequestLipSyncer{}
	if err := c.BindJSON(request); err != nil {
		errString := fmt.Sprintf("%s BindJSON failed err=[%s]", funName, err.Error())
		log4plus.Error(errString)
		sendError(c, header.JsonParseError, errString)
		return
	}
	if strings.Trim(request.Fun, " ") == "" {
		request.Fun = "lip_syncer"
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

	err, taskId := implementation.SingtonFaceFusion().LipSyncer(funName, apiKey, data)
	if err != nil {
		errString := fmt.Sprintf("%s LipSyncer failed err=[%s]", funName, err.Error())
		log4plus.Error(errString)
		sendError(c, LipSyncerError, errString)
		return
	}
	response := header.ResponseLipSyncer{
		ResultCode: 200,
		Message:    "success",
		TaskId:     taskId,
	}
	c.JSON(http.StatusOK, response)
}

func (w *FaceFusionWeb) faceSwap(c *gin.Context) {
	funName := "faceSwap"
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

	request := &header.RequestFaceSwap{}
	if err := c.BindJSON(request); err != nil {
		errString := fmt.Sprintf("%s BindJSON failed err=[%s]", funName, err.Error())
		log4plus.Error(errString)
		sendError(c, header.JsonParseError, errString)
		return
	}
	if strings.Trim(request.Fun, " ") == "" {
		request.Fun = "face_swap"
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

	err, taskId := implementation.SingtonFaceFusion().FaceSwap(funName, apiKey, data)
	if err != nil {
		errString := fmt.Sprintf("%s FaceSwap failed err=[%s] taskId=[%s]", funName, err.Error(), taskId)
		log4plus.Error(errString)
		sendError(c, FaceSwapError, errString)
		return
	}
	response := header.ResponseFaceSwap{
		ResultCode: 200,
		Message:    "success",
		TaskId:     taskId,
	}
	c.JSON(http.StatusOK, response)
}

func (w *FaceFusionWeb) frameEnhance(c *gin.Context) {
	funName := "frameEnhance"
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

	request := &header.RequestFrameEnhance{}
	if err := c.BindJSON(&request); err != nil {
		errString := fmt.Sprintf("%s BindJSON failed err=[%s]", funName, err.Error())
		log4plus.Error(errString)
		sendError(c, header.JsonParseError, errString)
		return
	}
	if strings.Trim(request.Fun, " ") == "" {
		request.Fun = "frameEnhance"
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

	err, taskId := implementation.SingtonFaceFusion().FrameEnhance(funName, apiKey, data)
	if err != nil {
		errString := fmt.Sprintf("%s FrameEnhance failed err=[%s] taskId=[%s]", funName, err.Error(), taskId)
		log4plus.Error(errString)
		sendError(c, FrameEnhanceError, errString)
		return
	}
	response := header.ResponseFaceSwap{
		ResultCode: 200,
		Message:    "success",
		TaskId:     taskId,
	}
	c.JSON(http.StatusOK, response)
}

type ResponseDataInterface struct {
	TaskId     interface{} `json:"task_id"`
	ResultCode interface{} `json:"result_code"`
}

func (w *FaceFusionWeb) Start(userGroup *gin.RouterGroup) {
	userGroup.POST("/blackWhite2Color", w.blackWhite2Color)
	userGroup.POST("/lipSyncer", w.lipSyncer)
	userGroup.POST("/faceSwap", w.faceSwap)
	userGroup.POST("/frameEnhance", w.frameEnhance)
}

func SingtonFaceFusionWeb() *FaceFusionWeb {
	if gFaceFusionWeb == nil {
		gFaceFusionWeb = &FaceFusionWeb{}
	}
	return gFaceFusionWeb
}
