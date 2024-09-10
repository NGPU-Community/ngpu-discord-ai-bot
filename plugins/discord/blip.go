package discord

import (
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	log4plus "github.com/nGPU/common/log4go"
	"github.com/nGPU/discordBot/common"
	"github.com/nGPU/discordBot/configure"
	"github.com/nGPU/discordBot/header"
	"github.com/nGPU/discordBot/implementation"
)

type Blip struct {
	roots        *x509.CertPool
	rootPEM      []byte
	store        header.PluginStore
	commandLines []*header.CommandLine
}

var gBlip *Blip

func (a *Blip) setCommand() {
	createtweetCmd := &discordgo.ApplicationCommand{
		Name:        "blip",
		Description: fmt.Sprintf("%s Text dubbing", configure.SingtonConfigure().Application.Name),
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "image",
				Description: "Please select the image you want to use",
				Type:        discordgo.ApplicationCommandOptionAttachment,
				Required:    true,
			},
		},
	}
	a.store.RegisteredCommand(createtweetCmd)
}

func (a *Blip) Command(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

func (a *Blip) parseBlip(i *discordgo.InteractionCreate) (error, []byte, header.RequestImg2Txt) {
	funName := "parseBlip"
	var imageUrl string
	if i.Type == discordgo.InteractionApplicationCommand {
		command := i.ApplicationCommandData()
		for _, v := range command.Options {
			if v.Type == discordgo.ApplicationCommandOptionAttachment {
				for _, value := range command.Resolved.Attachments {
					imageUrl = value.URL
					log4plus.Info("%s command.ID=[%s] value=[%s]", funName, command.ID, value.URL)
				}
			}
		}
	}
	if strings.Trim(imageUrl, " ") == "" {
		errString := fmt.Sprintf("%s Failed imageUrl is Empty", funName)
		log4plus.Info("%s errString=[%s]", funName, errString)
		return errors.New(errString), []byte(""), header.RequestImg2Txt{}
	}
	err, newUrl, localPath := a.store.DownloadFile(imageUrl)
	if err != nil {
		errString := fmt.Sprintf("%s DownloadFile sourceUrl=[%s]", funName, imageUrl)
		log4plus.Info("%s errString=[%s]", funName, errString)
		return errors.New(errString), []byte(""), header.RequestImg2Txt{}
	}
	log4plus.Info("%s DownloadFile newUrl=[%s] localPath=[%s]", funName, newUrl, localPath)

	input := header.BaseImg2Txt{
		Task:  "image_captioning",
		Image: newUrl,
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

func (a *Blip) blip(s *discordgo.Session, i *discordgo.InteractionCreate) {
	funName := "blip"
	cmdName := "/blip"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s consumption time=%d(s)", funName, time.Now().Unix()-now)
	}()

	//获取用户id
	discordID := common.GetUserId(s, i)
	log4plus.Info("%s discordID=[%s]", funName, discordID)

	err, body, blipBody := a.parseBlip(i)
	if err != nil {
		errString := fmt.Sprintf("%s parseBlip Failed discordID=[%s]", funName, discordID)
		log4plus.Error(errString)
		common.SetCommandErrorResult(fmt.Sprintf("%s\n", cmdName), cmdName, errString, s, i)
		return
	}
	apiKey := "123456"

	a.setFirst(fmt.Sprintf("%s\n", cmdName), s, i)
	err, txt := implementation.SingtonBlip().Blip(funName, apiKey, body)
	if err != nil {
		a.setAnswerError(cmdName, err.Error(), s, i)
		return
	}

	var cmdlines []string
	cmdlines = append(cmdlines, fmt.Sprintf("%s", cmdName))
	cmdlines = append(cmdlines, fmt.Sprintf("source:%s", blipBody.Input.Image))
	cmdlines = append(cmdlines, fmt.Sprintf("%s", txt))
	cmdlines = append(cmdlines, fmt.Sprintf("For detailed interface explanations, please refer to:%s", configure.SingtonConfigure().Interfaces.Blip.Urls.Blip.Comment))
	a.setFirstComplete(cmdlines, s, i)
	cmdlines = cmdlines[:0]
}

func (a *Blip) setFuncs() {
	a.store.SetDiscordFunction("blip", a.blip)
}

func (a *Blip) initCommand() {
	a.commandLines = append(a.commandLines, &header.CommandLine{
		Command: "blip",
		En:      "The /blip command can provide an explanation for the image you upload.",
		Zh:      "blip命令，可以对您上传的图片进行解释",
	})
	a.store.AddCommandLine(a.commandLines)
}

func (a *Blip) setFirst(txt string, s *discordgo.Session, i *discordgo.InteractionCreate) string {
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

func (a *Blip) setFirstComplete(txts []string, s *discordgo.Session, i *discordgo.InteractionCreate) string {
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

func (a *Blip) setAnswerError(cmd string, err string, s *discordgo.Session, i *discordgo.InteractionCreate) string {
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

func SingtonBlip(store header.PluginStore) *Blip {
	if nil == gBlip {
		gBlip = &Blip{
			store: store,
		}
		gBlip.setCommand()
		gBlip.setFuncs()
		gBlip.initCommand()
	}
	return gBlip
}
