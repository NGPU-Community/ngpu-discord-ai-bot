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

var RemoveBGWorkSpaceId = string("ngpu_000000000000006")

type RemoveBG struct {
	roots   *x509.CertPool
	rootPEM []byte
	// store        header.DiscordPluginStore
	commandLines []*header.CommandLine
}

var gRemoveBG *RemoveBG

func (a *RemoveBG) postData(url string, data []byte) (error, string) {
	funName := "postData"
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
	request.Header.Add("Authorization", RemoveBGWorkSpaceId)

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

	if int(result.ResultCode.(float64)) != 200 {
		errString := fmt.Sprintf("%s result.ResultCode != 200 result.ResultCode=[%d]", funName, int(result.ResultCode.(float64)))
		log4plus.Error(errString)
		return errors.New(errString), errString
	}
	return nil, string(responseBody)
}

func (a *RemoveBG) GetData(url string, taskId string) (error, string) {
	funName := "GetData"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s url=[%s] consumption time=%d(s)", funName, url, time.Now().Unix()-now)
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

	getUrl := fmt.Sprintf("%s?taskID=%s", url, taskId)
	request, err := http.NewRequest("GET", getUrl, nil)
	if err != nil {
		errString := fmt.Sprintf("%s http.NewRequest Failed url=[%s] err=[%s]", funName, url, err.Error())
		log4plus.Error(errString)
		return err, errString
	}
	response, err := client.Do(request)
	if err != nil {
		errString := fmt.Sprintf("%s client.Do Failed url=[%s] err=[%s]", funName, url, err.Error())
		log4plus.Error(errString)
		return err, errString
	}
	defer response.Body.Close()

	// log4plus.Info("%s url=[%s] client.Do Result=[%+v]", funName, url, response)
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
		return err, errString
	}
	return nil, string(responseBody)
}

func (a *RemoveBG) RemoveBG(method, apiKey string, data []byte) (error, string) {
	funName := "RemoveBG"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s consumption time=%d(s)", funName, time.Now().Unix()-now)
	}()

	url := configure.SingtonConfigure().Interfaces.RemoveBG.Urls.RemoveBG.MethodUrl
	requestTime := time.Now()

	err, resultString := a.postData(url, data)
	if err != nil {
		errString := fmt.Sprintf("%s postData Failed err=[%s]", funName, err.Error())
		log4plus.Error(errString)
		return err, errString
	}
	log4plus.Info("%s postData ResultString=[%s]", funName, resultString)

	var result ResponsePostDataInterface
	if err = json.Unmarshal([]byte(resultString), &result); err != nil {
		errString := fmt.Sprintf("%s Unmarshal Failed err=[%s]", funName, err.Error())
		log4plus.Error(errString)
		return err, errString
	}
	if result.ResultCode == nil {
		errString := fmt.Sprintf("%s result.ResultCode is nil err=[%s]", funName, err.Error())
		log4plus.Error(errString)
		return err, errString
	}
	if int(result.ResultCode.(float64)) < 0 {
		errString := fmt.Sprintf("%s int(result.ResultCode.(float64)) = [%d]", funName, int(result.ResultCode.(float64)))
		log4plus.Error(errString)
		return err, errString
	}
	log4plus.Info("%s postData result=[%s]", funName, resultString)
	db.SingtonAPITasksDB().InsertAiTask(result.TaskId.(string), apiKey, string(data), fmt.Sprintf("%s", requestTime.Format("2006-01-02 15:04:05")), url, method)

	responseTime := time.Now()
	db.SingtonAPITasksDB().SetAiTaskRunning(result.TaskId.(string), resultString, fmt.Sprintf("%s", responseTime.Format("2006-01-02 15:04:05")), int(responseTime.Unix()-requestTime.Unix()))
	return nil, result.TaskId.(string)
}

func (a *RemoveBG) ReplaceBG(method, apiKey string, data []byte) (error, string) {
	funName := "ReplaceBG"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s consumption time=%d(s)", funName, time.Now().Unix()-now)
	}()

	url := configure.SingtonConfigure().Interfaces.ReplaceBG.Urls.ReplaceBG.MethodUrl
	requestTime := time.Now()

	err, resultString := a.postData(url, data)
	if err != nil {
		errString := fmt.Sprintf("%s postData Failed err=[%s]", funName, err.Error())
		log4plus.Error(errString)
		return err, errString
	}
	log4plus.Info("%s postData ResultString=[%s]", funName, resultString)

	var result ResponsePostDataInterface
	if err = json.Unmarshal([]byte(resultString), &result); err != nil {
		errString := fmt.Sprintf("%s Unmarshal Failed err=[%s]", funName, err.Error())
		log4plus.Error(errString)
		return err, errString
	}
	if result.ResultCode == nil {
		errString := fmt.Sprintf("%s result.ResultCode is nil err=[%s]", funName, err.Error())
		log4plus.Error(errString)
		return err, errString
	}
	if int(result.ResultCode.(float64)) < 0 {
		errString := fmt.Sprintf("%s int(result.ResultCode.(float64)) = [%d]", funName, int(result.ResultCode.(float64)))
		log4plus.Error(errString)
		return err, errString
	}
	log4plus.Info("%s postData result=[%s]", funName, resultString)
	db.SingtonAPITasksDB().InsertAiTask(result.TaskId.(string), apiKey, string(data), fmt.Sprintf("%s", requestTime.Format("2006-01-02 15:04:05")), url, method)

	responseTime := time.Now()
	db.SingtonAPITasksDB().SetAiTaskRunning(result.TaskId.(string), resultString, fmt.Sprintf("%s", responseTime.Format("2006-01-02 15:04:05")), int(responseTime.Unix()-requestTime.Unix()))
	return nil, result.TaskId.(string)
}

func (a *RemoveBG) pollTasks() {
	funName := "pollTasks"
	for {
		time.Sleep(time.Duration(5) * time.Second)

		methodNames := []string{"removeBG", "replaceBG"}
		err, tasks := db.SingtonAPITasksDB().GetAiTasks(methodNames)
		if err != nil {
			log4plus.Info("%s GetAiTasks err=[%s]", funName, err.Error())
			continue
		}
		for _, v := range tasks {
			log4plus.Info("%s v.Method=[%s] TaskID=[%s]", funName, v.Method, v.TaskID)

			err, resultData := a.GetData(configure.SingtonConfigure().Interfaces.RemoveBG.Urls.RemoveBG.GetTask, v.TaskID)
			if err != nil {
				log4plus.Error("%s GetData failed taskId=[%s] err=[%s]", funName, v.TaskID, err.Error())
				continue
			}
			log4plus.Info("%s GetData resultData=[%s]", funName, resultData)

			var getResult ResponseGetDataRsultInterface
			if err = json.Unmarshal([]byte(resultData), &getResult); err != nil {
				log4plus.Error("%s Unmarshal Failed err=[%s]", funName, err.Error())
				continue
			}
			if getResult.Data == nil {
				log4plus.Error("%s Unmarshal failed result.ResultCode is null", funName)
				continue
			}
			jsonData, err := json.Marshal(getResult.Data)
			if err != nil {
				log4plus.Error("%s Marshal failed result.ResultCode is null err=[%s]", funName, err.Error())
				continue
			}

			type AIModuleRsultInterface struct {
				ResultCode int `json:"result_code"`
			}
			var aiModule AIModuleRsultInterface
			if err = json.Unmarshal(jsonData, &aiModule); err != nil {
				log4plus.Error("%s Unmarshal Failed err=[%s]", funName, err.Error())
				continue
			}
			if aiModule.ResultCode < 0 {
				curTime := time.Now()
				if err != nil {
					log4plus.Error("%s time.Parse err=[%s]", funName, err.Error())
					continue
				}
				db.SingtonAPITasksDB().SetAiTaskFail(v.TaskID, resultData, fmt.Sprintf("%s", curTime.Format("2006-01-02 15:04:05")))
				continue
			}

			if aiModule.ResultCode > 0 && aiModule.ResultCode != 101 && aiModule.ResultCode != 102 {
				curTime := time.Now()
				// 将字符串转换为 time.Time 类型
				requestTime, err := time.Parse("2006-01-02 15:04:05", v.RequestTime)
				if err != nil {
					log4plus.Error("%s time.Parse err=[%s]", funName, err.Error())
					continue
				}
				duration := curTime.Sub(requestTime)
				db.SingtonAPITasksDB().SetAiTaskSuccess(v.TaskID, resultData, fmt.Sprintf("%s", curTime.Format("2006-01-02 15:04:05")), int(duration.Seconds()))
			}
		}
	}
}

func SingtonRemoveBG() *RemoveBG {
	if nil == gRemoveBG {
		gRemoveBG = &RemoveBG{}
		go gRemoveBG.pollTasks()
	}
	return gRemoveBG
}
