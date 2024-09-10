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

type FaceFusion struct {
	roots        *x509.CertPool
	rootPEM      []byte
	store        header.PluginStore
	commandLines []*header.CommandLine
}

var gFaceFusion *FaceFusion

func (a *FaceFusion) setCommand() {
	backWhite2ColorCmd := &discordgo.ApplicationCommand{
		Name:        "blackwhite2color",
		Description: fmt.Sprintf("%s Text dubbing", configure.SingtonConfigure().Application.Name),
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "source",
				Description: "Please select the source image you want to use",
				Type:        discordgo.ApplicationCommandOptionAttachment,
				Required:    true,
			}, {
				Name:        "strength",
				Description: "using custom strength",
				Type:        discordgo.ApplicationCommandOptionNumber,
				Required:    true,
			},
		},
	}
	a.store.RegisteredCommand(backWhite2ColorCmd)

	lipSyncerCmd := &discordgo.ApplicationCommand{
		Name:        "lipsyncer",
		Description: fmt.Sprintf("%s Text dubbing", configure.SingtonConfigure().Application.Name),
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "audio",
				Description: "Please select the Audio Object you want to use",
				Type:        discordgo.ApplicationCommandOptionAttachment,
				Required:    true,
			}, {
				Name:        "video",
				Description: "Please select the Video Object you want to use",
				Type:        discordgo.ApplicationCommandOptionAttachment,
				Required:    true,
			}, {
				Name:        "isenhance",
				Description: "using custom isenhance",
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Required:    true,
			},
		},
	}
	a.store.RegisteredCommand(lipSyncerCmd)

	faceSwapCmd := &discordgo.ApplicationCommand{
		Name:        "faceswap",
		Description: fmt.Sprintf("%s Text dubbing", configure.SingtonConfigure().Application.Name),
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "reference",
				Description: "Please select the Audio Object you want to use",
				Type:        discordgo.ApplicationCommandOptionAttachment,
				Required:    true,
			}, {
				Name:        "target",
				Description: "Please select the Video Object you want to use",
				Type:        discordgo.ApplicationCommandOptionAttachment,
				Required:    true,
			}, {
				Name:        "strength",
				Description: "using custom strength",
				Type:        discordgo.ApplicationCommandOptionNumber,
				Required:    true,
			}, {
				Name:        "isenhance",
				Description: "using custom isenhance",
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Required:    true,
			}, {
				Name:        "enhancestrength",
				Description: "using custom strength",
				Type:        discordgo.ApplicationCommandOptionNumber,
				Required:    true,
			},
		},
	}
	a.store.RegisteredCommand(faceSwapCmd)

	frameEnhanceCmd := &discordgo.ApplicationCommand{
		Name:        "frameenhance",
		Description: fmt.Sprintf("%s Text dubbing", configure.SingtonConfigure().Application.Name),
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "source",
				Description: "Please select the source Object you want to use",
				Type:        discordgo.ApplicationCommandOptionAttachment,
				Required:    true,
			}, {
				Name:        "strength",
				Description: "using custom strength",
				Type:        discordgo.ApplicationCommandOptionNumber,
				Required:    true,
			}, {
				Name:        "isenhance",
				Description: "using custom isenhance",
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Required:    true,
			}, {
				Name:        "enhancestrength",
				Description: "using custom strength",
				Type:        discordgo.ApplicationCommandOptionNumber,
				Required:    true,
			},
		},
	}
	a.store.RegisteredCommand(frameEnhanceCmd)
}

func (a *FaceFusion) Command(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

func (a *FaceFusion) parseBackWhite2Color(i *discordgo.InteractionCreate) (error, []byte, header.RequestBlackWhite2Color) {
	funName := "parseBackWhite2Color"
	//解析参数
	var sourceUrl string
	var strength float64
	if i.Type == discordgo.InteractionApplicationCommand {
		command := i.ApplicationCommandData()
		for _, v := range command.Options {
			if v.Type == discordgo.ApplicationCommandOptionAttachment {
				for _, value := range command.Resolved.Attachments {
					sourceUrl = value.URL
					log4plus.Info("%s command.ID=[%s] value=[%s]", funName, command.ID, value.URL)
				}
			} else if v.Type == discordgo.ApplicationCommandOptionNumber {
				if v.Name == "strength" {
					strength = v.FloatValue()
					log4plus.Info("%s strength=[%.2f]", funName, strength)
				}
			}
		}
	}
	if strings.Trim(sourceUrl, " ") == "" {
		errString := fmt.Sprintf("%s Failed sourceUrl is Empty", funName)
		log4plus.Info("%s errString=[%s]", funName, errString)
		return errors.New(errString), []byte(""), header.RequestBlackWhite2Color{}
	}
	err, newUrl, localPath := a.store.DownloadFile(sourceUrl)
	if err != nil {
		errString := fmt.Sprintf("%s DownloadFile sourceUrl=[%s]", funName, sourceUrl)
		log4plus.Info("%s errString=[%s]", funName, errString)
		return errors.New(errString), []byte(""), header.RequestBlackWhite2Color{}
	}
	log4plus.Info("%s DownloadFile newUrl=[%s] localPath=[%s]", funName, newUrl, localPath)

	request := header.RequestBlackWhite2Color{
		Fun:      "blackwhite2color",
		Source:   newUrl,
		Strength: float32(strength),
	}

	tmpData, err := json.Marshal(request)
	if err != nil {
		errString := fmt.Sprintf("%s json.Marshal Failed err=[%s]", funName, err.Error())
		log4plus.Info("%s errString=[%s]", funName, errString)
		return errors.New(errString), []byte(""), header.RequestBlackWhite2Color{}
	}

	var body header.RequestData
	body.BTCAddr = "0000000000000000000000000GFg7xJaNVN2"
	body.Data = json.RawMessage(tmpData)
	data, err := json.Marshal(body)
	if err != nil {
		errString := fmt.Sprintf("%s json.Marshal Failed err=[%s]", funName, err.Error())
		log4plus.Info("%s errString=[%s]", funName, errString)
		return errors.New(errString), []byte(""), header.RequestBlackWhite2Color{}
	}

	return nil, data, request
}

func (a *FaceFusion) blackWhite2Color(s *discordgo.Session, i *discordgo.InteractionCreate) {
	funName := "blackWhite2Color"
	cmdName := "/blackWhite2Color"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s consumption time=%d(s)", funName, time.Now().Unix()-now)
	}()
	//获取用户id
	discordID := common.GetUserId(s, i)
	log4plus.Info("%s discordID=[%s]", funName, discordID)

	err, body, backWhite2ColorBody := a.parseBackWhite2Color(i)
	if err != nil {
		errString := fmt.Sprintf("%s parseBackWhite2Color Failed discordID=[%s]", funName, discordID)
		log4plus.Error(errString)
		common.SetCommandErrorResult(fmt.Sprintf("%s\n", cmdName), cmdName, errString, s, i)
		return
	}
	apiKey := "123456"
	//先进行返回
	a.setFirst(fmt.Sprintf("%s\n", cmdName), s, i)
	err, taskId := implementation.SingtonFaceFusion().BlackWhite2Color(funName, apiKey, body)
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
			Data interface{} `json:"data"`
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
		imageUrl := result.Data.Data.(string)

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
		cmdlines = append(cmdlines, fmt.Sprintf("source:%s", backWhite2ColorBody.Source))
		cmdlines = append(cmdlines, fmt.Sprintf("dest:%s", newUrl))
		cmdlines = append(cmdlines, fmt.Sprintf("For detailed interface explanations, please refer to:%s", configure.SingtonConfigure().Interfaces.FaceFusion.Urls.BlackWhite2Color.Comment))
		a.setFirstComplete(cmdlines, s, i)
	}
}

func (a *FaceFusion) parseLipSyncer(i *discordgo.InteractionCreate) (error, []byte, header.RequestLipSyncer) {
	funName := "parseLipSyncer"
	//解析参数
	var audioUrl string = ""
	var videoUrl string = ""
	var audioId string
	var videoId string
	var isEnhance bool
	if i.Type == discordgo.InteractionApplicationCommand {
		command := i.ApplicationCommandData()
		for _, v := range command.Options {
			if v.Type == discordgo.ApplicationCommandOptionAttachment {
				if v.Name == "audio" {
					audioId = v.Value.(string)
					log4plus.Info("%s audioId=[%s] audioUrl=[%s]", funName, audioId, command.Resolved.Attachments[audioId].URL)
				} else if v.Name == "video" {
					videoId = v.Value.(string)
					log4plus.Info("%s videoId=[%s] videoUrl=[%s]", funName, videoId, command.Resolved.Attachments[videoId].URL)
				}
			} else if v.Type == discordgo.ApplicationCommandOptionBoolean {
				isEnhance = v.Value.(bool)
			}
		}
		audioUrl = command.Resolved.Attachments[audioId].URL
		videoUrl = command.Resolved.Attachments[videoId].URL
		if !strings.Contains(strings.ToLower(audioUrl), ".wav") &&
			!strings.Contains(strings.ToLower(audioUrl), ".wma") &&
			!strings.Contains(strings.ToLower(audioUrl), ".mp3") &&
			!strings.Contains(strings.ToLower(audioUrl), ".m4a") &&
			!strings.Contains(strings.ToLower(audioUrl), ".amr") {
			log4plus.Info("%s convert before audioUrl=[%s] videoUrl=[%s]", funName, audioUrl, videoUrl)
			tmpUrl := audioUrl
			audioUrl = videoUrl
			videoUrl = tmpUrl
			log4plus.Info("%s convert after audioUrl=[%s] videoUrl=[%s]", funName, audioUrl, videoUrl)
		}
		log4plus.Info("%s audioUrl：%s videoUrl：%s", funName, audioUrl, videoUrl)
	}
	if strings.Trim(audioUrl, " ") == "" {
		errString := fmt.Sprintf("%s Failed audioUrl is Empty", funName)
		log4plus.Error(errString)
		return errors.New(errString), []byte(""), header.RequestLipSyncer{}
	}
	if strings.Trim(videoUrl, " ") == "" {
		errString := fmt.Sprintf("%s Failed videoUrl is Empty", funName)
		log4plus.Error(errString)
		return errors.New(errString), []byte(""), header.RequestLipSyncer{}
	}

	//这里下载上传的源文件形成新的文件地址
	err, newAudioUrl, localAudioPath := a.store.DownloadFile(audioUrl)
	if err != nil {
		errString := fmt.Sprintf("%s DownloadFile audioUrl=[%s]", funName, audioUrl)
		log4plus.Info("%s errString=[%s]", funName, errString)
		return errors.New(errString), []byte(""), header.RequestLipSyncer{}
	}
	log4plus.Info("%s DownloadFile newAudioUrl=[%s] localAudioPath=[%s]", funName, newAudioUrl, localAudioPath)

	err, newVideoUrl, localVideoPath := a.store.DownloadFile(videoUrl)
	if err != nil {
		errString := fmt.Sprintf("%s DownloadFile videoUrl=[%s]", funName, videoUrl)
		log4plus.Info("%s errString=[%s]", funName, errString)
		return errors.New(errString), []byte(""), header.RequestLipSyncer{}
	}
	log4plus.Info("%s DownloadFile newVideoUrl=[%s] localVideoPath=[%s]", funName, newVideoUrl, localVideoPath)

	request := header.RequestLipSyncer{
		Fun:       "lip_syncer",
		Audio:     newAudioUrl,
		Video:     newVideoUrl,
		IsEnhance: isEnhance,
	}

	tmpData, err := json.Marshal(request)
	if err != nil {
		errString := fmt.Sprintf("%s json.Marshal Failed err=[%s]", funName, err.Error())
		log4plus.Info("%s errString=[%s]", funName, errString)
		return errors.New(errString), []byte(""), header.RequestLipSyncer{}
	}

	var body header.RequestData
	body.BTCAddr = "0000000000000000000000000GFg7xJaNVN2"
	body.Data = json.RawMessage(tmpData)

	data, err := json.Marshal(body)
	if err != nil {
		errString := fmt.Sprintf("%s json.Marshal Failed err=[%s]", funName, err.Error())
		log4plus.Info("%s errString=[%s]", funName, errString)
		return errors.New(errString), []byte(""), header.RequestLipSyncer{}
	}

	return nil, data, request
}

func (a *FaceFusion) lipSyncer(s *discordgo.Session, i *discordgo.InteractionCreate) {
	funName := "lipSyncer"
	cmdName := "/lipSyncer"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s consumption time=%d(s)", funName, time.Now().Unix()-now)
	}()
	//获取用户id
	discordID := common.GetUserId(s, i)
	log4plus.Info("%s discordID=[%s]", funName, discordID)

	err, body, _ := a.parseLipSyncer(i)
	if err != nil {
		errString := fmt.Sprintf("%s parseLipSyncer Failed discordID=[%s]", funName, discordID)
		log4plus.Error(errString)
		common.SetCommandErrorResult(fmt.Sprintf("%s\n", cmdName), cmdName, errString, s, i)
		return
	}
	apiKey := "123456"
	//先进行返回
	a.setFirst(fmt.Sprintf("%s\n", cmdName), s, i)
	err, taskId := implementation.SingtonFaceFusion().LipSyncer(funName, apiKey, body)
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
			Data interface{} `json:"data"`
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
		imageUrl := result.Data.Data.(string)

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
		cmdlines = append(cmdlines, fmt.Sprintf("%s", newUrl))
		cmdlines = append(cmdlines, fmt.Sprintf("For detailed interface explanations, please refer to:%s", configure.SingtonConfigure().Interfaces.FaceFusion.Urls.LipSyncer.Comment))
		a.setFirstComplete(cmdlines, s, i)
	}
}

func (a *FaceFusion) parseFaceSwap(i *discordgo.InteractionCreate) (error, []byte, header.RequestFaceSwap) {
	funName := "parseFaceSwap"
	//解析参数
	var referenceUrl string = ""
	var targetUrl string = ""
	var referenceId string
	var targetId string
	var strength float64
	var isFaceEnhance bool
	var faceEnhanceStrength float64
	if i.Type == discordgo.InteractionApplicationCommand {
		command := i.ApplicationCommandData()
		for _, v := range command.Options {
			if v.Type == discordgo.ApplicationCommandOptionAttachment {
				if v.Name == "reference" {

					referenceId = v.Value.(string)
					log4plus.Info("%s referenceId=[%s] referenceUrl=[%s]", funName, referenceId, command.Resolved.Attachments[referenceId].URL)
				} else if v.Name == "target" {

					targetId = v.Value.(string)
					log4plus.Info("%s targetId=[%s] targetUrl=[%s]", funName, targetId, command.Resolved.Attachments[targetId].URL)
				}
			} else if v.Type == discordgo.ApplicationCommandOptionBoolean {
				isFaceEnhance = v.Value.(bool)
			} else if v.Type == discordgo.ApplicationCommandOptionNumber {
				if v.Name == "strength" {

					strength = v.FloatValue()
					log4plus.Info("%s strength=[%.2f]", funName, strength)
				} else if v.Name == "enhancestrength" {

					faceEnhanceStrength = v.FloatValue()
					log4plus.Info("%s faceEnhanceStrength=[%.2f]", funName, faceEnhanceStrength)
				}
			}
		}
		referenceUrl = command.Resolved.Attachments[referenceId].URL
		targetUrl = command.Resolved.Attachments[targetId].URL
		log4plus.Info("%s referenceUrl=[%s] targetUrl=[%s]", funName, referenceUrl, targetUrl)
	}
	if strings.Trim(referenceUrl, " ") == "" {
		log4plus.Error("%s Failed reference is Empty", funName)
		return errors.New(""), []byte(""), header.RequestFaceSwap{}
	}
	if strings.Trim(targetUrl, " ") == "" {
		log4plus.Error("%s Failed target is Empty", funName)
		return errors.New(""), []byte(""), header.RequestFaceSwap{}
	}

	err, newReferenceUrl, localReferencePath := a.store.DownloadFile(referenceUrl)
	if err != nil {
		errString := fmt.Sprintf("%s DownloadFile referenceUrl=[%s]", funName, referenceUrl)
		log4plus.Info("%s errString=[%s]", funName, errString)
		return errors.New(errString), []byte(""), header.RequestFaceSwap{}
	}
	log4plus.Info("%s DownloadFile newReferenceUrl=[%s] localReferencePath=[%s]", funName, newReferenceUrl, localReferencePath)

	err, newTargetUrl, localTargetPath := a.store.DownloadFile(targetUrl)
	if err != nil {
		errString := fmt.Sprintf("%s DownloadFile targetUrl=[%s]", funName, targetUrl)
		log4plus.Info("%s errString=[%s]", funName, errString)
		return errors.New(errString), []byte(""), header.RequestFaceSwap{}
	}
	log4plus.Info("%s DownloadFile newTargetUrl=[%s] localTargetPath=[%s]", funName, newTargetUrl, localTargetPath)

	request := header.RequestFaceSwap{
		Fun:                 "face_swap",
		FaceReference:       newReferenceUrl,
		FaceTarget:          newTargetUrl,
		Strength:            float32(strength),
		IsFaceEnhance:       isFaceEnhance,
		FaceEnhanceStrength: float32(faceEnhanceStrength),
	}
	log4plus.Info("%s request=[%+v]", funName, request)

	tmpData, err := json.Marshal(request)
	if err != nil {
		errString := fmt.Sprintf("%s json.Marshal Failed err=[%s]", funName, err.Error())
		log4plus.Info("%s errString=[%s]", funName, errString)
		return errors.New(errString), []byte(""), header.RequestFaceSwap{}
	}

	var body header.RequestData
	body.BTCAddr = "0000000000000000000000000GFg7xJaNVN2"
	body.Data = json.RawMessage(tmpData)

	data, err := json.Marshal(body)
	if err != nil {
		errString := fmt.Sprintf("%s json.Marshal Failed err=[%s]", funName, err.Error())
		log4plus.Info("%s errString=[%s]", funName, errString)
		return errors.New(errString), []byte(""), header.RequestFaceSwap{}
	}

	return nil, data, request
}

func (a *FaceFusion) faceSwap(s *discordgo.Session, i *discordgo.InteractionCreate) {
	funName := "faceSwap"
	cmdName := "/faceSwap"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s consumption time=%d(s)", funName, time.Now().Unix()-now)
	}()
	//获取用户id
	discordID := common.GetUserId(s, i)
	log4plus.Info("%s discordID=[%s]", funName, discordID)

	err, body, _ := a.parseFaceSwap(i)
	if err != nil {
		errString := fmt.Sprintf("%s parseFrameEnhance Failed discordID=[%s]", funName, discordID)
		log4plus.Error(errString)
		common.SetCommandErrorResult(fmt.Sprintf("%s\n", cmdName), cmdName, errString, s, i)
		return
	}
	apiKey := "123456"
	a.setFirst(fmt.Sprintf("%s\n", cmdName), s, i)
	err, taskId := implementation.SingtonFaceFusion().FaceSwap(funName, apiKey, body)
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
			Data interface{} `json:"data"`
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
		imageUrl := result.Data.Data.(string)

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
		cmdlines = append(cmdlines, fmt.Sprintf("%s", newUrl))
		cmdlines = append(cmdlines, fmt.Sprintf("For detailed interface explanations, please refer to:%s", configure.SingtonConfigure().Interfaces.FaceFusion.Urls.FaceSwap.Comment))
		a.setFirstComplete(cmdlines, s, i)

	}
}

func (a *FaceFusion) parseFrameEnhance(i *discordgo.InteractionCreate) (error, []byte, header.RequestFrameEnhance) {
	funName := "parseFrameEnhance"
	//解析参数
	var sourceUrl string = ""
	var sourceId string
	var strength float64
	var isFaceEnhance bool
	var faceEnhanceStrength float64
	if i.Type == discordgo.InteractionApplicationCommand {
		command := i.ApplicationCommandData()
		for _, v := range command.Options {
			if v.Type == discordgo.ApplicationCommandOptionAttachment {
				if v.Name == "source" {

					sourceId = v.Value.(string)
					log4plus.Info("%s referenceId=[%s] sourceUrl=[%s]", funName, sourceId, command.Resolved.Attachments[sourceId].URL)
				}
			} else if v.Type == discordgo.ApplicationCommandOptionBoolean {
				isFaceEnhance = v.Value.(bool)
			} else if v.Type == discordgo.ApplicationCommandOptionNumber {
				if v.Name == "strength" {

					strength = v.FloatValue()
					log4plus.Info("%s strength=[%.2f]", funName, strength)
				} else if v.Name == "enhancestrength" {

					faceEnhanceStrength = v.FloatValue()
					log4plus.Info("%s faceEnhanceStrength=[%.2f]", funName, faceEnhanceStrength)
				}
			}
		}
		sourceUrl = command.Resolved.Attachments[sourceId].URL
		log4plus.Info("%s sourceUrl=[%s]", funName, sourceUrl)
	}
	if strings.Trim(sourceUrl, " ") == "" {
		log4plus.Error("%s Failed reference is Empty", funName)
		return errors.New(""), []byte(""), header.RequestFrameEnhance{}
	}

	err, newSourceUrl, localSourcePath := a.store.DownloadFile(sourceUrl)
	if err != nil {
		errString := fmt.Sprintf("%s DownloadFile sourceUrl=[%s]", funName, sourceUrl)
		log4plus.Info("%s errString=[%s]", funName, errString)
		return errors.New(errString), []byte(""), header.RequestFrameEnhance{}
	}
	log4plus.Info("%s DownloadFile newSourceUrl=[%s] localSourcePath=[%s]", funName, newSourceUrl, localSourcePath)

	request := header.RequestFrameEnhance{
		Fun:                 "frameEnhance",
		Source:              newSourceUrl,
		Strength:            float32(strength),
		IsFaceEnhance:       isFaceEnhance,
		FaceEnhanceStrength: float32(faceEnhanceStrength),
	}
	log4plus.Info("%s request=[%+v]", funName, request)

	tmpData, err := json.Marshal(request)
	if err != nil {
		errString := fmt.Sprintf("%s json.Marshal Failed err=[%s]", funName, err.Error())
		log4plus.Info("%s errString=[%s]", funName, errString)
		return errors.New(errString), []byte(""), header.RequestFrameEnhance{}
	}

	var body header.RequestData
	body.BTCAddr = "0000000000000000000000000GFg7xJaNVN2"
	body.Data = json.RawMessage(tmpData)

	data, err := json.Marshal(body)
	if err != nil {
		errString := fmt.Sprintf("%s json.Marshal Failed err=[%s]", funName, err.Error())
		log4plus.Info("%s errString=[%s]", funName, errString)
		return errors.New(errString), []byte(""), header.RequestFrameEnhance{}
	}

	return nil, data, request
}

func (a *FaceFusion) frameEnhance(s *discordgo.Session, i *discordgo.InteractionCreate) {
	funName := "frameEnhance"
	cmdName := "/frameEnhance"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s consumption time=%d(s)", funName, time.Now().Unix()-now)
	}()
	//获取用户id
	discordID := common.GetUserId(s, i)
	log4plus.Info("%s discordID=[%s]", funName, discordID)

	err, body, _ := a.parseFrameEnhance(i)
	if err != nil {
		errString := fmt.Sprintf("%s parseFrameEnhance Failed discordID=[%s]", funName, discordID)
		log4plus.Error(errString)
		common.SetCommandErrorResult(fmt.Sprintf("%s\n", cmdName), cmdName, errString, s, i)
		return
	}
	apiKey := "123456"
	//先进行返回
	a.setFirst(fmt.Sprintf("%s\n", cmdName), s, i)
	err, taskId := implementation.SingtonFaceFusion().FrameEnhance(funName, apiKey, body)
	if err != nil {
		a.setAnswerError(funName, err.Error(), s, i)
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
			Data interface{} `json:"data"`
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
		imageUrl := result.Data.Data.(string)

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
		cmdlines = append(cmdlines, fmt.Sprintf("%s", newUrl))
		cmdlines = append(cmdlines, fmt.Sprintf("For detailed interface explanations, please refer to:%s", configure.SingtonConfigure().Interfaces.FaceFusion.Urls.FrameEnhance.Comment))
		a.setFirstComplete(cmdlines, s, i)
	}
}

func (a *FaceFusion) setFuncs() {
	a.store.SetDiscordFunction("blackwhite2color", a.blackWhite2Color)
	a.store.SetDiscordFunction("lipsyncer", a.lipSyncer)
	a.store.SetDiscordFunction("faceswap", a.faceSwap)
	a.store.SetDiscordFunction("frameenhance", a.frameEnhance)
}

func (a *FaceFusion) initCommand() {
	a.commandLines = append(a.commandLines, &header.CommandLine{
		Command: "backWhite2Color",
		En:      "The /backWhite2Color command allows you to colorize black-and-white photos.",
		Zh:      "backWhite2Color命令，允许您为黑白照片进行上色",
	})
	a.commandLines = append(a.commandLines, &header.CommandLine{
		Command: "lipSyncer",
		En:      "The /lipSyncer command allows you to replace a character's voice in a video using a specified audio.",
		Zh:      "lipSyncer命令，允许您使用指定的声音对视频中的人物进行声音替换",
	})
	a.commandLines = append(a.commandLines, &header.CommandLine{
		Command: "faceSwap",
		En:      "The /faceSwap command allows you to swap faces in a video using a photo of your choice.",
		Zh:      "faceSwap命令，允许您使用自己选择的照片对视频中的任务进行换脸",
	})
	a.commandLines = append(a.commandLines, &header.CommandLine{
		Command: "frameEnhance",
		En:      "The /frameEnhance command can enhance the image you provide",
		Zh:      "frameEnhance命令，能否将您给的图片进行增强",
	})
	a.store.AddCommandLine(a.commandLines)
}

func (a *FaceFusion) setFirst(txt string, s *discordgo.Session, i *discordgo.InteractionCreate) string {
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

func (a *FaceFusion) setFirstComplete(txts []string, s *discordgo.Session, i *discordgo.InteractionCreate) string {
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

func (a *FaceFusion) setAnswerError(cmd string, err string, s *discordgo.Session, i *discordgo.InteractionCreate) string {
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

func SingtonFaceFusion(store header.PluginStore) *FaceFusion {
	if nil == gFaceFusion {
		gFaceFusion = &FaceFusion{
			store: store,
		}
		gFaceFusion.setCommand()
		gFaceFusion.setFuncs()
		gFaceFusion.initCommand()
	}
	return gFaceFusion
}
