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
	"strings"
	"time"

	"github.com/nGPU/bot/configure"
	"github.com/nGPU/bot/db"
	"github.com/nGPU/bot/header"
	log4plus "github.com/nGPU/common/log4go"
)

var Txt2ImgWorkSpaceId = string("ngpu_000000000000007")

type Txt2Img struct {
	roots        *x509.CertPool
	rootPEM      []byte
	commandLines []*header.CommandLine
}

var gTxt2Img *Txt2Img

func (a *Txt2Img) postData(url string, data []byte) (error, string, string) {
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
		return err, errString, ""
	}
	request.Header.Add("Authorization", Txt2ImgWorkSpaceId)

	response, err := client.Do(request)
	if err != nil {
		errString := fmt.Sprintf("%s client.Do Failed url=[%s] err=[%s]", funName, url, err.Error())
		log4plus.Error(errString)
		return err, errString, ""
	}
	defer response.Body.Close()

	log4plus.Info("%s url=[%s]", funName, url)
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		errString := fmt.Sprintf("%s ioutil.ReadAll Failed url=[%s] err=[%s]", funName, url, err.Error())
		log4plus.Error(errString)
		return err, errString, ""
	}
	log4plus.Info("%s Check StatusCode response.StatusCode=[%d] responseBody=[%s]", funName, response.StatusCode, string(responseBody))
	if response.StatusCode != 200 {
		errString := fmt.Sprintf("%s client.Do url=[%s] response.StatusCode=[%d] responseBody=[%s]", funName, url, response.StatusCode, string(responseBody))
		log4plus.Error(errString)
		return errors.New(errString), errString, ""
	}

	type ResponseResult struct {
		ResultCode interface{} `json:"result_code"`
	}
	var result ResponseResult
	if err = json.Unmarshal(responseBody, &result); err != nil {
		errString := fmt.Sprintf("%s Unmarshal url=[%s] responseBody=[%s]", funName, url, string(responseBody))
		log4plus.Error(errString)
		return err, errString, ""
	}
	if int(result.ResultCode.(float64)) == 200 {
		type Ali struct {
			ImageUrl string `json:"imageUrl"`
			Illegal  bool   `json:"illegal"`
		}

		type Response struct {
			ResultCode  int    `json:"result_code"`
			Msg         string `json:"msg"`
			ResultSize  int    `json:"result_size"`
			TaskDursion int    `json:"task_dursion"`
			Prompt      string `json:"prompt"`
			Alis        []Ali  `json:"alis"`
		}
		var response Response
		err := json.Unmarshal(responseBody, &response)
		if err != nil {
			errString := fmt.Sprintf("%s Unmarshal url=[%s] responseBody=[%s]", funName, url, string(responseBody))
			log4plus.Error(errString)
			return err, errString, ""
		}

		var imageUrls []string
		for _, v := range response.Alis {
			imageUrls = append(imageUrls, v.ImageUrl)
		}
		result := strings.Join(imageUrls, "\n")
		return nil, response.Prompt, result
	} else {
		errString := fmt.Sprintf("%s result result_code=[%d]", funName, int(result.ResultCode.(float64)))
		log4plus.Error(errString)
		return errors.New(errString), string(responseBody), ""
	}
}

func (a *Txt2Img) Txt2Img(method, apiKey string, data []byte) (error, string, string) {
	funName := "Txt2Img"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s consumption time=%d(s)", funName, time.Now().Unix()-now)
	}()
	url := configure.SingtonConfigure().Interfaces.Txt2Img.Urls.Txt2img.MethodUrl
	requestTime := time.Now()

	//提交生成音频文件
	err, resultString, prompt := a.postData(url, data)
	if err != nil {
		errString := fmt.Sprintf("%s postData Failed err=[%s]", funName, err.Error())
		log4plus.Error(errString)
		return err, errString, ""
	}
	log4plus.Info("%s postData ResultString=[%s]", funName, resultString)

	taskId := header.GetTaskId()
	db.SingtonAPITasksDB().InsertAiTask(taskId, apiKey, string(data), fmt.Sprintf("%s", requestTime.Format("2006-01-02 15:04:05")), url, method)

	responseTime := time.Now()
	db.SingtonAPITasksDB().SetAiTaskRunning(taskId, resultString, fmt.Sprintf("%s", responseTime.Format("2006-01-02 15:04:05")), int(responseTime.Unix()-requestTime.Unix()))
	return nil, resultString, prompt
}

func SingtonTxt2Img() *Txt2Img {
	if nil == gTxt2Img {
		gTxt2Img = &Txt2Img{}
	}
	return gTxt2Img
}
