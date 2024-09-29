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
	txt2imgPrompt = [...]string{
		"Please enter the descriptive words for the image you want to generate: ",
	}
)

type Txt2Img struct {
	roots        *x509.CertPool
	rootPEM      []byte
	store        header.TelegramPluginStore
	commandLines []*header.CommandLine
}

var gTxt2Img *Txt2Img

func (a *Txt2Img) parseTxt2Img(c tele.Context, user *header.UserStep) (error, []byte, header.RequestTxt2Img) {
	funName := "parseTxt2Img"
	if user == nil {
		user := &header.UserStep{
			TelegramId:   c.Sender().ID,
			MaxStep:      0,
			FunctionName: "txt2img",
			MessageId:    c.Message().ID,
		}
		a.store.AddUser(user)
		a.store.GetBotObject().Send(c.Chat(), txt2imgPrompt[user.MaxStep], &tele.SendOptions{
			ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
		})
		return nil, []byte(""), header.RequestTxt2Img{}
	}
	if user.MaxStep == 0 {
		prompt := c.Text()
		log4plus.Info("%s prompt=[%s]", funName, prompt)

		request := header.RequestTxt2Img{
			Prompt: prompt,
			Width:  int(512),
			Height: int(512),
		}
		tmpData, err := json.Marshal(request)
		if err != nil {
			errString := fmt.Sprintf("%s json.Marshal Failed err=[%s]", funName, err.Error())
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestTxt2Img{}
		}

		var body header.RequestData
		body.BTCAddr = "0000000000000000000000000GFg7xJaNVN2"
		body.Data = json.RawMessage(tmpData)

		data, err := json.Marshal(body)
		if err != nil {
			errString := fmt.Sprintf("%s json.Marshal Failed err=[%s]", funName, err.Error())
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestTxt2Img{}
		}
		return nil, data, request
	}
	errString := fmt.Sprintf("%s Step Failed current Step is one", funName)
	log4plus.Info("%s errString=[%s]", funName, errString)
	return errors.New(errString), []byte(""), header.RequestTxt2Img{}
}

func (a *Txt2Img) txt2img(c tele.Context) error {
	funName := "txt2img"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s consumption time=%d(s)", funName, time.Now().Unix()-now)
	}()
	telegramID := c.Sender()
	log4plus.Info("%s telegramID=[%d]", funName, telegramID.ID)

	user := a.store.FindUser(telegramID.ID)
	if user == nil {
		a.parseTxt2Img(c, nil)
		return nil
	}
	err, body, txt2ImgBody := a.parseTxt2Img(c, user)
	if err != nil {
		errString := fmt.Sprintf("%s parseTxt2Img Failed telegramID=[%d]", funName, telegramID.ID)
		log4plus.Error(errString)
		a.store.GetBotObject().Send(c.Chat(), errString, &tele.SendOptions{
			ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
		})
		return err
	}
	apiKey := "123456"
	err, prompt, txt := implementation.SingtonTxt2Img().Txt2Img(funName, apiKey, body)
	if err != nil {
		a.store.GetBotObject().Send(c.Chat(), err.Error(), &tele.SendOptions{
			ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
		})
		log4plus.Error("%s err=[%s]", funName, err.Error())
		return err
	}

	var replyMessage []string
	replyMessage = append(replyMessage, fmt.Sprintf("original prompt: %s", txt2ImgBody.Prompt))
	replyMessage = append(replyMessage, fmt.Sprintf("expanded prompt: %s", prompt))
	replyMessage = append(replyMessage, fmt.Sprintf("%s", txt))
	cmdlines := strings.Join(replyMessage, "\n")
	a.store.GetBotObject().Send(c.Chat(), cmdlines, &tele.SendOptions{
		ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
	})
	a.store.DeleteUser(telegramID.ID)
	return nil
}

func (a *Txt2Img) setFuncs() {
	a.store.SetTelegramFunction("txt2img", "txt2img", "🔥 txt2img", a.txt2img)
}

func SingtonTxt2Img(store header.TelegramPluginStore) *Txt2Img {
	if nil == gTxt2Img {
		gTxt2Img = &Txt2Img{
			store: store,
		}
		gTxt2Img.setFuncs()
	}
	return gTxt2Img
}
