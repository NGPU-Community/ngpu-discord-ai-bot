package implementation

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/nGPU/bot/configure"
	"github.com/nGPU/bot/db"
	"github.com/nGPU/bot/header"
	log4plus "github.com/nGPU/common/log4go"
)

type Fengshui struct {
	roots   *x509.CertPool
	rootPEM []byte
	// store        header.DiscordPluginStore
	commandLines []*header.CommandLine
}

var gFengshui *Fengshui

func (a *Fengshui) PostData(method, apiKey string, taskId string, chat []byte) (error, json.RawMessage) {
	funName := "PostData"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s consumption time=%d(s)", funName, time.Now().Unix()-now)
	}()

	//写入db
	url := fmt.Sprintf("%s", configure.SingtonConfigure().Interfaces.LLm.Urls.LLM)
	log4plus.Info("%s parse Url=[%s]", funName, url)

	requestTime := time.Now()
	db.SingtonAPITasksDB().InsertAiTask(taskId, apiKey, string(chat), fmt.Sprintf("%s", requestTime.Format("2006-01-02 15:04:05")), url, method)

	client := &http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				c, err := net.DialTimeout(netw, addr, time.Minute*5)
				if err != nil {
					log4plus.Error("%s dail timeout err=[%s]", funName, err.Error())
					return nil, err
				}
				return c, nil
			},
			MaxIdleConnsPerHost:   10,
			ResponseHeaderTimeout: time.Minute * 5,
		},
	}
	defer client.CloseIdleConnections()

	// jsonData, err := json.Marshal(chat)
	// if err != nil {
	// 	errString := fmt.Sprintf("%s json.Marshal Failed err=[%s]", funName, err.Error())
	// 	log4plus.Error(errString)
	// 	db.SingtonAPITasksDB().SetAiTaskFail(taskId, errString, fmt.Sprintf("%s", time.Now().Format("2006-01-02 15:04:05")))
	// 	return err, json.RawMessage{}
	// }
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(chat))
	if err != nil {
		errString := fmt.Sprintf("%s http.NewRequest Failed url=[%s] err=[%s]", funName, url, err.Error())
		log4plus.Error(errString)
		db.SingtonAPITasksDB().SetAiTaskFail(taskId, errString, fmt.Sprintf("%s", time.Now().Format("2006-01-02 15:04:05")))
		return err, json.RawMessage{}
	}
	response, err := client.Do(request)
	if err != nil {
		errString := fmt.Sprintf("%s client.Do Failed url=[%s] err=[%s]", funName, url, err.Error())
		log4plus.Error(errString)
		db.SingtonAPITasksDB().SetAiTaskFail(taskId, errString, fmt.Sprintf("%s", time.Now().Format("2006-01-02 15:04:05")))
		return err, json.RawMessage{}
	}
	defer response.Body.Close()

	log4plus.Info("%s url=[%s] client.Do Result=[%+v]", funName, url, response)
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		errString := fmt.Sprintf("%s ioutil.ReadAll Failed url=[%s] err=[%s]", funName, url, err.Error())
		log4plus.Error(errString)
		db.SingtonAPITasksDB().SetAiTaskFail(taskId, errString, fmt.Sprintf("%s", time.Now().Format("2006-01-02 15:04:05")))
		return err, json.RawMessage{}
	}

	log4plus.Info("%s Check StatusCode response.StatusCode=[%d] responseBody=[%s]", funName, response.StatusCode, string(responseBody))
	if response.StatusCode != 200 {
		errString := fmt.Sprintf("%s client.Do url=[%s] response.StatusCode=[%d] responseBody=[%s]", funName, url, response.StatusCode, string(responseBody))
		log4plus.Error(errString)
		db.SingtonAPITasksDB().SetAiTaskFail(taskId, errString, fmt.Sprintf("%s", time.Now().Format("2006-01-02 15:04:05")))
		return err, json.RawMessage{}
	}

	type ResponseResult struct {
		TaskId     interface{} `json:"task_id"`
		ResultCode interface{} `json:"result_code"`
	}
	var result ResponseResult
	if err = json.Unmarshal(responseBody, &result); err != nil {
		errString := fmt.Sprintf("%s Unmarshal Failed url=[%s] err=[%s]", funName, url, err.Error())
		log4plus.Error(errString)
		db.SingtonAPITasksDB().SetAiTaskFail(taskId, errString, fmt.Sprintf("%s", time.Now().Format("2006-01-02 15:04:05")))
		return err, json.RawMessage{}
	}
	if result.ResultCode == nil {
		errString := fmt.Sprintf("%s result.ResultCode is null", funName)
		log4plus.Error(errString)
		db.SingtonAPITasksDB().SetAiTaskFail(taskId, errString, fmt.Sprintf("%s", time.Now().Format("2006-01-02 15:04:05")))
		return err, json.RawMessage{}
	}
	if int(result.ResultCode.(float64)) == 0 {
		log4plus.Info(fmt.Sprintf("%s int(result.ResultCode.(float64)) == 0", funName))
		responseTime := time.Now()
		db.SingtonAPITasksDB().SetAiTaskSuccess(taskId, string(responseBody), fmt.Sprintf("%s", responseTime.Format("2006-01-02 15:04:05")), int(responseTime.Unix()-requestTime.Unix()))
		return nil, json.RawMessage(responseBody)
	} else {
		errString := fmt.Sprintf("%s int(result.ResultCode.(float64))=[%d]", funName, int(result.ResultCode.(float64)))
		log4plus.Error(errString)
		responseTime := time.Now()
		db.SingtonAPITasksDB().SetAiTaskFail(taskId, string(responseBody), fmt.Sprintf("%s", responseTime.Format("2006-01-02 15:04:05")))
		return nil, json.RawMessage(responseBody)
	}
}

func SingtonFengshui() *Fengshui {
	if nil == gFengshui {
		gFengshui = &Fengshui{}
	}
	return gFengshui
}
