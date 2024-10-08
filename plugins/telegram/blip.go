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
	prompt = [...]string{
		"Please upload the image you need to analyze: ",
	}
)

type Blip struct {
	roots        *x509.CertPool
	rootPEM      []byte
	store        header.TelegramPluginStore
	commandLines []*header.CommandLine
}

var gBlip *Blip

func (a *Blip) parseBlip(c tele.Context, user *header.UserStep) (error, []byte, header.RequestImg2Txt) {
	funName := "parseBlip"
	if user == nil {
		user := &header.UserStep{
			TelegramId:   c.Sender().ID,
			MaxStep:      0,
			FunctionName: "blip",
			MessageId:    c.Message().ID,
		}
		a.store.AddUser(user)
		a.store.GetBotObject().Send(c.Chat(), prompt[user.MaxStep], &tele.SendOptions{
			ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
		})
		return nil, []byte(""), header.RequestImg2Txt{}
	}
	if user.MaxStep == 0 {
		photo := c.Message().Photo
		if photo == nil {
			errString := fmt.Sprintf("%s c.Message().Photo is null", funName)
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestImg2Txt{}
		}
		err, imageUrl := a.store.SaveMediaFile(photo.FileID)
		if err != nil {
			errString := fmt.Sprintf("%s a.store.SaveMediaFile Failed fileId=[%s]", funName, photo.FileID)
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestImg2Txt{}
		}
		log4plus.Info("%s imageUrl=[%s]", funName, imageUrl)

		input := header.BaseImg2Txt{
			Task:  "image_captioning",
			Image: imageUrl,
		}
		request := header.RequestImg2Txt{
			Input: input,
		}

		tmpData, err := json.Marshal(request)
		if err != nil {
			errString := fmt.Sprintf("%s json.Marshal Failed err=[%s]", funName, err.Error())
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestImg2Txt{}
		}
		var body header.RequestData
		body.BTCAddr = "0000000000000000000000000GFg7xJaNVN2"
		body.Data = json.RawMessage(tmpData)

		data, err := json.Marshal(body)
		if err != nil {
			errString := fmt.Sprintf("%s json.Marshal Failed err=[%s]", funName, err.Error())
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestImg2Txt{}
		}
		return nil, data, request
	}
	errString := fmt.Sprintf("%s Step Failed current Step is one", funName)
	log4plus.Info("%s errString=[%s]", funName, errString)
	return errors.New(errString), []byte(""), header.RequestImg2Txt{}
}

func (a *Blip) blip(c tele.Context) error {
	funName := "blip"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s consumption time=%d(s)", funName, time.Now().Unix()-now)
	}()
	telegramID := c.Sender()
	log4plus.Info("%s telegramID=[%d]", funName, telegramID.ID)

	user := a.store.FindUser(telegramID.ID)
	if user == nil {
		a.parseBlip(c, nil)
		return nil
	}
	err, body, _ := a.parseBlip(c, user)
	if err != nil {
		errString := fmt.Sprintf("%s parseBlip Failed telegramID=[%d]", funName, telegramID.ID)
		log4plus.Error(errString)
		a.store.GetBotObject().Send(c.Chat(), errString, &tele.SendOptions{
			ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
		})
		return err
	}
	apiKey := "123456"
	err, txt := implementation.SingtonBlip().Blip(funName, apiKey, body)
	if err != nil {
		a.store.GetBotObject().Send(c.Chat(), err.Error(), &tele.SendOptions{
			ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
		})
		log4plus.Error("%s err=[%s]", funName, err.Error())
		return err
	}
	var replyMessage []string
	replyMessage = append(replyMessage, fmt.Sprintf("analysis words: %s", txt))
	cmdlines := strings.Join(replyMessage, "\n")
	a.store.GetBotObject().Send(c.Chat(), cmdlines, &tele.SendOptions{
		ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
	})
	a.store.DeleteUser(telegramID.ID)
	return nil
}

func (a *Blip) setFuncs() {
	a.store.SetTelegramFunction("blip", "blip", "🖼️ blip", a.blip)
}

func SingtonBlip(store header.TelegramPluginStore) *Blip {
	if nil == gBlip {
		gBlip = &Blip{
			store: store,
		}
		gBlip.setFuncs()
	}
	return gBlip
}
