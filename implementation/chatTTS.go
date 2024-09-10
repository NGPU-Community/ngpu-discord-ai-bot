package implementation

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/nGPU/discordBot/db"

	log4plus "github.com/nGPU/common/log4go"
	"github.com/nGPU/discordBot/common"
	"github.com/nGPU/discordBot/configure"
	"github.com/nGPU/discordBot/header"
)

type ChatTTS struct {
	roots        *x509.CertPool
	rootPEM      []byte
	store        header.PluginStore
	commandLines []*header.CommandLine
}

var gChatTTS *ChatTTS

func (a *ChatTTS) GenerateAudio(method, apiKey string, chat header.ChatTTSData) (error, header.ResponseTTS) {
	funName := "GenerateAudio"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s consumption time=%d(s)", funName, time.Now().Unix()-now)
	}()

	//提交生成音频文件
	err, audioFile := a.generateAudio(method, apiKey, chat)
	if err != nil {
		errString := fmt.Sprintf("%s generateAudio Failed err=[%s]", funName, err.Error())
		log4plus.Error(errString)
		return err, header.ResponseTTS{}
	}
	//获取音频文件
	err, audioFileUrl := a.getAudio(audioFile)
	if err != nil {
		errString := fmt.Sprintf("%s getAudioPath Failed err=[%s]", funName, err.Error())
		log4plus.Error(errString)
		return err, header.ResponseTTS{}
	}

	var reponse header.ResponseTTS
	reponse.AudioUrl = audioFileUrl
	return nil, reponse
}

func (a *ChatTTS) generateAudio(method, apiKey string, chat header.ChatTTSData) (error, string) {
	funName := "generateAudio"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s chat=[%+v] consumption time=%d(s)", funName, chat, time.Now().Unix()-now)
	}()

	//写入db
	// url := fmt.Sprintf("%s", configure.SingtonConfigure().Interfaces.ChatTTS.Urls.GenerateAudio)
	url := ""
	log4plus.Info("%s parse Url=[%s]", funName, url)

	requestTime := time.Now()
	db.SingtonAPITasksDB().InsertAiTask(chat.TaskId, apiKey, fmt.Sprintf("%+v", chat), fmt.Sprintf("%s", requestTime.Format("2006-01-02 15:04:05")), url, method)

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

	jsonData, err := json.Marshal(chat)
	if err != nil {
		errString := fmt.Sprintf("%s json.Marshal Failed err=[%s]", funName, err.Error())
		log4plus.Error(errString)
		db.SingtonAPITasksDB().SetAiTaskFail(chat.TaskId, errString, fmt.Sprintf("%s", time.Now().Format("2006-01-02 15:04:05")))
		return err, ""
	}
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		errString := fmt.Sprintf("%s http.NewRequest Failed url=[%s] err=[%s]", funName, url, err.Error())
		log4plus.Error(errString)
		db.SingtonAPITasksDB().SetAiTaskFail(chat.TaskId, errString, fmt.Sprintf("%s", time.Now().Format("2006-01-02 15:04:05")))
		return err, ""
	}
	response, err := client.Do(request)
	if err != nil {
		errString := fmt.Sprintf("%s client.Do Failed url=[%s] err=[%s]", funName, url, err.Error())
		log4plus.Error(errString)
		db.SingtonAPITasksDB().SetAiTaskFail(chat.TaskId, errString, fmt.Sprintf("%s", time.Now().Format("2006-01-02 15:04:05")))
		return err, ""
	}
	defer response.Body.Close()

	log4plus.Info("%s url=[%s] client.Do Result=[%+v]", funName, url, response)
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		errString := fmt.Sprintf("%s ioutil.ReadAll Failed url=[%s] err=[%s]", funName, url, err.Error())
		log4plus.Error(errString)
		db.SingtonAPITasksDB().SetAiTaskFail(chat.TaskId, errString, fmt.Sprintf("%s", time.Now().Format("2006-01-02 15:04:05")))
		return err, ""
	}

	log4plus.Info("%s Check StatusCode response.StatusCode=[%d] responseBody=[%s]", funName, response.StatusCode, string(responseBody))
	if response.StatusCode != 200 {
		errString := fmt.Sprintf("%s client.Do url=[%s] response.StatusCode=[%d] responseBody=[%s]", funName, url, response.StatusCode, string(responseBody))
		log4plus.Error(errString)
		db.SingtonAPITasksDB().SetAiTaskFail(chat.TaskId, errString, fmt.Sprintf("%s", time.Now().Format("2006-01-02 15:04:05")))
		return err, ""
	}

	var result header.ResultRecord
	if err := json.Unmarshal(responseBody, &result); err != nil {
		errString := fmt.Sprintf("%s while parsing response err=[%s]", funName, err.Error())
		log4plus.Error(errString)
		db.SingtonAPITasksDB().SetAiTaskFail(chat.TaskId, errString, fmt.Sprintf("%s", time.Now().Format("2006-01-02 15:04:05")))
		return errors.New(errString), ""
	}
	responseTime := time.Now()
	db.SingtonAPITasksDB().SetAiTaskSuccess(chat.TaskId, string(responseBody), fmt.Sprintf("%s", responseTime.Format("2006-01-02 15:04:05")), int(responseTime.Unix()-requestTime.Unix()))
	return nil, result.AudioUrl
}

func (a *ChatTTS) getAudio(audioFile string) (error, string) {
	funName := "getAudio"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s audioFile=[%s] consumption time=%d(s)", funName, audioFile, time.Now().Unix()-now)
	}()

	url := fmt.Sprintf("%s", audioFile)
	log4plus.Info("%s parse Url=[%s]", funName, url)
	client := &http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				c, err := net.DialTimeout(netw, addr, time.Minute*10)
				if err != nil {
					log4plus.Error("%s dail timeout err=[%s]", funName, err.Error())
					return nil, err
				}
				return c, nil
			},
			MaxIdleConnsPerHost:   10,
			ResponseHeaderTimeout: time.Minute * 10,
		},
	}
	defer client.CloseIdleConnections()

	log4plus.Info("%s NewRequest Url=[%s]", funName, url)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log4plus.Error("%s http.NewRequest Failed url=[%s] err=[%s]", funName, url, err.Error())
		return err, ""
	}
	response, err := client.Do(request)
	if err != nil {
		log4plus.Error("%s client.Do Failed url=[%s] err=[%s]", funName, url, err.Error())
		return err, ""
	}
	defer response.Body.Close()

	log4plus.Info("%s url=[%s] client.Do response.StatusCode=[%d]", funName, url, response.StatusCode)
	if response.StatusCode != 200 {
		errString := fmt.Sprintf("%s response.StatusCode not is 200", funName)
		log4plus.Error(errString)
		return errors.New(errString), ""
	}

	//设置保存文件名
	fileExt := common.GetFileExtension(url)
	fileName := fmt.Sprintf("%s%s.%s", time.Now().Format("20060102150405"), fmt.Sprintf("%06d", time.Now().Nanosecond()/1e3), fileExt)
	filePath := fmt.Sprintf("%s/%s", configure.SingtonConfigure().Resource.ResourcePath, fileName)

	// 创建文件
	outFile, err := os.Create(filePath)
	if err != nil {
		errString := fmt.Sprintf("%s os.Create Failed err=[%s]", funName, err.Error())
		log4plus.Error(errString)
		return errors.New(errString), ""
	}
	defer outFile.Close()

	// 将响应的Body复制到文件中
	_, err = io.Copy(outFile, response.Body)
	if err != nil {
		errString := fmt.Sprintf("%s os.Copy Failed err=[%s]", funName, err.Error())
		log4plus.Error(errString)
		return errors.New(errString), ""
	}

	//设置返回Url地址
	audioUrl := fmt.Sprintf("%s/%s", configure.SingtonConfigure().Resource.Domain, fileName)
	log4plus.Info("%s audioUrl=[%s]", funName, audioUrl)
	return nil, audioUrl
}

func SingtonChatTTS() *ChatTTS {
	if nil == gChatTTS {
		gChatTTS = &ChatTTS{}
	}
	return gChatTTS
}
