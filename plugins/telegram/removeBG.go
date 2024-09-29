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
	removePrompt = [...]string{
		"Please upload the image from which you want to remove the background: ",
	}

	replacePrompt = [...]string{
		"Please upload the image for which you want to replace the background: ",
		"Please upload the replacement background image: ",
	}
)

type RemoveBG struct {
	roots        *x509.CertPool
	rootPEM      []byte
	store        header.TelegramPluginStore
	commandLines []*header.CommandLine
}

var gRemoveBG *RemoveBG

func (a *RemoveBG) parseRemoveBG(c tele.Context, user *header.UserStep) (error, []byte, header.RequestRemoveBG) {
	funName := "parseRemoveBG"
	if user == nil {
		user := &header.UserStep{
			TelegramId:   c.Sender().ID,
			MaxStep:      0,
			FunctionName: "removeBG",
			MessageId:    c.Message().ID,
		}
		a.store.AddUser(user)
		a.store.GetBotObject().Send(c.Chat(), removePrompt[user.MaxStep], &tele.SendOptions{
			ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
		})
		return nil, []byte(""), header.RequestRemoveBG{}
	}
	if user.MaxStep == 0 {
		photo := c.Message().Photo
		if photo == nil {
			errString := fmt.Sprintf("%s c.Message().Photo is null", funName)
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestRemoveBG{}
		}
		err, imageUrl := a.store.SaveMediaFile(photo.FileID)
		if err != nil {
			errString := fmt.Sprintf("%s a.store.SaveMediaFile Failed fileId=[%s]", funName, photo.FileID)
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestRemoveBG{}
		}
		log4plus.Info("%s imageUrl=[%s]", funName, imageUrl)

		request := header.RequestRemoveBG{
			Kind:     "photo",
			Obj:      "any",
			ImageUrl: imageUrl,
			BGColor:  "0,0,0,0",
		}
		user.MaxStep = 1
		step := header.Step{
			StepId:      1,
			RequestTime: time.Now(),
		}
		step.Data, err = json.Marshal(request)
		if err != nil {
			errString := fmt.Sprintf("%s DownloadFile sourceUrl=[%s]", funName, imageUrl)
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestRemoveBG{}
		}
		user.Steps = append(user.Steps, step)

		tmpData, err := json.Marshal(request)
		if err != nil {
			errString := fmt.Sprintf("%s json.Marshal Failed err=[%s]", funName, err.Error())
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestRemoveBG{}
		}

		var body header.RequestData
		body.BTCAddr = "0000000000000000000000000GFg7xJaNVN2"
		body.Data = json.RawMessage(tmpData)

		data, err := json.Marshal(body)
		if err != nil {
			errString := fmt.Sprintf("%s json.Marshal Failed err=[%s]", funName, err.Error())
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestRemoveBG{}
		}
		return nil, data, request
	}
	errString := fmt.Sprintf("%s current step Failed maxStep=[%d]", funName, user.MaxStep)
	log4plus.Info("%s errString=[%s]", funName, errString)
	return errors.New(errString), []byte(""), header.RequestRemoveBG{}
}

func (a *RemoveBG) removeBG(c tele.Context) error {
	funName := "removeBG"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s consumption time=%d(s)", funName, time.Now().Unix()-now)
	}()
	telegramID := c.Sender()
	log4plus.Info("%s telegramID=[%d]", funName, telegramID.ID)

	user := a.store.FindUser(telegramID.ID)
	if user == nil {
		log4plus.Info("%s first parseRemoveBG=[%d]", funName, telegramID.ID)
		a.parseRemoveBG(c, nil)
		return nil
	}
	log4plus.Info("%s user=[%+v]", funName, *user)

	if user.MaxStep == 0 {
		err, body, _ := a.parseRemoveBG(c, user)
		if err != nil {
			errString := fmt.Sprintf("%s parseRemoveBG Failed telegramID=[%s]", funName, telegramID)
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

			var replyMessage []string
			replyMessage = append(replyMessage, fmt.Sprintf("obtained image: %s", newUrl))
			cmdlines := strings.Join(replyMessage, "\n")
			a.store.GetBotObject().Edit(taskMessage, cmdlines, &tele.SendOptions{
				ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
			})
			a.store.DeleteUser(telegramID.ID)
			return nil
		}
		errString := fmt.Sprintf("%s data is Failed data=[%s]", funName, resultData)
		log4plus.Error(errString)
		a.store.GetBotObject().Send(c.Chat(), errString, &tele.SendOptions{
			ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
		})
		return errors.New(errString)
	}
	errString := fmt.Sprintf("%s current step mexStep=[%d]", funName, user.MaxStep)
	log4plus.Error(errString)
	a.store.GetBotObject().Send(c.Chat(), errString, &tele.SendOptions{
		ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
	})
	return errors.New(errString)
}

func (a *RemoveBG) parseReplaceBG(c tele.Context, user *header.UserStep) (error, []byte, header.RequestReplaceBG) {
	funName := "parseReplaceBG"
	if user == nil {
		user := &header.UserStep{
			TelegramId:   c.Sender().ID,
			MaxStep:      0,
			FunctionName: "replaceBG",
			MessageId:    c.Message().ID,
		}
		a.store.AddUser(user)
		a.store.GetBotObject().Send(c.Chat(), replacePrompt[user.MaxStep], &tele.SendOptions{
			ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
		})
		return nil, []byte(""), header.RequestReplaceBG{}
	}
	if user.MaxStep == 0 {
		photo := c.Message().Photo
		if photo == nil {
			errString := fmt.Sprintf("%s c.Message().Photo is null", funName)
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestReplaceBG{}
		}
		err, imageUrl := a.store.SaveMediaFile(photo.FileID)
		if err != nil {
			errString := fmt.Sprintf("%s a.store.SaveMediaFile Failed fileId=[%s]", funName, photo.FileID)
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestReplaceBG{}
		}
		log4plus.Info("%s imageUrl=[%s]", funName, imageUrl)

		request := header.RequestReplaceBG{
			Kind:    "photo",
			Obj:     "any",
			SrcUrl:  imageUrl,
			BGColor: "0,0,0,0",
			BGUrl:   "",
		}
		user.MaxStep = 1
		step := header.Step{
			StepId:      1,
			RequestTime: time.Now(),
		}
		step.Data, err = json.Marshal(request)
		if err != nil {
			errString := fmt.Sprintf("%s DownloadFile sourceUrl=[%s]", funName, imageUrl)
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestReplaceBG{}
		}
		user.Steps = append(user.Steps, step)
		c.Send(replacePrompt[user.MaxStep])

		log4plus.Info("%s user=[%+v]", funName, *user)
		return nil, []byte(""), header.RequestReplaceBG{}

	} else if user.MaxStep == 1 {
		var request header.RequestReplaceBG
		err := json.Unmarshal(user.Steps[user.MaxStep-1].Data, &request)
		if err != nil {
			errString := fmt.Sprintf("%s Unmarshal err=[%s]", funName, err.Error())
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestReplaceBG{}
		}
		photo := c.Message().Photo
		if photo == nil {
			errString := fmt.Sprintf("%s c.Message().Photo is null", funName)
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestReplaceBG{}
		}
		err, imageUrl := a.store.SaveMediaFile(photo.FileID)
		if err != nil {
			errString := fmt.Sprintf("%s a.store.SaveMediaFile Failed fileId=[%s]", funName, photo.FileID)
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestReplaceBG{}
		}
		log4plus.Info("%s imageUrl=[%s]", funName, imageUrl)
		request.BGUrl = imageUrl

		tmpData, err := json.Marshal(request)
		if err != nil {
			errString := fmt.Sprintf("%s json.Marshal Failed err=[%s]", funName, err.Error())
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestReplaceBG{}
		}

		var body header.RequestData
		body.BTCAddr = "0000000000000000000000000GFg7xJaNVN2"
		body.Data = json.RawMessage(tmpData)

		data, err := json.Marshal(body)
		if err != nil {
			errString := fmt.Sprintf("%s json.Marshal Failed err=[%s]", funName, err.Error())
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestReplaceBG{}
		}
		return nil, data, request
	}
	errString := fmt.Sprintf("%s current step Failed maxStep=[%d]", funName, user.MaxStep)
	log4plus.Info("%s errString=[%s]", funName, errString)
	return errors.New(errString), []byte(""), header.RequestReplaceBG{}
}

func (a *RemoveBG) replaceBG(c tele.Context) error {
	funName := "replaceBG"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s consumption time=%d(s)", funName, time.Now().Unix()-now)
	}()
	telegramID := c.Sender()
	log4plus.Info("%s telegramID=[%d]", funName, telegramID.ID)

	user := a.store.FindUser(telegramID.ID)
	if user == nil {
		log4plus.Info("%s first parseReplaceBG=[%d]", funName, telegramID.ID)
		a.parseReplaceBG(c, nil)
		return nil
	}
	log4plus.Info("%s user=[%+v]", funName, *user)

	if user.MaxStep == 0 {
		log4plus.Info("%s second parseReplaceBG=[%d]", funName, telegramID.ID)
		a.parseReplaceBG(c, user)
		return nil
	} else if user.MaxStep == 1 {
		log4plus.Info("%s third parseReplaceBG=[%d]", funName, telegramID.ID)
		err, body, _ := a.parseReplaceBG(c, user)
		if err != nil {
			errString := fmt.Sprintf("%s parseReplaceBG Failed telegramID=[%d]", funName, telegramID.ID)
			log4plus.Error(errString)
			a.store.GetBotObject().Send(c.Chat(), errString, &tele.SendOptions{
				ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
			})
			return err
		}
		apiKey := "123456"
		err, taskId := implementation.SingtonRemoveBG().ReplaceBG(funName, apiKey, body)
		if err != nil {
			a.store.GetBotObject().Send(c.Chat(), err.Error(), &tele.SendOptions{
				ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
			})
			log4plus.Error(err.Error())
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
				log4plus.Info("%s errString=[%s]", funName, errString)
				a.store.GetBotObject().Edit(taskMessage, errString, &tele.SendOptions{
					ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
				})
				return errors.New(errString)
			}
			log4plus.Info("%s DownloadFile newUrl=[%s] localPath=[%s]", funName, newUrl, localPath)

			var replyMessage []string
			replyMessage = append(replyMessage, newUrl)
			cmdlines := strings.Join(replyMessage, "\n")

			a.store.GetBotObject().Edit(taskMessage, cmdlines, &tele.SendOptions{
				ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
			})
			a.store.DeleteUser(telegramID.ID)
			return nil
		}
		errString := fmt.Sprintf("%s data is Failed data=[%s]", funName, resultData)
		log4plus.Error(errString)
		a.store.GetBotObject().Edit(taskMessage, errString, &tele.SendOptions{
			ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
		})
		return errors.New(errString)
	}
	errString := fmt.Sprintf("%s current step maxStep=[%d]", funName, user.MaxStep)
	log4plus.Error(errString)
	a.store.GetBotObject().Send(c.Chat(), errString, &tele.SendOptions{
		ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
	})
	return errors.New(errString)
}

func (a *RemoveBG) setFuncs() {
	a.store.SetTelegramFunction("removeBG", "removeBG", "🗑️ removeBG", a.removeBG)
	a.store.SetTelegramFunction("replaceBG", "replaceBG", "🔄 replaceBG", a.replaceBG)
}

func SingtonRemoveBG(store header.TelegramPluginStore) *RemoveBG {
	if nil == gRemoveBG {
		gRemoveBG = &RemoveBG{
			store: store,
		}
		gRemoveBG.setFuncs()
	}
	return gRemoveBG
}
