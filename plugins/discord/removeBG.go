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

type RemoveBG struct {
	roots        *x509.CertPool
	rootPEM      []byte
	store        header.DiscordPluginStore
	commandLines []*header.CommandLine
}

var gRemoveBG *RemoveBG

func (a *RemoveBG) setCommand() {
	removeBGCmd := &discordgo.ApplicationCommand{
		Name:        "removebg",
		Description: fmt.Sprintf("%s Text dubbing", configure.SingtonConfigure().Application.Name),
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "image",
				Description: "Please select the source image you want to use",
				Type:        discordgo.ApplicationCommandOptionAttachment,
				Required:    true,
			}, {
				Name:        "mode",
				Description: "Please select the video background",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{
						Name:  "human",
						Value: "human",
					},
					{
						Name:  "cloth",
						Value: "cloth",
					},
					{
						Name:  "any",
						Value: "any",
					},
				},
			},
		},
	}
	a.store.RegisteredCommand(removeBGCmd)

	replaceBGCmd := &discordgo.ApplicationCommand{
		Name:        "replacebg",
		Description: fmt.Sprintf("%s Text dubbing", configure.SingtonConfigure().Application.Name),
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "image",
				Description: "Please select the source image you want to use",
				Type:        discordgo.ApplicationCommandOptionAttachment,
				Required:    true,
			}, {
				Name:        "imagebg",
				Description: "Please select the source imagebg you want to use",
				Type:        discordgo.ApplicationCommandOptionAttachment,
				Required:    true,
			}, {
				Name:        "mode",
				Description: "Please select the video background",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{
						Name:  "human",
						Value: "human",
					},
					{
						Name:  "cloth",
						Value: "cloth",
					},
					{
						Name:  "any",
						Value: "any",
					},
				},
			},
		},
	}
	a.store.RegisteredCommand(replaceBGCmd)
}

func (a *RemoveBG) Command(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

func (a *RemoveBG) getAttachment(s *discordgo.Session, i *discordgo.InteractionCreate, name string) string {
	// 通过 name 属性识别每个选项
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

func (a *RemoveBG) parseRemoveBG(s *discordgo.Session, i *discordgo.InteractionCreate) (error, []byte, header.RequestRemoveBG) {
	funName := "parseRemoveBG"
	var imageUrl, mode string
	if i.Type == discordgo.InteractionApplicationCommand {
		command := i.ApplicationCommandData()
		for _, v := range command.Options {

			if v.Type == discordgo.ApplicationCommandOptionAttachment {
				for _, value := range command.Resolved.Attachments {
					imageUrl = value.URL
					log4plus.Info("%s command.ID=[%s] value=[%s]", funName, command.ID, value.URL)
				}
			} else if v.Type == discordgo.ApplicationCommandOptionString {
				if v.Name == "mode" {
					mode = v.StringValue()
					log4plus.Info("%s mode=[%s]", funName, mode)
				}
			}

		}
	}
	log4plus.Info("%s image=[%s] mode=[%s]", funName, imageUrl, mode)
	if strings.Trim(imageUrl, " ") == "" {
		errString := fmt.Sprintf("%s Failed imageUrl is Empty", funName)
		log4plus.Info("%s errString=[%s]", funName, errString)
		return errors.New(errString), []byte(""), header.RequestRemoveBG{}
	}

	err, newImageUrl, localImagePath := a.store.DownloadFile(imageUrl)
	if err != nil {
		errString := fmt.Sprintf("%s DownloadFile imageUrl=[%s]", funName, imageUrl)
		log4plus.Info("%s errString=[%s]", funName, errString)
		return errors.New(errString), []byte(""), header.RequestRemoveBG{}
	}
	log4plus.Info("%s DownloadFile newImageUrl=[%s] localImagePath=[%s]", funName, newImageUrl, localImagePath)

	request := header.RequestRemoveBG{
		Kind:     "photo",
		Obj:      mode,
		ImageUrl: newImageUrl,
		BGColor:  "0,0,0,0",
	}

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

func (a *RemoveBG) removeBG(s *discordgo.Session, i *discordgo.InteractionCreate) {
	funName := "removeBG"
	cmdName := "/removeBG"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s consumption time=%d(s)", funName, time.Now().Unix()-now)
	}()
	discordID := common.GetUserId(s, i)
	log4plus.Info("%s discordID=[%s]", funName, discordID)

	err, body, removeBGBody := a.parseRemoveBG(s, i)
	if err != nil {
		errString := fmt.Sprintf("%s parseSadTalker Failed discordID=[%s]", funName, discordID)
		log4plus.Error(errString)
		common.SetCommandErrorResult(fmt.Sprintf("%s\n", cmdName), cmdName, errString, s, i)
		return
	}
	apiKey := "123456"
	a.setFirst(fmt.Sprintf("%s\n", cmdName), s, i)
	err, taskId := implementation.SingtonRemoveBG().RemoveBG(funName, apiKey, body)
	if err != nil {
		a.setAnswerError(cmdName, err.Error(), s, i)
		return
	}

	var cmdlines []string
	cmdlines = append(cmdlines, fmt.Sprintf("%s", cmdName))
	cmdlines = append(cmdlines, fmt.Sprintf("taskId=%s", taskId))
	a.setFirstComplete(cmdlines, s, i)
	cmdlines = cmdlines[:0]

	//循环获取返回的数据
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
			cmdlines = append(cmdlines, fmt.Sprintf("target image: %s", removeBGBody.ImageUrl))
			cmdlines = append(cmdlines, fmt.Sprintf("obtained image: %s", newUrl))
			cmdlines = append(cmdlines, fmt.Sprintf("For detailed interface explanations, please refer to:%s",
				configure.SingtonConfigure().Interfaces.RemoveBG.Urls.RemoveBG.Comment))
			a.setFirstComplete(cmdlines, s, i)
		} else {
			errString := fmt.Sprintf("%s data is Failed data=[%s]", funName, resultData)
			log4plus.Error(errString)
			a.setAnswerError(cmdName, errString, s, i)
			return
		}
	}
}

/**********************/

func (a *RemoveBG) parseReplaceBG(s *discordgo.Session, i *discordgo.InteractionCreate) (error, []byte, header.RequestReplaceBG) {
	funName := "parseReplaceBG"
	var imageUrl, imageBGUrl, mode string
	if i.Type == discordgo.InteractionApplicationCommand {
		command := i.ApplicationCommandData()
		for _, v := range command.Options {
			if v.Type == discordgo.ApplicationCommandOptionAttachment {
				if strings.ToLower(v.Name) == strings.ToLower("image") {
					imageUrl = a.getAttachment(s, i, strings.ToLower("image"))
					log4plus.Info("%s command.ID=[%s] imageUrl=[%s]", funName, command.ID, imageUrl)

				} else if strings.ToLower(v.Name) == strings.ToLower("imagebg") {
					imageBGUrl = a.getAttachment(s, i, strings.ToLower("imagebg"))
					log4plus.Info("%s command.ID=[%s] imageBGUrl=[%s]", funName, command.ID, imageBGUrl)
				}
			} else if v.Type == discordgo.ApplicationCommandOptionString {
				if v.Name == "mode" {
					mode = v.StringValue()
					log4plus.Info("%s mode=[%s]", funName, mode)
				}
			}

		}
	}
	log4plus.Info("%s image=[%s] imageBGUrl=[%s] mode=[%s]", funName, imageUrl, imageBGUrl, mode)
	if strings.Trim(imageUrl, " ") == "" {
		errString := fmt.Sprintf("%s Failed imageUrl is Empty", funName)
		log4plus.Info("%s errString=[%s]", funName, errString)
		return errors.New(errString), []byte(""), header.RequestReplaceBG{}
	}
	if strings.Trim(imageBGUrl, " ") == "" {
		errString := fmt.Sprintf("%s Failed imageBGUrl is Empty", funName)
		log4plus.Info("%s errString=[%s]", funName, errString)
		return errors.New(errString), []byte(""), header.RequestReplaceBG{}
	}

	err, newImageUrl, localImagePath := a.store.DownloadFile(imageUrl)
	if err != nil {
		errString := fmt.Sprintf("%s DownloadFile imageUrl=[%s]", funName, imageUrl)
		log4plus.Info("%s errString=[%s]", funName, errString)
		return errors.New(errString), []byte(""), header.RequestReplaceBG{}
	}
	err, newImageBGUrl, localImageBGPath := a.store.DownloadFile(imageBGUrl)
	if err != nil {
		errString := fmt.Sprintf("%s DownloadFile imageBGUrl=[%s]", funName, imageBGUrl)
		log4plus.Info("%s errString=[%s]", funName, errString)
		return errors.New(errString), []byte(""), header.RequestReplaceBG{}
	}
	log4plus.Info("%s DownloadFile newImageUrl=[%s] localImagePath=[%s] newImageBGUrl=[%s] localImageBGPath=[%s]", funName, newImageUrl, localImagePath, newImageBGUrl, localImageBGPath)

	request := header.RequestReplaceBG{
		Kind:    "photo",
		Obj:     mode,
		SrcUrl:  newImageUrl,
		BGColor: "0,0,0,0",
		BGUrl:   newImageBGUrl,
	}

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

func (a *RemoveBG) replaceBG(s *discordgo.Session, i *discordgo.InteractionCreate) {
	funName := "replaceBG"
	cmdName := "/replaceBG"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s consumption time=%d(s)", funName, time.Now().Unix()-now)
	}()
	discordID := common.GetUserId(s, i)
	log4plus.Info("%s discordID=[%s]", funName, discordID)

	err, body, removeBGBody := a.parseReplaceBG(s, i)
	if err != nil {
		errString := fmt.Sprintf("%s parseReplaceBG Failed discordID=[%s]", funName, discordID)
		log4plus.Error(errString)
		common.SetCommandErrorResult(fmt.Sprintf("%s\n", cmdName), cmdName, errString, s, i)
		return
	}
	apiKey := "123456"
	a.setFirst(fmt.Sprintf("%s\n", cmdName), s, i)
	err, taskId := implementation.SingtonRemoveBG().ReplaceBG(funName, apiKey, body)
	if err != nil {
		a.setAnswerError(cmdName, err.Error(), s, i)
		return
	}

	var cmdlines []string
	cmdlines = append(cmdlines, fmt.Sprintf("%s", cmdName))
	cmdlines = append(cmdlines, fmt.Sprintf("taskId=%s", taskId))
	a.setFirstComplete(cmdlines, s, i)
	cmdlines = cmdlines[:0]

	//循环获取返回的数据
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
			cmdlines = append(cmdlines, fmt.Sprintf("target image: %s", removeBGBody.SrcUrl))
			cmdlines = append(cmdlines, fmt.Sprintf("replaced background image: %s", removeBGBody.BGUrl))
			cmdlines = append(cmdlines, fmt.Sprintf("obtained image: %s", newUrl))
			cmdlines = append(cmdlines, fmt.Sprintf("For detailed interface explanations, please refer to:%s",
				configure.SingtonConfigure().Interfaces.ReplaceBG.Urls.ReplaceBG.Comment))
			a.setFirstComplete(cmdlines, s, i)
		} else {
			errString := fmt.Sprintf("%s data is Failed data=[%s]", funName, resultData)
			log4plus.Error(errString)
			a.setAnswerError(cmdName, errString, s, i)
			return
		}
	}
}

func (a *RemoveBG) setFuncs() {
	a.store.SetDiscordFunction("removebg", a.removeBG)
	a.store.SetDiscordFunction("replacebg", a.replaceBG)
}

func (a *RemoveBG) initCommand() {
	a.commandLines = append(a.commandLines, &header.CommandLine{
		Command: "removebg",
		En:      "The /removebg command can remove the background from the image you upload, making it a transparent background",
		Zh:      "removebg命令，可以剔除您上传图片的背景，使之成为透明背景",
	})
	a.commandLines = append(a.commandLines, &header.CommandLine{
		Command: "replacebg",
		En:      "The /replacebg command allows you to replace the background of the image you uploaded",
		Zh:      "replacebg命令，可以替换您上传的图片背景",
	})
	a.store.AddCommandLine(a.commandLines)
}

func (a *RemoveBG) setFirst(txt string, s *discordgo.Session, i *discordgo.InteractionCreate) string {
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

func (a *RemoveBG) setFirstComplete(txts []string, s *discordgo.Session, i *discordgo.InteractionCreate) string {
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

func (a *RemoveBG) setAnswerError(cmd string, err string, s *discordgo.Session, i *discordgo.InteractionCreate) string {
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

func SingtonRemoveBG(store header.DiscordPluginStore) *RemoveBG {
	if nil == gRemoveBG {
		gRemoveBG = &RemoveBG{
			store: store,
		}
		gRemoveBG.setCommand()
		gRemoveBG.setFuncs()
		gRemoveBG.initCommand()
	}
	return gRemoveBG
}
