package implementation

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"errors"
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

type LLM struct {
	roots        *x509.CertPool
	rootPEM      []byte
	commandLines []*header.CommandLine
}

var gLLM *LLM

func (a *LLM) PostData(method, apiKey string, taskId string, chat []byte) (error, json.RawMessage) {
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

func (a *LLM) postData(url string, data []byte) (error, string) {
	funName := "postData"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s consumption time=%d(s)", funName, time.Now().Unix()-now)
	}()

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

	log4plus.Info("%s data=[%s]", funName, string(data))
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		errString := fmt.Sprintf("%s http.NewRequest Failed url=[%s] err=[%s]", funName, url, err.Error())
		log4plus.Error(errString)
		return err, errString
	}
	request.Header.Add("Authorization", BlipWorkSpaceId)

	response, err := client.Do(request)
	if err != nil {
		errString := fmt.Sprintf("%s client.Do Failed url=[%s] err=[%s]", funName, url, err.Error())
		log4plus.Error(errString)
		return err, errString
	}
	defer response.Body.Close()

	log4plus.Info("%s url=[%s]", funName, url)
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		errString := fmt.Sprintf("%s ioutil.ReadAll Failed url=[%s] err=[%s]", funName, url, err.Error())
		log4plus.Error(errString)
		return err, errString
	}
	log4plus.Info("%s Check StatusCode response.StatusCode=[%d] responseBody=[%s]", funName, response.StatusCode, string(responseBody))
	if response.StatusCode != 200 {
		errString := fmt.Sprintf("%s client.Do url=[%s] response.StatusCode=[%d] responseBody=[%s]", funName, url, response.StatusCode, string(responseBody))
		log4plus.Error(errString)
		return errors.New(errString), errString
	}

	type ResponseResult struct {
		ResultCode interface{} `json:"result_code"`
	}
	var result ResponseResult
	if err = json.Unmarshal(responseBody, &result); err != nil {
		errString := fmt.Sprintf("%s Unmarshal url=[%s] responseBody=[%s]", funName, url, string(responseBody))
		log4plus.Error(errString)
		return err, errString
	}
	if int(result.ResultCode.(float64)) == 200 {
		type Data struct {
			Status string `json:"status"`
			Output string `json:"output"`
		}

		type Response struct {
			ResultCode  int    `json:"result_code"`
			Msg         string `json:"msg"`
			ResultSize  int    `json:"result_size"`
			TaskDursion int    `json:"task_dursion"`
			Data        Data   `json:"data"`
		}
		var response Response
		err := json.Unmarshal(responseBody, &response)
		if err != nil {
			errString := fmt.Sprintf("%s Unmarshal url=[%s] responseBody=[%s]", funName, url, string(responseBody))
			log4plus.Error(errString)
			return err, errString
		}
		return nil, response.Data.Output
	} else {
		errString := fmt.Sprintf("%s result result_code=[%d]", funName, int(result.ResultCode.(float64)))
		log4plus.Error(errString)
		return errors.New(errString), string(responseBody)
	}
}

func (a *LLM) Chat(method, apiKey string, data []byte) (error, string) {
	funName := "Chat"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s consumption time=%d(s)", funName, time.Now().Unix()-now)
	}()
	url := configure.SingtonConfigure().Interfaces.LLm.Urls.LLM.MethodUrl
	requestTime := time.Now()

	//提交生成音频文件
	err, resultString := a.postData(url, data)
	if err != nil {
		errString := fmt.Sprintf("%s postData Failed err=[%s]", funName, err.Error())
		log4plus.Error(errString)
		return err, errString
	}
	log4plus.Info("%s postData ResultString=[%s]", funName, resultString)

	taskId := header.GetTaskId()
	db.SingtonAPITasksDB().InsertAiTask(taskId, apiKey, string(data), fmt.Sprintf("%s", requestTime.Format("2006-01-02 15:04:05")), url, method)

	responseTime := time.Now()
	db.SingtonAPITasksDB().SetAiTaskRunning(taskId, resultString, fmt.Sprintf("%s", responseTime.Format("2006-01-02 15:04:05")), int(responseTime.Unix()-requestTime.Unix()))
	return nil, resultString
}

func SingtonLlm() *LLM {
	if nil == gLLM {
		gLLM = &LLM{}
	}
	return gLLM
}
