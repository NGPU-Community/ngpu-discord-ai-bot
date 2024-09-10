package plugins

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	// "github.com/nGPU/discordBot/plugins/telegram"

	log4plus "github.com/nGPU/common/log4go"
	"github.com/nGPU/discordBot/configure"
	"github.com/nGPU/discordBot/db"
	"github.com/nGPU/discordBot/header"
	tele "gopkg.in/telebot.v3"
)

type Telegram struct {
	roots    *x509.CertPool
	rootPEM  []byte
	msgFuncs map[string]tele.HandlerFunc
	bot      *tele.Bot
}

var gTelegram *Telegram

func (w *Telegram) Version() string {
	return "Telegram Version 1.0.0"
}

func (w *Telegram) SetTelegramFunction(funName string, fun tele.HandlerFunc) {
	w.msgFuncs[funName] = fun
}

func (w *Telegram) CheckBTCAddress(btcAddress string) bool {
	funName := "CheckBTCAddress"
	err, exist := w.getBRC20(btcAddress, w.roots)
	if err != nil {
		log4plus.Error(fmt.Sprintf("%s getBRC20 Failed err=[%s]", funName, err.Error()))
		return false
	}
	return exist
}

func (w *Telegram) GetUserBase(telegramId string) (error, bool, header.ResponseUserInfo) {
	funName := "GetUserBase"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s telegramId=[%s] consumption time=%d(s)", funName, telegramId, time.Now().Unix()-now)
	}()
	err, userInfo := db.SingtonUserDB().GetUser(telegramId)
	if err != nil {
		return err, false, header.ResponseUserInfo{}
	}
	return nil, true, userInfo
}

func (w *Telegram) getBRC20(BtcAddress string, roots *x509.CertPool) (error, bool) {
	funName := "getBRC20"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s BtcAddress=[%s] consumption time=%d(s)", funName, BtcAddress, time.Now().Unix()-now)
	}()
	tr := &http.Transport{
		TLSClientConfig:     &tls.Config{RootCAs: roots},
		TLSHandshakeTimeout: 30 * time.Second,
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Duration(30) * time.Second,
	}
	defer client.CloseIdleConnections()
	//创建请求
	var reqUri = fmt.Sprintf("%s?address=%s", configure.SingtonConfigure().Interfaces.BTC.BtcAmtUri, BtcAddress)
	req, err := http.NewRequest("GET", reqUri, nil)
	log4plus.Info("%s http.NewRequest url=[%s]", funName, reqUri)
	//发起请求
	response, err := client.Do(req)
	if err != nil {
		log4plus.Error("%s client.Do Failed err=[%s] url=[%s]", funName, err.Error(), reqUri)
		return err, false
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusOK {
		log4plus.Info("%s StatusCode is http.StatusOK response=[%+v]", funName, *response)
		body, errRes := ioutil.ReadAll(response.Body)
		if errRes != nil {
			log4plus.Error("%s ReadAll Failed err=[%s] url=[%s]", funName, errRes.Error(), reqUri)
			return errRes, false
		}
		brcResponse := new(struct {
			Exists   bool     `json:"exists"`   //是否存在指定的铭文
			Names    []string `json:"names"`    //名称
			Balances []int64  `json:"balances"` //数量
		})
		if err := json.Unmarshal(body, &brcResponse); err != nil {
			log4plus.Error("%s ReadAll Failed err=[%s] url=[%s]", funName, err.Error(), reqUri)
			return err, false
		}
		log4plus.Info("%s Result=[%+v]", funName, *brcResponse)
		return nil, brcResponse.Exists

	} else if response.StatusCode == 400 {
		log4plus.Info("%s StatusCode=[%d] BtcAddress=[%s]", funName, response.StatusCode, BtcAddress)
		return nil, false
	}
	log4plus.Error("%s response=[%+v]", funName, *response)
	errString := fmt.Sprintf("%s StatusCode=[%d]", funName, response.StatusCode)
	return errors.New(errString), false
}

func (w *Telegram) WaitingTaskId(taskId string) (error, string) {
	funName := "WaitingTaskId"
	// waiting
	maxTimeout := int64(5 * 60 * 60)
	for {
		time.Sleep(time.Duration(5) * time.Second)
		maxTimeout = maxTimeout - 10
		if maxTimeout <= 0 {
			return errors.New("task timeout"), ""
		}
		err, taskInformation := db.SingtonAPITasksDB().GetTaskId(taskId)
		if err != nil {
			return err, ""
		}
		if taskInformation.State == header.FinishState {
			return nil, taskInformation.Response
		} else if taskInformation.State == header.ErrorState {
			type ResponseResult struct {
				ResultCode interface{} `json:"result_code"`
			}
			var result ResponseResult
			json.Unmarshal([]byte(taskInformation.Response), &result)
			if result.ResultCode == nil {
				errString := fmt.Sprintf("%s result.ResultCode is nil", funName)
				log4plus.Error(errString)
				return errors.New(errString), ""
			}
			errString := fmt.Sprintf("%s Ai Module Result Code is [%d]", funName, int(result.ResultCode.(float64)))
			log4plus.Error(errString)
			return errors.New(errString), ""
		}
	}
}

func (w *Telegram) getUrlExtension(urlString string) string {
	funName := "getUrlExtension"
	parsedURL, err := url.Parse(urlString)
	if err != nil {
		log4plus.Error("%s Failed err=%s", funName, err.Error())
		return ""
	}
	fileName := path.Base(parsedURL.Path)
	fileExtension := strings.TrimPrefix(path.Ext(fileName), ".")
	log4plus.Info("%s fileExtension=%s", funName, fileExtension)
	return fileExtension
}

func (w *Telegram) DownloadFile(url string) (err error, newUrl string, localPath string) {
	funName := "DownloadFile"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s url=%s consumption time=%d(s)", funName, url, time.Now().Unix()-now)
	}()

	response, err := http.Get(url)
	if err != nil {
		log4plus.Error("%s Get Failed err=%s", funName, err.Error())
		return err, "", ""
	}
	defer response.Body.Close()

	fileExt := w.getUrlExtension(url)
	fileName := fmt.Sprintf("%s%s.%s", time.Now().Format("20060102150405"), fmt.Sprintf("%06d", time.Now().Nanosecond()/1e3), fileExt)
	localPath = fmt.Sprintf("%s/%s", configure.SingtonConfigure().Resource.ResourcePath, fileName)
	file, err := os.Create(localPath)
	if err != nil {
		log4plus.Error("%s Create Failed err=%s", funName, err.Error())
		return err, "", ""
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		log4plus.Error("%s Copy Failed err=%s", funName, err.Error())
		return err, "", ""
	}
	newUrl = fmt.Sprintf("%s/%s", configure.SingtonConfigure().Resource.Domain, fileName)
	return nil, newUrl, localPath
}

func (w *Telegram) Init() {
	funName := "Init"
	pref := tele.Settings{
		Token:  configure.SingtonConfigure().Token.Telegram.Token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}
	var err error
	if w.bot, err = tele.NewBot(pref); err != nil {
		log4plus.Error("%s tele.NewBot=[%s]", funName, err.Error())
		return
	}
	for key, v := range w.msgFuncs {
		w.bot.Handle(key, v)
	}
	w.bot.Start()
}

func SingtonTelegram() *Telegram {
	if nil == gTelegram {
		gTelegram = &Telegram{}
		gTelegram.Init()

		// log4plus.Info("telegram.SingtonBlip(gTelegram)")
		// telegram.SingtonBlip(gTelegram)

		gTelegram.Init()

	}
	return gTelegram
}
