package discord

import (
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/nGPU/bot/aiModule"
	"github.com/nGPU/bot/common"
	"github.com/nGPU/bot/configure"
	"github.com/nGPU/bot/header"
	log4plus "github.com/nGPU/common/log4go"
)

type Fengshui struct {
	roots        *x509.CertPool
	rootPEM      []byte
	store        header.DiscordPluginStore
	commandLines []*header.CommandLine
}

var gFengshui *Fengshui

func (a *Fengshui) setCommand() {
	createtweetCmd := &discordgo.ApplicationCommand{
		Name:        "fengshui",
		Description: fmt.Sprintf("%s Text dubbing", configure.SingtonConfigure().Application.Name),
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "image",
				Description: "Please select the image you want to use",
				Type:        discordgo.ApplicationCommandOptionAttachment,
				Required:    true,
			}, {
				Name:        "prompt",
				Description: "using custom prompt",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    false,
			},
		},
	}
	a.store.RegisteredCommand(createtweetCmd)
}

func (a *Fengshui) Command(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

func (a *Fengshui) getUrlExtension(urlString string) string {
	funName := "getUrlExtension"
	parsedURL, err := url.Parse(urlString)
	if err != nil {
		log4plus.Error("%s Failed err=%s", funName, err.Error())
		return ""
	}
	fileName := path.Base(parsedURL.Path)
	fileExtension := strings.TrimPrefix(path.Ext(fileName), ".")
	log4plus.Info("%s fileExtension=%s", funName, fileExtension)
	return fileExtension
}

func (a *Fengshui) downloadFile(url string) (err error, localUrl, localPath, imageType string) {
	funName := "downloadFile"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s url=%s consumption time=%d(s)", funName, url, time.Now().Unix()-now)
	}()

	response, err := http.Get(url)
	if err != nil {
		log4plus.Error("%s Get Failed err=%s", funName, err.Error())
		return err, "", "", ""
	}
	defer response.Body.Close()

	imageType = response.Header.Get("Content-Type")
	fileExt := a.getUrlExtension(url)
	fileName := fmt.Sprintf("%s%s.%s", time.Now().Format("20060102150405"), fmt.Sprintf("%06d", time.Now().Nanosecond()/1e3), fileExt)
	localPath = fmt.Sprintf("%s/%s", configure.SingtonConfigure().Resource.ResourcePath, fileName)
	file, err := os.Create(localPath)
	if err != nil {
		log4plus.Error("%s Create Failed err=%s", funName, err.Error())
		return err, "", "", ""
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		log4plus.Error("%s Copy Failed err=%s", funName, err.Error())
		return err, "", "", ""
	}
	localUrl = fmt.Sprintf("%s/%s", configure.SingtonConfigure().Resource.Domain, fileName)
	return nil, localUrl, localPath, imageType
}

func (a *Fengshui) parseFengshui(i *discordgo.InteractionCreate) (err error, imageUrl, imagePath, imageType, prompt string) {
	funName := "parseFengshui"
	if i.Type == discordgo.InteractionApplicationCommand {
		command := i.ApplicationCommandData()
		for _, v := range command.Options {
			if v.Type == discordgo.ApplicationCommandOptionAttachment {
				for _, value := range command.Resolved.Attachments {
					imageUrl = value.URL
					log4plus.Info("%s command.ID=[%s] value=[%s]", funName, command.ID, value.URL)
				}
			} else if v.Type == discordgo.ApplicationCommandOptionString {
				if v.Name == "prompt" {
					prompt = v.StringValue()
					log4plus.Info("%s prompt=[%s]", funName, prompt)
				}
			}
		}
	}
	if strings.Trim(imageUrl, " ") == "" {
		errString := fmt.Sprintf("%s Failed imageUrl is Empty", funName)
		log4plus.Info("%s errString=[%s]", funName, errString)
		return errors.New(errString), "", "", "", ""
	}
	err, localUrl, localPath, imageType := a.downloadFile(imageUrl)
	if err != nil {
		errString := fmt.Sprintf("%s downloadFile localPath=[%s] imageType=[%s]", funName, localPath, imageType)
		log4plus.Info("%s errString=[%s]", funName, errString)
		return errors.New(errString), "", "", "", ""
	}
	log4plus.Info("%s downloadFile imageUrl=[%s] localPath=[%s] imageType=[%s]", funName, imageUrl, localPath, imageType)
	imageUrl = localUrl

	return nil, imageUrl, localPath, imageType, prompt
}

func (a *Fengshui) fengshui(s *discordgo.Session, i *discordgo.InteractionCreate) {
	funName := "fengshui"
	cmdName := "/fengshui"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s consumption time=%d(s)", funName, time.Now().Unix()-now)
	}()

	//获取用户id
	discordID := common.GetUserId(s, i)
	log4plus.Info("%s discordID=[%s]", funName, discordID)

	err, imageUrl, imagePath, imageType, question := a.parseFengshui(i)
	if err != nil {
		errString := fmt.Sprintf("%s parseBlip Failed discordID=[%s]", funName, discordID)
		log4plus.Error(errString)
		common.SetCommandErrorResult(fmt.Sprintf("%s\n", cmdName), cmdName, errString, s, i)
		return
	}

	if strings.Trim(question, " ") == "" {
		question = `请首先判断这张图的类型是山水图、房屋图还是人物正脸图，并根据图像类型进行详细分析：

1: 山水图：
风水情况：
吉方位：请说明图中哪些方位主吉。例如，山峰的高低、流水的方向是否有利于风水。
凶方位：请说明图中哪些方位主凶。例如，悬崖、急流、障碍物等可能带来的负面影响。
其它方位：简述其他方位的吉凶情况及其影响。
走势情况：请说明山水图的整体走势情况，如水流的方向、山脉的起伏是否有助于聚集正能量。

2:房屋图：
风水情况：
大吉方位：请说明图中哪些地方具有大吉的风水。例如，主卧室的位置是否得当，财位是否被合理利用。
大凶方位：请说明图中哪些地方具有大凶的风水。例如，卫生间的位置是否影响了其他重要区域，是否有煞气影响。
改善建议：针对大凶方位提供具体的改善建议，包括放置什么风水物件、调整房间布局等方法以趋吉避凶。
趋吉避凶措施：给出如何优化房屋风水的具体措施，例如调整家具摆放、增加风水吉祥物等。

3:人像图：
面相分析：
运势概况：根据面相描述该人物的运势，包括性格特点、健康状态等。
事业运势：分析该人物的事业运势，例如职业发展潜力、适合的职业方向等。
吉凶情况：评估该人物的吉凶情况，包括财运、感情运等方面。
寿命情况：根据面相特征（如面部线条、形状）推测该人物的寿命情况，并解释相关的面相特征。`
	}

	a.setFirst(fmt.Sprintf("%s\n", cmdName), s, i)
	err, txt := aiModule.SingtonAnthropic().FengShui(imagePath, imageType, question)
	if err != nil {
		a.setAnswerError(cmdName, err.Error(), s, i)
		return
	}

	var cmdlines []string
	cmdlines = append(cmdlines, fmt.Sprintf("%s", cmdName))
	cmdlines = append(cmdlines, fmt.Sprintf("source:%s", imageUrl))
	cmdlines = append(cmdlines, fmt.Sprintf("%s", txt))
	cmdlines = append(cmdlines, "For detailed interface explanations, please refer to:https://www.anthropic.com/")
	a.setFirstComplete(cmdlines, s, i)
	cmdlines = cmdlines[:0]
}

func (a *Fengshui) setFuncs() {
	a.store.SetDiscordFunction("fengshui", a.fengshui)
}

func (a *Fengshui) initCommand() {
	a.commandLines = append(a.commandLines, &header.CommandLine{
		Command: "fengshui",
		En:      "The /fengshui command can provide an explanation for the image you upload.",
		Zh:      "fengshui命令，可以根据您上传的图片来判断这张图片中的风水如何，以及解释",
	})
	a.store.AddCommandLine(a.commandLines)
}

func (a *Fengshui) setFirst(txt string, s *discordgo.Session, i *discordgo.InteractionCreate) string {
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

func (a *Fengshui) setFirstComplete(txts []string, s *discordgo.Session, i *discordgo.InteractionCreate) string {
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

func (a *Fengshui) setAnswerError(cmd string, err string, s *discordgo.Session, i *discordgo.InteractionCreate) string {
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

func SingtonFengshui(store header.DiscordPluginStore) *Fengshui {
	if nil == gFengshui {
		gFengshui = &Fengshui{
			store: store,
		}
		gFengshui.setCommand()
		gFengshui.setFuncs()
		gFengshui.initCommand()
	}
	return gFengshui
}
