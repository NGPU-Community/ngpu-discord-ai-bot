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

type SadTalker struct {
	roots        *x509.CertPool
	rootPEM      []byte
	store        header.DiscordPluginStore
	commandLines []*header.CommandLine
}

var gSadTalker *SadTalker

func (a *SadTalker) setCommand() {
	sadtalkerCmd := &discordgo.ApplicationCommand{
		Name:        "sadtalker",
		Description: fmt.Sprintf("%s Text dubbing", configure.SingtonConfigure().Application.Name),
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "image",
				Description: "Please select the source image you want to use",
				Type:        discordgo.ApplicationCommandOptionAttachment,
				Required:    true,
			}, {
				Name:        "text",
				Description: "using custom strength",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			}, {
				Name:        "pronouncer",
				Description: "using custom strength",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{
						Name:  "Englist(USA)-Guy-Male",
						Value: "en-US-GuyNeural",
					},
					{
						Name:  "zh-CN-XiaoXiaoNeural-Female",
						Value: "zh-CN-XiaoxiaoNeural",
					},
					{
						Name:  "zh-CN-YunxiNeural-Male",
						Value: "zh-CN-YunxiNeural",
					},
					{
						Name:  "zh-HK-HiuGaaiNeural-Female",
						Value: "zh-HK-HiuGaaiNeural",
					},
				},
			}, {
				Name:        "background",
				Description: "Please select the video background",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    false,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{
						Name:  "tourism",
						Value: "https://ngpu.ai/anvd/background/tourism.png",
					},
					{
						Name:  "AINN",
						Value: "https://ngpu.ai/anvd/background/AINN_BG.png",
					},
					{
						Name:  "coins",
						Value: "https://ngpu.ai/anvd/background/coins.png",
					},
				},
			}, {
				Name:        "logo",
				Description: "using custom strength",
				Type:        discordgo.ApplicationCommandOptionAttachment,
				Required:    false,
			},
		},
	}
	a.store.RegisteredCommand(sadtalkerCmd)
}

func (a *SadTalker) Command(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var content string = ""
	for _, value := range a.commandLines {
		content = makeContent(value.Command, value.En, value.Zh)
	}

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

	data := makeData("make/help Command", "Command Description", content, common.GetColor())
	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: data,
	}
	if err = s.InteractionRespond(i.Interaction, response); err != nil {
		log4plus.Error("Embed responseMsg Failed err=%s", err.Error())
	}
}

func (a *SadTalker) getAttachment(s *discordgo.Session, i *discordgo.InteractionCreate, name string) string {
	for _, option := range i.ApplicationCommandData().Options {
		if option.Type == discordgo.ApplicationCommandOptionAttachment {
			if name == option.Name {
				attachmentId := option.Value.(string)
				attachment := i.ApplicationCommandData().Resolved.Attachments[attachmentId]
				return attachment.URL
			}
		}
	}
	return ""
}

func (a *SadTalker) parseSadTalker(s *discordgo.Session, i *discordgo.InteractionCreate) (error, []byte, header.RequestSadTalker) {
	funName := "parseSadTalker"

	backgroundName := ""
	logoUrl := ""
	var imageUrl string
	var text string
	var pronouncer string
	if i.Type == discordgo.InteractionApplicationCommand {
		command := i.ApplicationCommandData()
		for _, v := range command.Options {
			if v.Type == discordgo.ApplicationCommandOptionAttachment {
				if strings.ToLower(v.Name) == strings.ToLower("image") {
					imageUrl = a.getAttachment(s, i, strings.ToLower("image"))
				} else if strings.ToLower(v.Name) == strings.ToLower("logo") {
					logo := a.getAttachment(s, i, strings.ToLower("logo"))
					if strings.Trim(logo, " ") == "" {
						logoUrl = ""
					}
				}
			} else if v.Type == discordgo.ApplicationCommandOptionString {
				if v.Name == "text" {
					text = v.StringValue()
					log4plus.Info("%s text=[%s]", funName, text)
				} else if v.Name == "pronouncer" {
					pronouncer = v.StringValue()
					log4plus.Info("%s pronouncer=[%s]", funName, pronouncer)
				} else if strings.ToLower(v.Name) == strings.ToLower("background") {
					backgroundName = v.StringValue()
				}
			}
		}
	}
	if strings.Trim(imageUrl, " ") == "" {
		errString := fmt.Sprintf("%s Failed imageUrl is Empty", funName)
		log4plus.Info("%s errString=[%s]", funName, errString)
		return errors.New(errString), []byte(""), header.RequestSadTalker{}
	}

	err, newImageUrl, localImagePath := a.store.DownloadFile(imageUrl)
	if err != nil {
		errString := fmt.Sprintf("%s DownloadFile imageUrl=[%s]", funName, imageUrl)
		log4plus.Info("%s errString=[%s]", funName, errString)
		return errors.New(errString), []byte(""), header.RequestSadTalker{}
	}
	log4plus.Info("%s DownloadFile newImageUrl=[%s] localImagePath=[%s]", funName, newImageUrl, localImagePath)

	if len(logoUrl) == 0 {
		logoUrl = ""
	} else {
		err, logoUrl, _ := a.store.DownloadFile(logoUrl)
		if err != nil {
			errString := fmt.Sprintf("%s DownloadFile logoUrl=[%s]", funName, logoUrl)
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestSadTalker{}
		}
	}
	log4plus.Info("%s DownloadFile logoUrl=[%s]", funName, logoUrl)
	request := header.RequestSadTalker{
		ImageUrl:       newImageUrl,
		Text:           text,
		Pronouncer:     pronouncer,
		BackGroundName: backgroundName,
		LogoUrl:        logoUrl,
	}

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

func (a *SadTalker) sadTalker(s *discordgo.Session, i *discordgo.InteractionCreate) {
	funName := "sadTalker"
	cmdName := "/sadTalker"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s consumption time=%d(s)", funName, time.Now().Unix()-now)
	}()
	discordID := common.GetUserId(s, i)
	log4plus.Info("%s discordID=[%s]", funName, discordID)

	err, body, sadTalkerBody := a.parseSadTalker(s, i)
	if err != nil {
		errString := fmt.Sprintf("%s parseSadTalker Failed discordID=[%s]", funName, discordID)
		log4plus.Error(errString)
		common.SetCommandErrorResult(fmt.Sprintf("%s\n", cmdName), cmdName, errString, s, i)
		return
	}
	apiKey := "123456"
	a.setFirst(fmt.Sprintf("%s\n", cmdName), s, i)
	err, taskId := implementation.SingtonSadTalker().SadTalker(funName, apiKey, body)
	if err != nil {
		a.setAnswerError(cmdName, err.Error(), s, i)
		return
	}

	var cmdlines []string
	cmdlines = append(cmdlines, fmt.Sprintf("%s", cmdName))
	cmdlines = append(cmdlines, fmt.Sprintf("taskId=%s", taskId))
	a.setFirstComplete(cmdlines, s, i)
	cmdlines = cmdlines[:0]

	err, resultData := a.store.WaitingTaskId(taskId)
	if err != nil {
		cmdlines = append(cmdlines, fmt.Sprintf("%s", cmdName))
		cmdlines = append(cmdlines, fmt.Sprintf("err=%s", err.Error()))
		a.setFirstComplete(cmdlines, s, i)
		cmdlines = cmdlines[:0]
		return
	} else {
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
			a.setAnswerError(cmdName, errString, s, i)
			return
		}
		if result.Data.ResultCode == nil {
			errString := fmt.Sprintf("%s data is nil", funName)
			log4plus.Error(errString)
			a.setAnswerError(cmdName, errString, s, i)
			return
		} else if int(result.Data.ResultCode.(float64)) == 100 {
			imageUrl := result.Data.ResultUrl.(string)

			//这里下载上传的源文件形成新的文件地址
			err, newUrl, localPath := a.store.DownloadFile(imageUrl)
			if err != nil {
				errString := fmt.Sprintf("%s DownloadFile newUrl=[%s]", funName, newUrl)
				log4plus.Info("%s errString=[%s]", funName, errString)
				a.setAnswerError(cmdName, errString, s, i)
				return
			}
			log4plus.Info("%s DownloadFile newUrl=[%s] localPath=[%s]", funName, newUrl, localPath)

			cmdlines = append(cmdlines, fmt.Sprintf("%s", cmdName))
			cmdlines = append(cmdlines, fmt.Sprintf("original image: %s", sadTalkerBody.ImageUrl))
			cmdlines = append(cmdlines, fmt.Sprintf("voice content: %s", sadTalkerBody.Text))
			cmdlines = append(cmdlines, fmt.Sprintf("generate video: %s", newUrl))
			cmdlines = append(cmdlines, fmt.Sprintf("For detailed interface explanations, please refer to:%s",
				configure.SingtonConfigure().Interfaces.SadTalker.Urls.SadTalker.Comment))
			a.setFirstComplete(cmdlines, s, i)
		} else {
			errString := fmt.Sprintf("%s result_code is not 100 result_code=[%d] data=[%s]", funName, int(result.Data.ResultCode.(float64)), resultData)
			log4plus.Error(errString)
			a.setAnswerError(cmdName, errString, s, i)
			return
		}
	}
}

func (a *SadTalker) setFuncs() {
	a.store.SetDiscordFunction("sadtalker", a.sadTalker)
}

func (a *SadTalker) initCommand() {
	a.commandLines = append(a.commandLines, &header.CommandLine{
		Command: "sadtalker",
		En:      "The /sadtalker command allows you to generate a corresponding video with just a single photo",
		Zh:      "sadtalker命令，您只需要一张照片，就可以生成相应的视频",
	})
	a.store.AddCommandLine(a.commandLines)
}

func (a *SadTalker) setFirst(txt string, s *discordgo.Session, i *discordgo.InteractionCreate) string {
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

func (a *SadTalker) setFirstComplete(txts []string, s *discordgo.Session, i *discordgo.InteractionCreate) string {
	funName := "setFirstComplete"
	content := strings.Join(txts, "\n")
	log4plus.Error("%s content=[%s]", funName, content)

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

func (a *SadTalker) setAnswerError(cmd string, err string, s *discordgo.Session, i *discordgo.InteractionCreate) string {
	funName := "setAnswerError"
	content := fmt.Sprintf("%s \nu274C %s", cmd, err)
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

func SingtonSadTalker(store header.DiscordPluginStore) *SadTalker {
	if nil == gSadTalker {
		gSadTalker = &SadTalker{
			store: store,
		}
		gSadTalker.setCommand()
		gSadTalker.setFuncs()
		gSadTalker.initCommand()
	}
	return gSadTalker
}
