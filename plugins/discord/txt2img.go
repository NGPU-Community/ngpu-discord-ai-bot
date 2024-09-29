package discord

import (
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/nGPU/bot/common"
	"github.com/nGPU/bot/configure"
	"github.com/nGPU/bot/header"
	"github.com/nGPU/bot/implementation"
	log4plus "github.com/nGPU/common/log4go"
)

type Txt2Img struct {
	roots        *x509.CertPool
	rootPEM      []byte
	store        header.DiscordPluginStore
	commandLines []*header.CommandLine
}

var gTxt2Img *Txt2Img

func (a *Txt2Img) setCommand() {
	createtweetCmd := &discordgo.ApplicationCommand{
		Name:        "txt2img",
		Description: fmt.Sprintf("%s Text dubbing", configure.SingtonConfigure().Application.Name),
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "prompt",
				Description: "using custom prompt",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			}, {
				Name:        "width",
				Description: "Please select image width",
				Type:        discordgo.ApplicationCommandOptionInteger,
				Required:    false,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{
						Name:  "512",
						Value: "521",
					},
					{
						Name:  "1024",
						Value: "1024",
					},
				},
			}, {
				Name:        "height",
				Description: "Please select image height",
				Type:        discordgo.ApplicationCommandOptionInteger,
				Required:    false,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{
						Name:  "512",
						Value: "521",
					},
					{
						Name:  "1024",
						Value: "1024",
					},
				},
			},
		},
	}
	a.store.RegisteredCommand(createtweetCmd)
}

func (a *Txt2Img) Command(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var content string = ""
	for _, value := range a.commandLines {
		content = makeContent(value.Command, value.En, value.Zh)
	}

	/*Subscription 的指令*/
	discordId := common.GetUserId(s, i)
	disabledPayment := false
	err, exist, user := a.store.GetUserBase(discordId)
	if err != nil {
		disabledPayment = true
	}
	if !exist {
		disabledPayment = true
	} else {
		disabledPayment = false
	}
	header.MonthPaymentButton.URL = header.MakePaymentUrl(configure.SingtonConfigure().Stripe.PaymentUrl, discordId, header.MonthPrice, user.EMail)
	header.MonthPaymentButton.Disabled = disabledPayment

	header.YearPaymentButton.URL = header.MakePaymentUrl(configure.SingtonConfigure().Stripe.PaymentUrl, discordId, header.YearPrice, user.EMail)
	header.YearPaymentButton.Disabled = disabledPayment

	// 构建交互回应数据
	data := makeData("make/help Command", "Command Description", content, common.GetColor())
	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: data,
	}
	if err = s.InteractionRespond(i.Interaction, response); err != nil {
		log4plus.Error("Embed responseMsg Failed err=%s", err.Error())
	}
}

func (a *Txt2Img) parseTxt2Img(i *discordgo.InteractionCreate) (error, []byte, header.RequestTxt2Img) {
	funName := "parseTxt2Img"
	var prompt string = ""
	var width int64 = 0
	var height int64 = 0
	if i.Type == discordgo.InteractionApplicationCommand {
		command := i.ApplicationCommandData()
		for _, v := range command.Options {
			if v.Type == discordgo.ApplicationCommandOptionString {
				if v.Name == "prompt" {
					prompt = v.StringValue()
					log4plus.Info("%s prompt=[%s]", funName, prompt)
				}
			} else if v.Type == discordgo.ApplicationCommandOptionInteger {
				if v.Name == "width" {
					width = v.IntValue()
					log4plus.Info("%s width=[%d]", funName, width)
				} else if v.Name == "height" {

					height = v.IntValue()
					log4plus.Info("%s height=[%d]", funName, height)
				}
			}
		}
	}

	request := header.RequestTxt2Img{
		Prompt: prompt,
		Width:  int(width),
		Height: int(height),
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

func (a *Txt2Img) txt2img(s *discordgo.Session, i *discordgo.InteractionCreate) {
	funName := "txt2img"
	cmdName := "/txt2img"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s consumption time=%d(s)", funName, time.Now().Unix()-now)
	}()

	//获取用户id
	discordID := common.GetUserId(s, i)
	log4plus.Info("%s discordID=[%s]", funName, discordID)

	err, body, txt2ImgBody := a.parseTxt2Img(i)
	if err != nil {
		errString := fmt.Sprintf("%s parseTxt2Img Failed discordID=[%s]", funName, discordID)
		log4plus.Error(errString)
		common.SetCommandErrorResult(fmt.Sprintf("%s\n", cmdName), cmdName, errString, s, i)
		return
	}
	apiKey := "123456"

	a.setFirst(fmt.Sprintf("%s\n", cmdName), s, i)
	err, prompt, txt := implementation.SingtonTxt2Img().Txt2Img(funName, apiKey, body)
	if err != nil {
		a.setAnswerError(cmdName, err.Error(), s, i)
		return
	}

	var cmdlines []string
	cmdlines = append(cmdlines, fmt.Sprintf("%s", cmdName))
	cmdlines = append(cmdlines, fmt.Sprintf("original prompt: %s", txt2ImgBody.Prompt))
	cmdlines = append(cmdlines, fmt.Sprintf("expanded prompt: %s", prompt))
	cmdlines = append(cmdlines, fmt.Sprintf("%s", txt))
	cmdlines = append(cmdlines, fmt.Sprintf("For detailed interface explanations, please refer to:%s", configure.SingtonConfigure().Interfaces.Txt2Img.Urls.Txt2img.Comment))
	a.setFirstComplete(cmdlines, s, i)
	cmdlines = cmdlines[:0]
}

func (a *Txt2Img) setFuncs() {
	a.store.SetDiscordFunction("txt2img", a.txt2img)
}

func (a *Txt2Img) initCommand() {
	a.commandLines = append(a.commandLines, &header.CommandLine{
		Command: "txt2img",
		En:      "The /txt2img command can generate an image based on the prompt you provide.",
		Zh:      "txt2img命令，可以根据您输入的prompt生成相应的图片",
	})
	a.store.AddCommandLine(a.commandLines)
}

func (a *Txt2Img) setFirst(txt string, s *discordgo.Session, i *discordgo.InteractionCreate) string {
	funName := "setFirst"
	content := fmt.Sprintf("%s", txt)
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	}); err != nil {
		log4plus.Error("%s InteractionRespond Failed txt=[%s] err=[%s]", funName, txt, err.Error())
		if _, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: content,
		}); err != nil {
			log4plus.Error("%s FollowupMessageCreate Failed err=[%s]", funName, err.Error())
		}
		return ""
	}
	return content
}

func (a *Txt2Img) setFirstComplete(txts []string, s *discordgo.Session, i *discordgo.InteractionCreate) string {
	funName := "setFirstComplete"
	content := strings.Join(txts, "\n")
	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	}); err != nil {
		log4plus.Error("%s InteractionResponseEdit Failed txt=[%s] err=[%s]", funName, content, err.Error())
		if _, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: content,
		}); err != nil {
			log4plus.Error("%s FollowupMessageCreate Failed err=[%s]", funName, err.Error())
		}
		return ""
	}
	return content
}

func (a *Txt2Img) setAnswerError(cmd string, err string, s *discordgo.Session, i *discordgo.InteractionCreate) string {
	funName := "setAnswerError"
	content := fmt.Sprintf("%s \u274C %s", cmd, err)
	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	}); err != nil {
		log4plus.Error("%s InteractionResponseEdit Failed err=[%s]", funName, err.Error())
		if _, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: content,
		}); err != nil {
			log4plus.Error("%s FollowupMessageCreate Failed err=[%s]", funName, err.Error())
		}
		return ""
	}
	return content
}

func SingtonStableDiffusion(store header.DiscordPluginStore) *Txt2Img {
	if nil == gTxt2Img {
		gTxt2Img = &Txt2Img{
			store: store,
		}
		gTxt2Img.setCommand()
		gTxt2Img.setFuncs()
		gTxt2Img.initCommand()
	}
	return gTxt2Img
}
