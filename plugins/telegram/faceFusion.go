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
	blackwhite2colorPrompt = [...]string{
		"Please upload the black-and-white image you want to convert: ",
	}

	lipsyncerPrompt = [...]string{
		"Please upload the black-and-white image you want to convert: ",
	}

	faceswapPrompt = [...]string{
		"Please upload the black-and-white image you want to convert: ",
	}

	frameenhancePrompt = [...]string{
		"Please upload the black-and-white image you want to convert: ",
	}
)

type FaceFusion struct {
	roots        *x509.CertPool
	rootPEM      []byte
	store        header.TelegramPluginStore
	commandLines []*header.CommandLine
}

var gFaceFusion *FaceFusion

func (a *FaceFusion) parseBlackwhite2color(c tele.Context, user *header.UserStep) (error, []byte, header.RequestBlackWhite2Color) {
	funName := "parseBlackwhite2color"
	if user == nil {
		user := &header.UserStep{
			TelegramId:   c.Sender().ID,
			MaxStep:      0,
			FunctionName: "blackwhite2color",
			MessageId:    c.Message().ID,
		}
		a.store.AddUser(user)
		a.store.GetBotObject().Send(c.Chat(), blackwhite2colorPrompt[user.MaxStep], &tele.SendOptions{
			ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
		})
		return nil, []byte(""), header.RequestBlackWhite2Color{}
	}
	if user.MaxStep == 0 {
		photo := c.Message().Photo
		if photo == nil {
			errString := fmt.Sprintf("%s c.Message().Photo is null", funName)
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestBlackWhite2Color{}
		}
		err, imageUrl := a.store.SaveMediaFile(photo.FileID)
		if err != nil {
			errString := fmt.Sprintf("%s a.store.SaveMediaFile Failed fileId=[%s]", funName, photo.FileID)
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestBlackWhite2Color{}
		}
		log4plus.Info("%s imageUrl=[%s]", funName, imageUrl)

		request := header.RequestBlackWhite2Color{
			Fun:      "blackwhite2color",
			Source:   imageUrl,
			Strength: float32(1),
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
			return errors.New(errString), []byte(""), header.RequestBlackWhite2Color{}
		}
		user.Steps = append(user.Steps, step)

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
	errString := fmt.Sprintf("%s current step Failed maxStep=[%d]", funName, user.MaxStep)
	log4plus.Info("%s errString=[%s]", funName, errString)
	return errors.New(errString), []byte(""), header.RequestBlackWhite2Color{}
}

func (a *FaceFusion) blackwhite2color(c tele.Context) error {
	funName := "blackwhite2color"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s consumption time=%d(s)", funName, time.Now().Unix()-now)
	}()
	log4plus.Info("%s telegramID=[%d]", funName, c.Sender().ID)

	user := a.store.FindUser(c.Sender().ID)
	if user == nil {
		log4plus.Info("%s first FindUser=[%d]", funName, c.Sender().ID)
		a.parseBlackwhite2color(c, nil)
		return nil
	}
	if user.MaxStep == 0 {
		err, body, _ := a.parseBlackwhite2color(c, user)
		if err != nil {
			errString := fmt.Sprintf("%s parseBlackwhite2color Failed telegramID=[%d]", funName, c.Sender().ID)
			log4plus.Error(errString)

			a.store.GetBotObject().Send(c.Chat(), errString, &tele.SendOptions{
				ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
			})
			return errors.New(errString)
		}
		apiKey := "123456"
		err, taskId := implementation.SingtonFaceFusion().BlackWhite2Color(funName, apiKey, body)
		if err != nil {
			errString := fmt.Sprintf("%s BlackWhite2Color Failed err=[%s]", funName, err.Error())
			log4plus.Error(errString)

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
			a.store.GetBotObject().Edit(taskMessage, errString, &tele.SendOptions{
				ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
			})
			return errors.New(errString)
		}
		imageUrl := result.Data.Data.(string)

		err, newUrl, localPath := a.store.DownloadFile(imageUrl)
		if err != nil {
			errString := fmt.Sprintf("%s DownloadFile newUrl=[%s]", funName, newUrl)
			log4plus.Error(errString)
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
		a.store.DeleteUser(c.Sender().ID)
		return nil
	}
	errString := fmt.Sprintf("%s current step mexStep=[%d]", funName, user.MaxStep)
	log4plus.Error(errString)
	a.store.GetBotObject().Send(c.Chat(), errString, &tele.SendOptions{
		ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
	})
	return errors.New(errString)
}

/********************************/

func (a *FaceFusion) parseLipsyncer(c tele.Context, user *header.UserStep) (error, []byte, header.RequestLipSyncer) {
	funName := "parseLipsyncer"

	if user == nil {
		user := &header.UserStep{
			TelegramId:   c.Sender().ID,
			MaxStep:      0,
			FunctionName: "lipsyncer",
			MessageId:    c.Message().ID,
		}
		a.store.AddUser(user)
		a.store.GetBotObject().Send(c.Chat(), lipsyncerPrompt[user.MaxStep], &tele.SendOptions{
			ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
		})
		return nil, []byte(""), header.RequestLipSyncer{}
	}
	if user.MaxStep == 0 {
		video := c.Message().Video
		if video == nil {
			errString := fmt.Sprintf("%s c.Message().Video is null", funName)
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestLipSyncer{}
		}
		err, videoUrl := a.store.SaveMediaFile(video.FileID)
		if err != nil {
			errString := fmt.Sprintf("%s a.store.SaveMediaFile Failed fileId=[%s]", funName, video.FileID)
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestLipSyncer{}
		}
		log4plus.Info("%s videoUrl=[%s]", funName, videoUrl)

		request := header.RequestLipSyncer{
			Fun:       "lip_syncer",
			Audio:     "",
			Video:     videoUrl,
			IsEnhance: true,
		}
		user.MaxStep = 1
		step := header.Step{
			StepId:      1,
			RequestTime: time.Now(),
		}
		step.Data, err = json.Marshal(request)
		if err != nil {
			errString := fmt.Sprintf("%s DownloadFile videoUrl=[%s]", funName, videoUrl)
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestLipSyncer{}
		}
		user.Steps = append(user.Steps, step)

		a.store.GetBotObject().Send(c.Chat(), lipsyncerPrompt[user.MaxStep], &tele.SendOptions{
			ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
		})
		return nil, []byte(""), header.RequestLipSyncer{}

	} else if user.MaxStep == 1 {

		audio := c.Message().Audio
		if audio == nil {
			errString := fmt.Sprintf("%s c.Message().Audio is null", funName)
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestLipSyncer{}
		}
		err, audioUrl := a.store.SaveMediaFile(audio.FileID)
		if err != nil {
			errString := fmt.Sprintf("%s a.store.SaveMediaFile Failed fileId=[%s]", funName, audio.FileID)
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestLipSyncer{}
		}
		log4plus.Info("%s audioUrl=[%s]", funName, audioUrl)

		var request header.RequestLipSyncer
		if err := json.Unmarshal(user.Steps[user.MaxStep].Data, &request); err != nil {
			errString := fmt.Sprintf("%s Unmarshal err=[%s]", funName, err.Error())
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestLipSyncer{}
		}
		request.Audio = audioUrl
		user.MaxStep++
		step := header.Step{
			StepId:      user.MaxStep,
			RequestTime: time.Now(),
		}
		step.Data, err = json.Marshal(request)
		if err != nil {
			errString := fmt.Sprintf("%s Marshal request.Audio=[%s]", funName, request.Audio)
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestLipSyncer{}
		}
		user.Steps = append(user.Steps, step)

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
	errString := fmt.Sprintf("%s current step Failed maxStep=[%d]", funName, user.MaxStep)
	log4plus.Info("%s errString=[%s]", funName, errString)
	return errors.New(errString), []byte(""), header.RequestLipSyncer{}
}

func (a *FaceFusion) lipsyncer(c tele.Context) error {
	funName := "lipsyncer"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s consumption time=%d(s)", funName, time.Now().Unix()-now)
	}()
	log4plus.Info("%s telegramID=[%d]", funName, c.Sender().ID)

	user := a.store.FindUser(c.Sender().ID)
	if user == nil {
		log4plus.Info("%s FindUser=[%d]", funName, c.Sender().ID)
		a.parseLipsyncer(c, nil)
		return nil
	}
	if user.MaxStep == 0 {
		err, body, _ := a.parseLipsyncer(c, user)
		if err != nil {
			errString := fmt.Sprintf("%s parseLipsyncer Failed telegramID=[%d]", funName, c.Sender().ID)
			log4plus.Error(errString)

			a.store.GetBotObject().Send(c.Chat(), errString, &tele.SendOptions{
				ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
			})
			return errors.New(errString)
		}
		apiKey := "123456"
		err, taskId := implementation.SingtonFaceFusion().LipSyncer(funName, apiKey, body)
		if err != nil {
			errString := fmt.Sprintf("%s LipSyncer Failed err=[%s]", funName, err.Error())
			log4plus.Error(errString)

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
			a.store.GetBotObject().Edit(taskMessage, errString, &tele.SendOptions{
				ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
			})
			return errors.New(errString)
		}
		imageUrl := result.Data.Data.(string)

		err, newUrl, localPath := a.store.DownloadFile(imageUrl)
		if err != nil {
			errString := fmt.Sprintf("%s DownloadFile newUrl=[%s]", funName, newUrl)
			log4plus.Error(errString)
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
		a.store.DeleteUser(c.Sender().ID)
		return nil
	}
	errString := fmt.Sprintf("%s current step mexStep=[%d]", funName, user.MaxStep)
	log4plus.Error(errString)
	a.store.GetBotObject().Send(c.Chat(), errString, &tele.SendOptions{
		ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
	})
	return errors.New(errString)
}

/********************************/

func (a *FaceFusion) parseFaceswap(c tele.Context, user *header.UserStep) (error, []byte, header.RequestFaceSwap) {
	funName := "parseFaceswap"

	if user == nil {
		user := &header.UserStep{
			TelegramId:   c.Sender().ID,
			MaxStep:      0,
			FunctionName: "faceswap",
			MessageId:    c.Message().ID,
		}
		a.store.AddUser(user)
		a.store.GetBotObject().Send(c.Chat(), faceswapPrompt[user.MaxStep], &tele.SendOptions{
			ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
		})
		return nil, []byte(""), header.RequestFaceSwap{}
	}
	if user.MaxStep == 0 {
		photo := c.Message().Photo
		if photo == nil {
			errString := fmt.Sprintf("%s c.Message().Photo is null", funName)
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestFaceSwap{}
		}
		err, newReferenceUrl := a.store.SaveMediaFile(photo.FileID)
		if err != nil {
			errString := fmt.Sprintf("%s a.store.SaveMediaFile Failed fileId=[%s]", funName, photo.FileID)
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestFaceSwap{}
		}
		log4plus.Info("%s newReferenceUrl=[%s]", funName, newReferenceUrl)

		request := header.RequestFaceSwap{
			Fun:                 "face_swap",
			FaceReference:       newReferenceUrl,
			FaceTarget:          "",
			Strength:            1,
			IsFaceEnhance:       true,
			FaceEnhanceStrength: 1,
		}
		user.MaxStep = 1
		step := header.Step{
			StepId:      1,
			RequestTime: time.Now(),
		}
		step.Data, err = json.Marshal(request)
		if err != nil {
			errString := fmt.Sprintf("%s DownloadFile newReferenceUrl=[%s]", funName, newReferenceUrl)
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestFaceSwap{}
		}
		user.Steps = append(user.Steps, step)

		a.store.GetBotObject().Send(c.Chat(), faceswapPrompt[user.MaxStep], &tele.SendOptions{
			ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
		})
		return nil, []byte(""), header.RequestFaceSwap{}

	} else if user.MaxStep == 1 {

		audio := c.Message().Audio
		if audio == nil {
			errString := fmt.Sprintf("%s c.Message().Audio is null", funName)
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestFaceSwap{}
		}
		err, newTargetUrl := a.store.SaveMediaFile(audio.FileID)
		if err != nil {
			errString := fmt.Sprintf("%s a.store.SaveMediaFile Failed fileId=[%s]", funName, audio.FileID)
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestFaceSwap{}
		}
		log4plus.Info("%s newTargetUrl=[%s]", funName, newTargetUrl)

		var request header.RequestFaceSwap
		if err := json.Unmarshal(user.Steps[user.MaxStep].Data, &request); err != nil {
			errString := fmt.Sprintf("%s Unmarshal err=[%s]", funName, err.Error())
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestFaceSwap{}
		}
		request.FaceTarget = newTargetUrl
		user.MaxStep++
		step := header.Step{
			StepId:      user.MaxStep,
			RequestTime: time.Now(),
		}
		step.Data, err = json.Marshal(request)
		if err != nil {
			errString := fmt.Sprintf("%s Marshal request.FaceTarget=[%s]", funName, request.FaceTarget)
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestFaceSwap{}
		}
		user.Steps = append(user.Steps, step)

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
	errString := fmt.Sprintf("%s current step Failed maxStep=[%d]", funName, user.MaxStep)
	log4plus.Info("%s errString=[%s]", funName, errString)
	return errors.New(errString), []byte(""), header.RequestFaceSwap{}
}

func (a *FaceFusion) faceswap(c tele.Context) error {
	funName := "faceswap"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s consumption time=%d(s)", funName, time.Now().Unix()-now)
	}()
	log4plus.Info("%s telegramID=[%d]", funName, c.Sender().ID)

	user := a.store.FindUser(c.Sender().ID)
	if user == nil {
		log4plus.Info("%s FindUser=[%d]", funName, c.Sender().ID)
		a.parseFaceswap(c, nil)
		return nil
	}

	if user.MaxStep == 0 {
		err, body, _ := a.parseFaceswap(c, user)
		if err != nil {
			errString := fmt.Sprintf("%s parseFaceswap Failed telegramID=[%d]", funName, c.Sender().ID)
			log4plus.Error(errString)

			a.store.GetBotObject().Send(c.Chat(), errString, &tele.SendOptions{
				ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
			})
			return errors.New(errString)
		}
		apiKey := "123456"
		err, taskId := implementation.SingtonFaceFusion().FaceSwap(funName, apiKey, body)
		if err != nil {
			errString := fmt.Sprintf("%s FaceSwap Failed err=[%s]", funName, err.Error())
			log4plus.Error(errString)

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
			a.store.GetBotObject().Edit(taskMessage, errString, &tele.SendOptions{
				ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
			})
			return errors.New(errString)
		}
		imageUrl := result.Data.Data.(string)

		err, newUrl, localPath := a.store.DownloadFile(imageUrl)
		if err != nil {
			errString := fmt.Sprintf("%s DownloadFile newUrl=[%s]", funName, newUrl)
			log4plus.Error(errString)
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
		a.store.DeleteUser(c.Sender().ID)
		return nil
	}
	errString := fmt.Sprintf("%s current step mexStep=[%d]", funName, user.MaxStep)
	log4plus.Error(errString)
	a.store.GetBotObject().Send(c.Chat(), errString, &tele.SendOptions{
		ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
	})
	return errors.New(errString)
}

/********************************/

func (a *FaceFusion) parseFrameenhance(c tele.Context, user *header.UserStep) (error, []byte, header.RequestFrameEnhance) {
	funName := "parseFrameenhance"

	if user == nil {
		user := &header.UserStep{
			TelegramId:   c.Sender().ID,
			MaxStep:      0,
			FunctionName: "frameenhance",
			MessageId:    c.Message().ID,
		}
		a.store.AddUser(user)
		a.store.GetBotObject().Send(c.Chat(), frameenhancePrompt[user.MaxStep], &tele.SendOptions{
			ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
		})
		return nil, []byte(""), header.RequestFrameEnhance{}
	}

	if user.MaxStep == 0 {
		photo := c.Message().Photo
		if photo == nil {
			errString := fmt.Sprintf("%s c.Message().Photo is null", funName)
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestFrameEnhance{}
		}
		err, newSourceUrl := a.store.SaveMediaFile(photo.FileID)
		if err != nil {
			errString := fmt.Sprintf("%s a.store.SaveMediaFile Failed fileId=[%s]", funName, photo.FileID)
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestFrameEnhance{}
		}
		log4plus.Info("%s newSourceUrl=[%s]", funName, newSourceUrl)

		request := header.RequestFrameEnhance{
			Fun:                 "frameEnhance",
			Source:              newSourceUrl,
			Strength:            1,
			IsFaceEnhance:       true,
			FaceEnhanceStrength: 1,
		}
		user.MaxStep = 1
		step := header.Step{
			StepId:      1,
			RequestTime: time.Now(),
		}
		step.Data, err = json.Marshal(request)
		if err != nil {
			errString := fmt.Sprintf("%s DownloadFile request.Source=[%s]", funName, request.Source)
			log4plus.Info("%s errString=[%s]", funName, errString)
			return errors.New(errString), []byte(""), header.RequestFrameEnhance{}
		}
		user.Steps = append(user.Steps, step)

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
	errString := fmt.Sprintf("%s current step Failed maxStep=[%d]", funName, user.MaxStep)
	log4plus.Info("%s errString=[%s]", funName, errString)
	return errors.New(errString), []byte(""), header.RequestFrameEnhance{}
}

func (a *FaceFusion) frameenhance(c tele.Context) error {
	funName := "frameenhance"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s consumption time=%d(s)", funName, time.Now().Unix()-now)
	}()
	log4plus.Info("%s telegramID=[%d]", funName, c.Sender().ID)

	user := a.store.FindUser(c.Sender().ID)
	if user == nil {
		log4plus.Info("%s FindUser=[%d]", funName, c.Sender().ID)
		a.parseFrameenhance(c, nil)
		return nil
	}

	if user.MaxStep == 0 {
		err, body, _ := a.parseFrameenhance(c, user)
		if err != nil {
			errString := fmt.Sprintf("%s parseFrameenhance Failed telegramID=[%d]", funName, c.Sender().ID)
			log4plus.Error(errString)

			a.store.GetBotObject().Send(c.Chat(), errString, &tele.SendOptions{
				ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
			})
			return errors.New(errString)
		}
		apiKey := "123456"
		err, taskId := implementation.SingtonFaceFusion().FrameEnhance(funName, apiKey, body)
		if err != nil {
			errString := fmt.Sprintf("%s FrameEnhance Failed err=[%s]", funName, err.Error())
			log4plus.Error(errString)

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
			a.store.GetBotObject().Edit(taskMessage, errString, &tele.SendOptions{
				ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
			})
			return errors.New(errString)
		}
		imageUrl := result.Data.Data.(string)

		err, newUrl, localPath := a.store.DownloadFile(imageUrl)
		if err != nil {
			errString := fmt.Sprintf("%s DownloadFile newUrl=[%s]", funName, newUrl)
			log4plus.Error(errString)
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
		a.store.DeleteUser(c.Sender().ID)
		return nil
	}
	errString := fmt.Sprintf("%s current step mexStep=[%d]", funName, user.MaxStep)
	log4plus.Error(errString)
	a.store.GetBotObject().Send(c.Chat(), errString, &tele.SendOptions{
		ReplyTo: &tele.Message{ID: user.MessageId, Chat: c.Chat()},
	})
	return errors.New(errString)
}

func (a *FaceFusion) setFuncs() {
	a.store.SetTelegramFunction("blackwhite2color", "blackwhite2color", "🎨 blackwhite2color", a.blackwhite2color)
	a.store.SetTelegramFunction("lipsyncer", "lipsyncer", "📷 lipSyncer", a.lipsyncer)

	a.store.SetTelegramFunction("faceswap", "faceswap", "🔃 faceswap", a.faceswap)
	a.store.SetTelegramFunction("frameenhance", "frameenhance", "🚀 frameenhance", a.frameenhance)
}

func SingtonFaceFusion(store header.TelegramPluginStore) *FaceFusion {
	if nil == gFaceFusion {
		gFaceFusion = &FaceFusion{
			store: store,
		}
		gFaceFusion.setFuncs()
	}
	return gFaceFusion
}
