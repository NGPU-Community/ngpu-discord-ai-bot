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
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/nGPU/bot/configure"
	"github.com/nGPU/bot/db"
	"github.com/nGPU/bot/header"
	"github.com/nGPU/bot/plugins/telegram"
	log4plus "github.com/nGPU/common/log4go"
	tele "gopkg.in/telebot.v3"
)

type ButtonInfo struct {
	Major   string
	Minor   string
	ShowTxt string
	Fun     header.TelegramMsgFunction
}

type InlineButtonInfo struct {
	Major string
	Data  string
	Fun   header.TelegramInlineFunction
}

type Telegram struct {
	roots       *x509.CertPool
	rootPEM     []byte
	msgFuncs    map[string]*ButtonInfo
	inlineFuncs []*InlineButtonInfo
	bot         *tele.Bot
	userLock    sync.Mutex
	users       map[int64]*header.UserStep
}

var gTelegram *Telegram

func (w *Telegram) AddUser(user *header.UserStep) error {
	if user == nil {
		log4plus.Error("user is nil")
		return errors.New("user is nil")
	}
	w.userLock.Lock()
	defer w.userLock.Unlock()
	w.users[user.TelegramId] = user
	return nil
}

func (w *Telegram) DeleteUser(telegramId int64) {
	w.userLock.Lock()
	defer w.userLock.Unlock()
	delete(w.users, telegramId)
}

func (w *Telegram) FindUser(telegramId int64) *header.UserStep {
	w.userLock.Lock()
	defer w.userLock.Unlock()
	user, Ok := w.users[telegramId]
	if Ok {
		return user
	}
	return nil
}

func (w *Telegram) Version() string {
	return "Telegram Version 1.0.0"
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

	var reqUri = fmt.Sprintf("%s?address=%s", configure.SingtonConfigure().Interfaces.BTC.BtcAmtUri, BtcAddress)
	req, err := http.NewRequest("GET", reqUri, nil)
	log4plus.Info("%s http.NewRequest url=[%s]", funName, reqUri)

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

func (w *Telegram) SetTelegramFunction(major, minor, showTxt string, fun header.TelegramMsgFunction) {
	button := &ButtonInfo{
		Major:   major,
		Minor:   minor,
		ShowTxt: showTxt,
		Fun:     fun,
	}
	w.msgFuncs[fmt.Sprintf("%s_%s", major, minor)] = button
}

func (w *Telegram) SetInlineFunction(major, data string, fun header.TelegramInlineFunction) {
	inline := &InlineButtonInfo{
		Major: major,
		Data:  data,
		Fun:   fun,
	}
	w.inlineFuncs = append(w.inlineFuncs, inline)
}

func (w *Telegram) ClearInlineFunction() {
	w.inlineFuncs = w.inlineFuncs[:0]
}

func (w *Telegram) showFunButtons(c tele.Context, buttons [][]tele.InlineButton) error {
	replyMarkup := tele.ReplyMarkup{
		InlineKeyboard: buttons,
	}
	return c.Send("Please select a specific function: ", &replyMarkup)
}

func (w *Telegram) addPlayGameButton(buttons [][]tele.InlineButton) [][]tele.InlineButton {
	gameButton := tele.InlineButton{
		Text: "🎮 Play Game",

		WebApp: &tele.WebApp{
			URL: "https://www.aiinfura.com/game/",
		},
	}
	buttons = append(buttons, []tele.InlineButton{gameButton})
	return buttons
}

func (w *Telegram) Init() {
	funName := "Init"
	pref := tele.Settings{
		Token:  configure.SingtonConfigure().Token.Telegram.Token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}
	log4plus.Info("%s Token=[%s]", funName, pref.Token)

	var err error
	if w.bot, err = tele.NewBot(pref); err != nil {
		log4plus.Error("%s tele.NewBot=[%s]", funName, err.Error())
		return
	}

	//sort Button
	var functionArray []*ButtonInfo
	for _, value := range w.msgFuncs {
		functionArray = append(functionArray, &ButtonInfo{
			Major:   value.Major,
			Minor:   value.Minor,
			ShowTxt: value.ShowTxt,
			Fun:     value.Fun,
		})
	}
	sort.Slice(functionArray, func(p, q int) bool {
		return functionArray[p].Major <= functionArray[q].Major
	})

	//Bind button format
	log4plus.Info("%s current msgFuns length=[%d]", funName, len(w.msgFuncs))
	var buttons [][]tele.InlineButton
	row := 0
	col := 0
	for _, value := range functionArray {
		log4plus.Info("%s append Major=[%s] row=[%d] col=[%d]", funName, value.Major, row, col)
		inlineBtn := tele.InlineButton{
			Text: value.ShowTxt,
			Data: value.Major,
		}
		if col == 0 {
			buttons = append(buttons, []tele.InlineButton{inlineBtn})
		} else {
			buttons[len(buttons)-1] = append(buttons[len(buttons)-1], inlineBtn)
		}
		col++
		if col == 2 {
			col = 0
			row++
		}
	}

	//Add Game Button
	buttons = w.addPlayGameButton(buttons)

	//Set up command processing
	w.bot.Handle("/start", func(c tele.Context) error {
		log4plus.Info("%s /start Handle telegramID=[%d]", funName, c.Sender().ID)

		user := w.FindUser(c.Sender().ID)
		if user != nil {
			w.DeleteUser(c.Sender().ID)
		}
		return w.showFunButtons(c, buttons)
	})

	//User uploads a file
	w.bot.Handle(tele.OnPhoto, func(c tele.Context) error {
		log4plus.Info("%s tele.OnPhoto Handle telegramID=[%d]", funName, c.Sender().ID)

		user := w.FindUser(c.Sender().ID)
		if user != nil {
			log4plus.Info("%s user Major=[%s]", funName, user.FunctionName)
			//check major
			for _, value := range w.msgFuncs {
				log4plus.Info("%s telegramID=[%d] value.Major=[%s] user.FunctionName=[%s]", funName, c.Sender().ID, value.Major, user.FunctionName)
				if strings.ToLower(value.Major) == strings.ToLower(user.FunctionName) {
					fun := header.TelegramMsgFunction(value.Fun)
					return fun(c)
				}
			}
		}

		//Return different responses based on the information from different groups
		chatType := string(c.Chat().Type)
		if strings.ToLower(chatType) == strings.ToLower("private") {
			return w.showFunButtons(c, buttons)
		} else if strings.ToLower(chatType) == strings.ToLower("group") || strings.ToLower(chatType) == strings.ToLower("supergroup") {
			return nil
		}
		return nil
	})

	//Processing user input text
	w.bot.Handle(tele.OnText, func(c tele.Context) error {
		log4plus.Info("%s tele.OnText Handle telegramID=[%d]", funName, c.Sender().ID)
		user := w.FindUser(c.Sender().ID)
		if user == nil {
			//Return different responses based on the information from different groups
			chatType := string(c.Chat().Type)
			if strings.ToLower(chatType) == strings.ToLower("private") {
				return w.showFunButtons(c, buttons)
			} else if strings.ToLower(chatType) == strings.ToLower("group") || strings.ToLower(chatType) == strings.ToLower("supergroup") {
				return nil
			} else if strings.ToLower(chatType) == strings.ToLower("channel") {
				return nil
			}
		}
		for _, value := range w.msgFuncs {
			log4plus.Info("%s telegramID=[%d] user.FunctionName=[%s]", funName, c.Sender().ID, user.FunctionName)
			if strings.ToLower(value.Major) == strings.ToLower(user.FunctionName) {
				fun := header.TelegramMsgFunction(value.Fun)
				return fun(c)
			}
		}
		return w.showFunButtons(c, buttons)
	})

	//Callback after the user clicks the button
	w.bot.Handle(tele.OnCallback, func(c tele.Context) error {
		log4plus.Info("%s tele.OnCallback Handle telegramID=[%d]", funName, c.Sender().ID)

		callback := c.Callback().Data
		log4plus.Info("%s select function click callback=[%s]", funName, callback)
		for _, value := range w.msgFuncs {
			if strings.ToLower(value.Major) == strings.ToLower(callback) {
				log4plus.Info("%s telegramID=[%d]", funName, c.Sender().ID)
				user := w.FindUser(c.Sender().ID)
				if user != nil {
					w.DeleteUser(c.Sender().ID)
				}
				fun := header.TelegramMsgFunction(value.Fun)
				return fun(c)
			}
		}

		log4plus.Info("%s select pronouncer click callback=[%s] telegramID=[%d]", funName, callback, c.Sender().ID)
		user := w.FindUser(c.Sender().ID)
		if user == nil {
			return w.showFunButtons(c, buttons)
		}
		for _, value := range w.inlineFuncs {
			if (strings.ToLower(value.Major) == strings.ToLower(user.FunctionName)) && (strings.ToLower(value.Data) == strings.ToLower(callback)) {
				fun := header.TelegramInlineFunction(value.Fun)
				return fun(c, callback)
			}
		}
		return w.showFunButtons(c, buttons)
	})

	//Listen for the event of new members joining the group.
	welcomeMessage := "Welcome! NGPU is a decentralized elastic computing network that provides you with commonly used AI services. You can type /start to begin your journey."
	w.bot.Handle(tele.OnUserJoined, func(c tele.Context) error {
		w.bot.Send(c.Sender(), welcomeMessage)
		return nil
	})

	log4plus.Info("%s start telegram--->>>", funName)
	w.bot.Start()
}

func (w *Telegram) GetBotObject() *tele.Bot {
	return w.bot
}

func (w *Telegram) SaveMediaFile(fileId string) (error, string) {
	funName := "SaveMediaFile"

	fileReader, err := w.bot.File(&tele.File{FileID: fileId})
	if err != nil {
		errString := fmt.Sprintf("%s a.store.GetBotObject().File Failed fileId=[%s]", funName, fileId)
		log4plus.Info("%s errString=[%s]", funName, errString)
		return errors.New(errString), ""
	}
	defer fileReader.Close()

	fileName := fmt.Sprintf("%s%s.%s", time.Now().Format("20060102150405"), fmt.Sprintf("%06d", time.Now().Nanosecond()/1e3), "png")
	localPath := fmt.Sprintf("%s/%s", configure.SingtonConfigure().Resource.ResourcePath, fileName)
	log4plus.Info("%s localPath=[%s]", funName, localPath)

	outFile, err := os.Create(localPath)
	if err != nil {
		errString := fmt.Sprintf("%s Create Failed err=%s", funName, err.Error())
		log4plus.Error("%s errString=[%s]", funName, errString)
		return errors.New(errString), ""
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, fileReader)
	if err != nil {
		errString := fmt.Sprintf("%s Copy Failed err=%s", funName, err.Error())
		log4plus.Error("%s errString=[%s]", funName, errString)
		return errors.New(errString), ""
	}
	imageUrl := fmt.Sprintf("%s/%s", configure.SingtonConfigure().Resource.Domain, fileName)
	log4plus.Info("%s imageUrl=[%s]", funName, imageUrl)
	return nil, imageUrl
}

func SingtonTelegram() *Telegram {
	if nil == gTelegram {
		gTelegram = &Telegram{
			bot:      nil,
			msgFuncs: make(map[string]*ButtonInfo),
			users:    make(map[int64]*header.UserStep),
		}

		log4plus.Info("telegram.SingtonBlip(gTelegram)")
		telegram.SingtonBlip(gTelegram)

		log4plus.Info("telegram.SingtonRemoveBG(gTelegram)")
		telegram.SingtonRemoveBG(gTelegram)

		log4plus.Info("telegram.SingtonFaceFusion(gTelegram)")
		telegram.SingtonFaceFusion(gTelegram)

		log4plus.Info("telegram.SingtonTxt2Img(gTelegram)")
		telegram.SingtonTxt2Img(gTelegram)

		log4plus.Info("telegram.SingtonSadTalker(gTelegram)")
		telegram.SingtonSadTalker(gTelegram)

		log4plus.Info("telegram.SingtonOllma(gTelegram)")
		telegram.SingtonOllma(gTelegram)

		go gTelegram.Init()
	}
	return gTelegram
}
