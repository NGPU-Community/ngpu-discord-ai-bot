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
	promptOllma = [...]string{
		"Please upload the image you need to analyze: ",
	}
)

type Ollma struct {
	roots        *x509.CertPool
	rootPEM      []byte
	store        header.TelegramPluginStore
	commandLines []*header.CommandLine
}

var gOllma *Ollma

func (a *Ollma) parseChat(c tele.Context, user *header.UserStep) (error, []byte, header.RequestLLM) {
	funName := "parseChat"
	if user == nil {
		user := &header.UserStep{
			TelegramId:   c.Sender().ID,
			MaxStep:      0,
			FunctionName: "chat",
			MessageId:    c.Message().ID,
		}
		a.store.AddUser(user)
		a.store.GetBotObject().Send(c.Chat(), promptOllma[user.MaxStep], &tele.SendOptions{
			ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
		})
		return nil, []byte(""), header.RequestLLM{}
	}
	if user.MaxStep == 0 {
		prompt := c.Text()
		log4plus.Info("%s prompt=[%s]", funName, prompt)

		request := header.RequestLLM{
			Prompt: prompt,
		}
		tmpData, err := json.Marshal(request)
		if err != nil {
			errString := fmt.Sprintf("%s json.Marshal Failed err=[%s]", funName, err.Error())
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestLLM{}
		}
		var body header.RequestData
		body.BTCAddr = "0000000000000000000000000GFg7xJaNVN2"
		body.Data = json.RawMessage(tmpData)

		data, err := json.Marshal(body)
		if err != nil {
			errString := fmt.Sprintf("%s json.Marshal Failed err=[%s]", funName, err.Error())
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestLLM{}
		}
		return nil, data, request
	}
	errString := fmt.Sprintf("%s Step Failed current Step is one", funName)
	log4plus.Info("%s errString=[%s]", funName, errString)
	return errors.New(errString), []byte(""), header.RequestLLM{}
}

func (a *Ollma) chat(c tele.Context) error {
	funName := "chat"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s consumption time=%d(s)", funName, time.Now().Unix()-now)
	}()
	log4plus.Info("%s telegramID=[%d]", funName, c.Sender().ID)

	user := a.store.FindUser(c.Sender().ID)
	if user == nil {
		a.parseChat(c, nil)
		return nil
	}
	err, body, _ := a.parseChat(c, user)
	if err != nil {
		errString := fmt.Sprintf("%s parseChat Failed telegramID=[%d]", funName, c.Sender().ID)
		log4plus.Error(errString)
		a.store.GetBotObject().Send(c.Chat(), errString, &tele.SendOptions{
			ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
		})
		return err
	}
	apiKey := "123456"
	err, txt := implementation.SingtonLlm().Chat(funName, apiKey, body)
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
	a.store.DeleteUser(c.Sender().ID)
	return nil
}

func (a *Ollma) setFuncs() {
	a.store.SetTelegramFunction("chat", "chat", "🗣️ chat", a.chat)
}

func SingtonOllma(store header.TelegramPluginStore) *Ollma {
	if nil == gOllma {
		gOllma = &Ollma{
			store: store,
		}
		gOllma.setFuncs()
	}
	return gOllma
}
