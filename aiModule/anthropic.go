package aiModule

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/liushuangls/go-anthropic/v2"
	"github.com/nGPU/bot/configure"
	log4plus "github.com/nGPU/common/log4go"
)

type Anthropic struct {
	roots   *x509.CertPool
	rootPEM []byte
	client  *anthropic.Client
}

var gAnthropic *Anthropic

func (a *Anthropic) FengShui(imagePath, imageType, question string) (error, string) {
	funName := "FengShui"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s consumption time=%d(s)", funName, time.Now().Unix()-now)
	}()

	imageMediaType := imageType
	imageFile, err := os.Open(imagePath)
	if err != nil {
		errString := fmt.Sprintf("%s Open Failed err=[%s]", funName, err.Error())
		log4plus.Error(errString)
		return err, errString
	}
	imageData, err := io.ReadAll(imageFile)
	if err != nil {
		errString := fmt.Sprintf("%s ReadAll Failed err=[%s]", funName, err.Error())
		log4plus.Error(errString)
		return err, errString
	}

	resp, err := a.client.CreateMessages(context.Background(), anthropic.MessagesRequest{
		Model: anthropic.ModelClaude3Haiku20240307,
		Messages: []anthropic.Message{
			{
				Role: anthropic.RoleUser,
				Content: []anthropic.MessageContent{
					anthropic.NewImageMessageContent(anthropic.MessageContentImageSource{
						Type:      "base64",
						MediaType: imageMediaType,
						Data:      imageData,
					}),
					anthropic.NewTextMessageContent(question),
				},
			},
		},
		MaxTokens: 4000,
	})

	if err != nil {
		var e *anthropic.APIError
		if errors.As(err, &e) {
			errString := fmt.Sprintf("%s CreateMessages Failed type=[%s] message=[%s]", funName, e.Type, e.Message)
			log4plus.Error(errString)
			return err, errString
		} else {
			errString := fmt.Sprintf("%s CreateMessages Failed err=[%s]", funName, err.Error())
			log4plus.Error(errString)
			return err, errString
		}
	}
	return nil, *resp.Content[0].Text
}

func (a *Anthropic) Chat(prompt string) (error, string) {
	funName := "Chat"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s consumption time=%d(s)", funName, time.Now().Unix()-now)
	}()

	resp, err := a.client.CreateMessages(context.Background(), anthropic.MessagesRequest{
		Model: anthropic.ModelClaude3Haiku20240307,
		Messages: []anthropic.Message{
			anthropic.NewUserTextMessage(prompt),
		},
		MaxTokens: 1000,
	})
	if err != nil {
		var e *anthropic.APIError
		if errors.As(err, &e) {
			errString := fmt.Sprintf("%s CreateMessages Failed type=[%s] message=[%s]", funName, e.Type, e.Message)
			log4plus.Error(errString)
			return err, errString
		} else {
			errString := fmt.Sprintf("%s CreateMessages Failed err=[%s]", funName, err.Error())
			log4plus.Error(errString)
			return err, errString
		}
	}
	fmt.Println(resp.Content[0].GetText())
	return nil, resp.Content[0].GetText()
}

func (a *Anthropic) init() error {
	funName := "init"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s consumption time=%d(s)", funName, time.Now().Unix()-now)
	}()

	client := anthropic.NewClient(configure.SingtonConfigure().AiModules.Anthropic.ApiKey)
	if client == nil {
		errString := fmt.Sprintf("%s NewClient Failed", funName)
		log4plus.Error(errString)
		return errors.New(errString)
	}
	a.client = client
	log4plus.Info("%s Anthropic Init Success configure.SingtonConfigure().AiModules.Anthropic.ApiKey=[%s]", funName, configure.SingtonConfigure().AiModules.Anthropic.ApiKey)

	return nil
}

func SingtonAnthropic() *Anthropic {
	if nil == gAnthropic {
		gAnthropic = &Anthropic{}
		gAnthropic.init()
	}
	return gAnthropic
}
