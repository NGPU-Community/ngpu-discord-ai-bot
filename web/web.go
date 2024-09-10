package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	log4plus "github.com/nGPU/common/log4go"
	"github.com/nGPU/discordBot/configure"
	"github.com/nGPU/discordBot/db"
	"github.com/nGPU/discordBot/header"
	"github.com/nGPU/discordBot/middleware"
)

type Web struct {
	webGin *gin.Engine
}

var gWeb *Web

func sendError(c *gin.Context, returnCode int, returnMsg string) {
	c.JSON(http.StatusOK, gin.H{
		"result_code": returnCode,
		"msg":         returnMsg,
	})
}

func getClientIP(c *gin.Context) string {
	reqIP := c.ClientIP()
	if reqIP == "::1" {
		reqIP = "127.0.0.1"
	}
	return reqIP
}

func (w *Web) getTask(c *gin.Context) {
	funName := "getTask"
	now := time.Now().Unix()
	clientIp := getClientIP(c)
	defer func() {
		log4plus.Info("%s clientIp=[%s] consumption time=%d(s)", funName, clientIp, time.Now().Unix()-now)
	}()

	taskID := c.DefaultQuery("taskID", "")
	log4plus.Info("%s taskID=[%s]", funName, taskID)
	if strings.Trim(taskID, " ") == "" {
		errString := fmt.Sprintf("%s DefaultQuery failed taskID is empty", funName)
		log4plus.Error(errString)
		sendError(c, header.ParamsIsNilError, errString)
		return
	}

	err, task := db.SingtonAPITasksDB().GetTaskId(taskID)
	if err != nil {
		errString := fmt.Sprintf("%s Error getting TaskID taskID=[%s] err=[%s]", funName, taskID, err.Error())
		log4plus.Error(errString)
		sendError(c, header.ParamsIsNilError, errString)
		return
	}
	if err != nil {
		errString := fmt.Sprintf("%s taskID not found in database taskId=[%s]", funName, taskID)
		log4plus.Error(errString)
		sendError(c, header.TaskNotFoundError, errString)
		return
	}

	if task.State == header.InitState {
		resultString := fmt.Sprintf("The task corresponding to TaskID is being created taskId=%s", taskID)
		sendError(c, http.StatusOK, resultString)
	} else if task.State == header.RunningState || task.State == header.IntermediateState {

		var result map[string]interface{}
		result = make(map[string]interface{})
		result["result_code"] = 101
		result["msg"] = "success"
		result["result_size"] = len(task.Response)
		result["task_dursion"] = task.RecordDursion

		var innerData map[string]interface{}
		innerData = make(map[string]interface{})
		err = json.Unmarshal([]byte(task.Response), &innerData)
		result["data"] = innerData
		c.JSON(http.StatusOK, result)

	} else if task.State == header.FinishState {

		var result map[string]interface{}
		result = make(map[string]interface{})

		result["result_code"] = 200
		result["msg"] = "success"
		result["result_size"] = len(task.Response)
		result["task_dursion"] = task.RecordDursion

		var innerData map[string]interface{}
		innerData = make(map[string]interface{})
		err = json.Unmarshal([]byte(task.Response), &innerData)
		result["data"] = innerData
		c.JSON(http.StatusOK, result)

	} else if task.State == header.ErrorState {
		var result map[string]interface{}
		result = make(map[string]interface{})
		result["result_code"] = 200
		result["msg"] = "success"
		result["result_size"] = len(task.Response)
		result["task_dursion"] = task.RecordDursion

		var innerData map[string]interface{}
		innerData = make(map[string]interface{})
		err = json.Unmarshal([]byte(task.Response), &innerData)
		result["data"] = innerData
		c.JSON(http.StatusOK, result)
	}
}

func (w *Web) start() {

	gWeb.webGin.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	userGroup := w.webGin.Group("/user")
	{
		userGroup.GET("/getTask", w.getTask)
		SingtonFaceFusionWeb().Start(userGroup)
		SingtonStableDiffustionWeb().Start(userGroup)
		SingtonBlipWeb().Start(userGroup)
		SingtonLlmWeb().Start(userGroup)
		SingtonDownloadWeb().Start(userGroup)
		SingtonRemoveWeb().Start(userGroup)
		SingtonSadTalkerWeb().Start(userGroup)
		SingtonFengshuiWeb().Start(userGroup)
	}
	log4plus.Info("start Run Listen=[%s]", configure.SingtonConfigure().Resource.Listen)
	if err := w.webGin.Run(configure.SingtonConfigure().Resource.Listen); err != nil {
		log4plus.Error("start Run Failed Not Use Http Error=[%s]", err.Error())
		return
	}
}

func SingtonWeb() *Web {
	if gWeb == nil {
		gWeb = &Web{}
		log4plus.Info("Create Web Manager")
		gWeb.webGin = gin.Default()
		gWeb.webGin.Use(middleware.Cors())
		gin.SetMode(gin.DebugMode)
		go gWeb.start()
	}
	return gWeb
}
