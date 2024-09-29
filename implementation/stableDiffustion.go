package implementation

import (
	"bytes"
	"crypto/x509"
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

type StableDiffusion struct {
	roots   *x509.CertPool
	rootPEM []byte
	// store        header.DiscordPluginStore
	commandLines []*header.CommandLine
}

var gStableDiffusion *StableDiffusion

func (a *StableDiffusion) PostData(method, apiKey string, taskId string, data []byte) (error, []byte) {
	funName := "PostData"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s len(data)=%d consumption time=%d(s)", funName, len(data), time.Now().Unix()-now)
	}()

	//写入db
	url := fmt.Sprintf("%s", configure.SingtonConfigure().Interfaces.StableDiffusion.Urls.StableDiffusion)
	log4plus.Info("%s parse Url=[%s]", funName, url)

	requestTime := time.Now()
	db.SingtonAPITasksDB().InsertAiTask(taskId, apiKey, string(data), fmt.Sprintf("%s", requestTime.Format("2006-01-02 15:04:05")), url, method)

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

	jsonData := []byte(data)
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		errString := fmt.Sprintf("%s http.NewRequest Failed url=[%s] err=[%s]", funName, url, err.Error())
		log4plus.Error(errString)
		db.SingtonAPITasksDB().SetAiTaskFail(taskId, errString, fmt.Sprintf("%s", time.Now().Format("2006-01-02 15:04:05")))
		return err, []byte("")
	}
	response, err := client.Do(request)
	if err != nil {
		errString := fmt.Sprintf("%s client.Do Failed url=[%s] err=[%s]", funName, url, err.Error())
		log4plus.Error(errString)
		db.SingtonAPITasksDB().SetAiTaskFail(taskId, errString, fmt.Sprintf("%s", time.Now().Format("2006-01-02 15:04:05")))
		return err, []byte("")
	}
	defer response.Body.Close()

	log4plus.Info("%s url=[%s] client.Do Result=[%+v]", funName, url, response)
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		errString := fmt.Sprintf("%s ioutil.ReadAll Failed url=[%s] err=[%s]", funName, url, err.Error())
		log4plus.Error(errString)
		db.SingtonAPITasksDB().SetAiTaskFail(taskId, errString, fmt.Sprintf("%s", time.Now().Format("2006-01-02 15:04:05")))
		return err, []byte("")
	}

	log4plus.Info("%s Check StatusCode response.StatusCode=[%d] responseBody=[%s]", funName, response.StatusCode, string(responseBody))
	if response.StatusCode != 200 {
		errString := fmt.Sprintf("%s client.Do url=[%s] response.StatusCode=[%d] responseBody=[%s]", funName, url, response.StatusCode, string(responseBody))
		log4plus.Error(errString)
		db.SingtonAPITasksDB().SetAiTaskFail(taskId, errString, fmt.Sprintf("%s", time.Now().Format("2006-01-02 15:04:05")))
		return err, []byte("")
	}

	responseTime := time.Now()
	db.SingtonAPITasksDB().SetAiTaskSuccess(taskId, string(responseBody), fmt.Sprintf("%s", responseTime.Format("2006-01-02 15:04:05")), int(responseTime.Unix()-requestTime.Unix()))
	return nil, responseBody
}

func SingtonStableDiffusion() *StableDiffusion {
	if nil == gStableDiffusion {
		gStableDiffusion = &StableDiffusion{}
	}
	return gStableDiffusion
}
