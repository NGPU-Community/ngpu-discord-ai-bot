package telegram

import (
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/nGPU/bot/header"
	"github.com/nGPU/bot/implementation"
	log4plus "github.com/nGPU/common/log4go"
	tele "gopkg.in/telebot.v3"
)

var (
	sadTalkerPrompt = [...]string{
		"Please upload the images for the video you want to create: ",
		"Please enter the text you will use in the video: ",
		"Please select the voiceover artist you will use in the video: ",
		"Please select the background image you will use in the video: ",
	}
)

type SadTalker struct {
	roots        *x509.CertPool
	rootPEM      []byte
	store        header.TelegramPluginStore
	commandLines []*header.CommandLine
}

var gSadTalker *SadTalker

func (a *SadTalker) parseSadTalker(c tele.Context, user *header.UserStep, data string) (error, []byte, header.RequestSadTalker) {
	funName := "parseSadTalker"
	if user == nil {
		user := &header.UserStep{
			TelegramId:   c.Sender().ID,
			MaxStep:      0,
			FunctionName: "sadTalker",
			MessageId:    c.Message().ID,
		}
		a.store.AddUser(user)
		a.store.GetBotObject().Send(c.Chat(), sadTalkerPrompt[user.MaxStep], &tele.SendOptions{
			ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
		})
		log4plus.Info("%s user=[%+v]", funName, *user)
		return nil, []byte(""), header.RequestSadTalker{}
	}
	log4plus.Info("%s user=[%+v]", funName, *user)

	if user.MaxStep == 0 {
		//ImageUrl
		photo := c.Message().Photo
		if photo == nil {
			errString := fmt.Sprintf("%s c.Message().Photo is null", funName)
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestSadTalker{}
		}
		err, imageUrl := a.store.SaveMediaFile(photo.FileID)
		if err != nil {
			errString := fmt.Sprintf("%s a.store.SaveMediaFile Failed fileId=[%s]", funName, photo.FileID)
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestSadTalker{}
		}
		log4plus.Info("%s imageUrl=[%s]", funName, imageUrl)

		request := header.RequestSadTalker{
			ImageUrl:       imageUrl,
			Text:           "",
			Pronouncer:     "",
			BackGroundName: "",
			LogoUrl:        "",
		}
		user.MaxStep++
		step := header.Step{
			StepId:      user.MaxStep,
			RequestTime: time.Now(),
		}
		step.Data, err = json.Marshal(request)
		if err != nil {
			errString := fmt.Sprintf("%s DownloadFile sourceUrl=[%s]", funName, imageUrl)
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestSadTalker{}
		}
		user.Steps = append(user.Steps, step)
		a.store.GetBotObject().Send(c.Chat(), sadTalkerPrompt[user.MaxStep], &tele.SendOptions{
			ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
		})
		return nil, []byte(""), header.RequestSadTalker{}

	} else if user.MaxStep == 1 {
		//Text
		txt := c.Text()
		log4plus.Info("%s txt=[%s]", funName, txt)

		var request header.RequestSadTalker
		err := json.Unmarshal(user.Steps[user.MaxStep-1].Data, &request)
		if err != nil {
			errString := fmt.Sprintf("%s Unmarshal err=[%s]", funName, err.Error())
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestSadTalker{}
		}
		request.Text = txt
		user.MaxStep++
		step := header.Step{
			StepId:      user.MaxStep,
			RequestTime: time.Now(),
		}
		step.Data, err = json.Marshal(request)
		if err != nil {
			errString := fmt.Sprintf("%s Marshal request.Text=[%s]", funName, request.Text)
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestSadTalker{}
		}
		user.Steps = append(user.Steps, step)

		//set pronouncer
		a.store.ClearInlineFunction()
		a.store.SetInlineFunction("sadTalker", "en-US-GuyNeural", a.sadtalkerCallback)
		guyNeuralButton := tele.InlineButton{
			Text: fmt.Sprintf("👨 en-US-GuyNeural"),
			Data: "en-US-GuyNeural",
		}

		a.store.SetInlineFunction("sadTalker", "zh-CN-XiaoxiaoNeural", a.sadtalkerCallback)
		xiaoxiaoNeuralButton := tele.InlineButton{
			Text: fmt.Sprintf("🦰 zh-CN-XiaoxiaoNeural"),
			Data: "zh-CN-XiaoxiaoNeural",
		}

		a.store.SetInlineFunction("sadTalker", "zh-CN-YunxiNeural", a.sadtalkerCallback)
		yunxiNeuralButton := tele.InlineButton{
			Text: fmt.Sprintf("👨 zh-CN-YunxiNeural"),
			Data: "zh-CN-YunxiNeural",
		}

		a.store.SetInlineFunction("sadTalker", "zh-HK-HiuGaaiNeural", a.sadtalkerCallback)
		hiuGaaiNeuralButton := tele.InlineButton{
			Text: fmt.Sprintf("🦰 zh-HK-HiuGaaiNeural"),
			Data: "zh-HK-HiuGaaiNeural",
		}
		pronouncerMarkup := &tele.ReplyMarkup{
			InlineKeyboard: [][]tele.InlineButton{
				{guyNeuralButton, xiaoxiaoNeuralButton},  // 第一行
				{yunxiNeuralButton, hiuGaaiNeuralButton}, // 第二行
			},
		}
		c.Send(sadTalkerPrompt[user.MaxStep], pronouncerMarkup)
		return nil, []byte(""), header.RequestSadTalker{}

	} else if user.MaxStep == 2 {

		var request header.RequestSadTalker
		err := json.Unmarshal(user.Steps[user.MaxStep-1].Data, &request)
		if err != nil {
			errString := fmt.Sprintf("%s Unmarshal err=[%s]", funName, err.Error())
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestSadTalker{}
		}
		request.Pronouncer = data
		user.MaxStep++
		step := header.Step{
			StepId:      user.MaxStep,
			RequestTime: time.Now(),
		}
		step.Data, err = json.Marshal(request)
		if err != nil {
			errString := fmt.Sprintf("%s Marshal request.Text=[%s]", funName, request.Text)
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestSadTalker{}
		}
		user.Steps = append(user.Steps, step)

		a.store.GetBotObject().Send(c.Chat(), sadTalkerPrompt[user.MaxStep], &tele.SendOptions{
			ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
		})
		return nil, []byte(""), header.RequestSadTalker{}
	} else if user.MaxStep == 3 {
		//BackGroundName
		txt := c.Text()
		log4plus.Info("%s txt=[%s]", funName, txt)

		var request header.RequestSadTalker
		err := json.Unmarshal(user.Steps[user.MaxStep-1].Data, &request)
		if err != nil {
			errString := fmt.Sprintf("%s Unmarshal err=[%s]", funName, err.Error())
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestSadTalker{}
		}
		request.BackGroundName = txt
		user.MaxStep++
		step := header.Step{
			StepId:      user.MaxStep,
			RequestTime: time.Now(),
		}
		step.Data, err = json.Marshal(request)
		if err != nil {
			errString := fmt.Sprintf("%s Marshal request.Text=[%s]", funName, request.Text)
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestSadTalker{}
		}
		user.Steps = append(user.Steps, step)

		tmpData, err := json.Marshal(request)
		if err != nil {
			errString := fmt.Sprintf("%s json.Marshal Failed err=[%s]", funName, err.Error())
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestSadTalker{}
		}

		var body header.RequestData
		body.BTCAddr = "0000000000000000000000000GFg7xJaNVN2"
		body.Data = json.RawMessage(tmpData)

		data, err := json.Marshal(body)
		if err != nil {
			errString := fmt.Sprintf("%s json.Marshal Failed err=[%s]", funName, err.Error())
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestSadTalker{}
		}
		return nil, data, request
	}
	errString := fmt.Sprintf("%s current step Failed maxStep=[%d]", funName, user.MaxStep)
	log4plus.Info("%s errString=[%s]", funName, errString)
	return errors.New(errString), []byte(""), header.RequestSadTalker{}
}

func (a *SadTalker) sadtalkerCallback(c tele.Context, data string) error {
	funName := "sadtalker"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s consumption time=%d(s)", funName, time.Now().Unix()-now)
	}()
	log4plus.Info("%s telegramID=[%d]", funName, c.Sender().ID)

	user := a.store.FindUser(c.Sender().ID)
	if user == nil {
		log4plus.Info("%s first parseSadTalker=[%d]", funName, c.Sender().ID)
		a.parseSadTalker(c, nil, data)
		return nil
	}
	if user.MaxStep < 3 {
		a.parseSadTalker(c, user, data)
		return nil
	}

	err, body, _ := a.parseSadTalker(c, user, data)
	if err != nil {
		errString := fmt.Sprintf("%s parseSadTalker Failed telegramID=[%s]", funName, c.Sender().ID)
		log4plus.Error(errString)
		a.store.GetBotObject().Send(c.Chat(), errString, &tele.SendOptions{
			ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
		})
		return errors.New(errString)
	}
	apiKey := "123456"
	err, taskId := implementation.SingtonRemoveBG().RemoveBG(funName, apiKey, body)
	if err != nil {
		a.store.GetBotObject().Send(c.Chat(), err.Error(), &tele.SendOptions{
			ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
		})
		return err
	}
	taskMessage, err := a.store.GetBotObject().Send(c.Chat(), fmt.Sprintf("%s processing, taskId=[%s] please wait patiently... ...", funName, taskId), &tele.SendOptions{
		ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
	})

	err, resultData := a.store.WaitingTaskId(taskId)
	if err != nil {
		errString := fmt.Sprintf("%s WaitingTaskId Failed err=[%s]", funName, err.Error())
		log4plus.Error(errString)
		a.store.GetBotObject().Edit(taskMessage, errString, &tele.SendOptions{
			ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
		})
		return errors.New(errString)
	}
	log4plus.Info("%s WaitingTaskId resultData=[%s]", funName, resultData)

	type ResponseAiModuleResult struct {
		ResultCode interface{} `json:"result_code"`
		ResultUrl  interface{} `json:"result_url"`
	}
	type ResponseResult struct {
		Data       ResponseAiModuleResult `json:"data"`
		ResultCode interface{}            `json:"result_code"`
	}
	var result ResponseResult
	json.Unmarshal([]byte(resultData), &result)
	if result.ResultCode == nil {
		errString := fmt.Sprintf("%s result.ResultCode is nil", funName)
		log4plus.Error(errString)
		a.store.GetBotObject().Edit(taskMessage, errString, &tele.SendOptions{
			ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
		})
		return errors.New(errString)
	}
	if result.Data.ResultCode == nil {
		errString := fmt.Sprintf("%s data is nil", funName)
		log4plus.Error(errString)
		a.store.GetBotObject().Edit(taskMessage, errString, &tele.SendOptions{
			ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
		})
		return errors.New(errString)
	} else if int(result.Data.ResultCode.(float64)) == 100 {
		imageUrl := result.Data.ResultUrl.(string)
		err, newUrl, localPath := a.store.DownloadFile(imageUrl)
		if err != nil {
			errString := fmt.Sprintf("%s DownloadFile newUrl=[%s]", funName, newUrl)
			log4plus.Error("%s errString=[%s]", funName, errString)
			a.store.GetBotObject().Edit(taskMessage, errString, &tele.SendOptions{
				ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
			})
			return errors.New(errString)
		}
		log4plus.Info("%s DownloadFile newUrl=[%s] localPath=[%s]", funName, newUrl, localPath)
		a.store.DeleteUser(c.Sender().ID)

		var replyMessage []string
		replyMessage = append(replyMessage, fmt.Sprintf("obtained image: %s", newUrl))
		cmdlines := strings.Join(replyMessage, "\n")
		a.store.GetBotObject().Edit(taskMessage, cmdlines, &tele.SendOptions{
			ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
		})
		return nil
	}
	errString := fmt.Sprintf("%s data is Failed data=[%s]", funName, resultData)
	log4plus.Error(errString)
	a.store.GetBotObject().Send(c.Chat(), errString, &tele.SendOptions{
		ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
	})
	return errors.New(errString)
}

func (a *SadTalker) sadtalker(c tele.Context) error {
	funName := "sadtalker"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s consumption time=%d(s)", funName, time.Now().Unix()-now)
	}()
	telegramID := c.Sender()
	log4plus.Info("%s telegramID=[%d]", funName, telegramID.ID)

	user := a.store.FindUser(telegramID.ID)
	if user == nil {
		log4plus.Info("%s first parseSadTalker=[%d]", funName, telegramID.ID)
		a.parseSadTalker(c, nil, "")
		return nil
	}
	if user.MaxStep < 3 {
		a.parseSadTalker(c, user, "")
		return nil
	}

	err, body, _ := a.parseSadTalker(c, user, "")
	if err != nil {
		errString := fmt.Sprintf("%s parseSadTalker Failed telegramID=[%s]", funName, telegramID.ID)
		log4plus.Error(errString)
		a.store.GetBotObject().Send(c.Chat(), errString, &tele.SendOptions{
			ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
		})
		return errors.New(errString)
	}
	apiKey := "123456"
	err, taskId := implementation.SingtonSadTalker().SadTalker(funName, apiKey, body)
	if err != nil {
		a.store.GetBotObject().Send(c.Chat(), err.Error(), &tele.SendOptions{
			ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
		})
		return err
	}
	taskMessage, err := a.store.GetBotObject().Send(c.Chat(), fmt.Sprintf("%s processing, taskId=[%s] please wait patiently... ...", funName, taskId), &tele.SendOptions{
		ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
	})

	err, resultData := a.store.WaitingTaskId(taskId)
	if err != nil {
		errString := fmt.Sprintf("%s WaitingTaskId Failed err=[%s]", funName, err.Error())
		log4plus.Error(errString)
		a.store.GetBotObject().Edit(taskMessage, errString, &tele.SendOptions{
			ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
		})
		return errors.New(errString)
	}
	log4plus.Info("%s WaitingTaskId resultData=[%s]", funName, resultData)

	type ResponseAiModuleResult struct {
		ResultCode interface{} `json:"result_code"`
		ResultUrl  interface{} `json:"result_url"`
	}
	type ResponseResult struct {
		Data       ResponseAiModuleResult `json:"data"`
		ResultCode interface{}            `json:"result_code"`
	}
	var result ResponseResult
	json.Unmarshal([]byte(resultData), &result)
	if result.ResultCode == nil {
		errString := fmt.Sprintf("%s result.ResultCode is nil", funName)
		log4plus.Error(errString)
		a.store.GetBotObject().Edit(taskMessage, errString, &tele.SendOptions{
			ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
		})
		return errors.New(errString)
	}
	if result.Data.ResultCode == nil {
		errString := fmt.Sprintf("%s data is nil", funName)
		log4plus.Error(errString)
		a.store.GetBotObject().Edit(taskMessage, errString, &tele.SendOptions{
			ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
		})
		return errors.New(errString)
	} else if int(result.Data.ResultCode.(float64)) == 100 {
		imageUrl := result.Data.ResultUrl.(string)
		err, newUrl, localPath := a.store.DownloadFile(imageUrl)
		if err != nil {
			errString := fmt.Sprintf("%s DownloadFile newUrl=[%s]", funName, newUrl)
			log4plus.Error("%s errString=[%s]", funName, errString)
			a.store.GetBotObject().Edit(taskMessage, errString, &tele.SendOptions{
				ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
			})
			return errors.New(errString)
		}
		log4plus.Info("%s DownloadFile newUrl=[%s] localPath=[%s]", funName, newUrl, localPath)
		a.store.DeleteUser(telegramID.ID)

		var replyMessage []string
		replyMessage = append(replyMessage, fmt.Sprintf("obtained image: %s", newUrl))
		cmdlines := strings.Join(replyMessage, "\n")
		a.store.GetBotObject().Edit(taskMessage, cmdlines, &tele.SendOptions{
			ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
		})
		return nil
	}
	errString := fmt.Sprintf("%s data is Failed data=[%s]", funName, resultData)
	log4plus.Error(errString)
	a.store.GetBotObject().Send(c.Chat(), errString, &tele.SendOptions{
		ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
	})
	return errors.New(errString)
}

func (a *SadTalker) setFuncs() {
	a.store.SetTelegramFunction("sadtalker", "sadtalker", "🎥 sadTalker", a.sadtalker)
}

func SingtonSadTalker(store header.TelegramPluginStore) *SadTalker {
	if nil == gSadTalker {
		gSadTalker = &SadTalker{
			store: store,
		}
		gSadTalker.setFuncs()
	}
	return gSadTalker
}
