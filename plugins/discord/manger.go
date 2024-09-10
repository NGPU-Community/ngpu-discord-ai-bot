package discord

import (
	"crypto/x509"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	log4plus "github.com/nGPU/common/log4go"
	"github.com/nGPU/discordBot/common"
	"github.com/nGPU/discordBot/configure"
	"github.com/nGPU/discordBot/db"
	"github.com/nGPU/discordBot/header"
)

type Manager struct {
	roots        *x509.CertPool
	rootPEM      []byte
	store        header.PluginStore
	commandLines []*header.CommandLine
}

var gManager *Manager

func (w *Manager) setCommand() {
	//注册用户
	registerCmd := &discordgo.ApplicationCommand{
		Name:        "register",
		Description: fmt.Sprintf("%s Register User", configure.SingtonConfigure().Application.Name),
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "email",
				Description: "Please enter your email address",
				Required:    true,
			},
		},
	}
	w.store.RegisteredCommand(registerCmd)

	//获取用户基本信息
	getuserCmd := &discordgo.ApplicationCommand{
		Name:        "getuser",
		Description: fmt.Sprintf("%s Get user information", configure.SingtonConfigure().Application.Name),
	}
	w.store.RegisteredCommand(getuserCmd)

	//帮助指令
	makeCmd := &discordgo.ApplicationCommand{
		Name:        "make",
		Description: fmt.Sprintf("%s Bot Shortcuts", configure.SingtonConfigure().Application.Name),
	}
	w.store.RegisteredCommand(makeCmd)

	//帮助指令
	helpCmd := &discordgo.ApplicationCommand{
		Name:        "help",
		Description: fmt.Sprintf("%s Bot Shortcuts", configure.SingtonConfigure().Application.Name),
	}
	w.store.RegisteredCommand(helpCmd)
}

func (w *Manager) setFuncs() {
	w.store.SetDiscordFunction("register", w.register)
	w.store.SetDiscordFunction("getuser", w.getuser)
	w.store.SetDiscordFunction("make", w.helpCommand)
	w.store.SetDiscordFunction("help", w.makeCommand)
}

func (w *Manager) initCommand() {
	w.commandLines = append(w.commandLines, &header.CommandLine{
		Command: "register",
		En:      "With the register command, you can register an account and automatically receive 60 seconds of video creation time ",
		Zh:      "register命令，您可以注册帐户并自动获得60秒的视频创建时间",
	})
	w.commandLines = append(w.commandLines, &header.CommandLine{
		Command: "getuser",
		En:      "The getuser command allows you to view your account information, which remains private and inaccessible to others ",
		Zh:      "getuser命令，允许您查看您的帐户信息，这些信息仍然是私人的，其他人无法访问",
	})
	w.commandLines = append(w.commandLines, &header.CommandLine{
		Command: "make",
		En:      "By using the make command, you can view all the currently supported commands and the functions that each command can perform ",
		Zh:      "通过使用make命令，您可以查看当前支持的所有命令以及每个命令可以执行的功能",
	})
	w.commandLines = append(w.commandLines, &header.CommandLine{
		Command: "help",
		En:      "The help command has the same functionality as the make command",
		Zh:      "help命令与make命令具有相同的功能",
	})
	w.store.AddCommandLine(w.commandLines)
}

func (w *Manager) channelCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	funName := "channelCommand"
	content := fmt.Sprintf("**%s please first use /register to sign up. After registering, you will be able to use all the AI features provided by the Bot**", configure.SingtonConfigure().Application.Name)
	log4plus.Info(content)
	keys, values := w.store.GetCommandLine()
	for index, value := range values {
		log4plus.Info("%s keys[index]=[%s] value=[%s]", funName, keys[index], value)
		content = makeContent(content, keys[index], value)
	}

	/*Subscription 的指令*/
	// discordId := common.GetUserId(s, i)
	// disabledPayment := false
	// err, exist, user := w.store.GetUserBase(discordId)
	// if err != nil {
	// 	disabledPayment = true
	// }
	// if !exist {
	// 	disabledPayment = true
	// } else {
	// 	disabledPayment = false
	// }
	// header.MonthPaymentButton.URL = header.MakePaymentUrl(configure.SingtonConfigure().Stripe.PaymentUrl, discordId, header.MonthPrice, user.EMail)
	// header.MonthPaymentButton.Disabled = disabledPayment

	// header.YearPaymentButton.URL = header.MakePaymentUrl(configure.SingtonConfigure().Stripe.PaymentUrl, discordId, header.YearPrice, user.EMail)
	// header.YearPaymentButton.Disabled = disabledPayment

	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: makeData("make/help Command", "Command Description", content, common.GetColor()),
	}
	if err := s.InteractionRespond(i.Interaction, response); err != nil {
		log4plus.Error("Embed responseMsg Failed err=%s", err.Error())
	}
}

func (w *Manager) command(funName string, s *discordgo.Session, i *discordgo.InteractionCreate) {
	content := fmt.Sprintf("**%s please first use /register to sign up. After registering, you will be able to use all the AI features provided by the Bot**", configure.SingtonConfigure().Application.Name)
	log4plus.Info(content)

	keys, values := w.store.GetCommandLine()
	for index, value := range values {
		content = makeContent(content, keys[index], value)
	}
	/*Subscription 的指令*/
	discordId := common.GetUserId(s, i)
	disabledPayment := false
	err, exist, user := w.store.GetUserBase(discordId)
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
	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: makeData("make/help Command", "Command Description", content, common.GetColor()),
	}
	if err = s.InteractionRespond(i.Interaction, response); err != nil {
		log4plus.Error("Embed responseMsg Failed err=%s", err.Error())
	}
}

func (w *Manager) helpCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	funName := "helpCommand"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s consumption time=%d(s)", funName, time.Now().Unix()-now)
	}()
	w.channelCommand(s, i)
}

func (w *Manager) makeCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	w.channelCommand(s, i)
}

func (w *Manager) register(s *discordgo.Session, i *discordgo.InteractionCreate) {
	funName := "/register"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s consumption time=%d(s)", funName, time.Now().Unix()-now)
	}()
	//获取用户id
	discordId := common.GetUserId(s, i)
	log4plus.Info("%s userId=[%s]", funName, discordId)
	//解析参数
	var email string = ""
	if i.Type == discordgo.InteractionApplicationCommand {
		command := i.ApplicationCommandData()
		for _, v := range command.Options {
			if v.Name == "email" {
				email = v.StringValue()
				break
			}
		}
	}
	if strings.Trim(email, " ") == "" {
		log4plus.Error("%s Failed email is Empty", funName)
		return
	}
	if err := db.SingtonUserDB().CreateUser(discordId, email); err != nil {
		log4plus.Error("%s CreateUser Failed err=[%s]", funName, err.Error())
		common.SetCommandErrorResult(fmt.Sprintf("%s\nemail：[%s]\n", funName, email), "Register a new user", "An error occurred while retrieving your Discord ID", s, i)
		return
	}
	err, userInfo := db.SingtonUserDB().GetUser(discordId)
	if err != nil {
		log4plus.Error("%s GetUser Failed err=[%s]", funName, err.Error())
		common.SetCommandErrorResult(fmt.Sprintf("%s\nemail：[%s]\n", funName, email), "Register a new user", "An error occurred while retrieving your Discord ID", s, i)
		return
	}
	result := fmt.Sprintf("discordId: [%s]\nuserKey：[%s]\neMail：[%s]\nremainingTime：[%d]\nsubscribed：[%d]\ncreateTime：[%s]\nstate：[%d]", userInfo.DiscordId, userInfo.UserKey, userInfo.EMail, userInfo.RemainingTime, userInfo.Subscribed, userInfo.CreateTime, userInfo.State)
	common.SetCommandResult(fmt.Sprintf("%s\nemail：[%s]\n", funName, email), "Register a new user", result, s, i)
}

func (w *Manager) getuser(s *discordgo.Session, i *discordgo.InteractionCreate) {
	funName := "/getuser"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s consumption time=%d(s)", funName, time.Now().Unix()-now)
	}()
	//获取用户id
	discordId := common.GetUserId(s, i)
	log4plus.Info("%s getUserId discordId=[%s]", funName, discordId)

	err, userInfo := db.SingtonUserDB().GetUser(discordId)
	if err != nil {
		log4plus.Error("%s GetUser Failed err=[%s]", funName, err.Error())
		common.SetCommandErrorResult(fmt.Sprintf("%s\n", funName), "Get user details", "Failed to get user details", s, i)
		return
	}
	result := fmt.Sprintf("discordId: [%s]\nuserKey：[%s]\neMail：[%s]\nremainingTime：[%d]\nsubscribed：[%d]\ncreateTime：[%s]\nstate：[%d]",
		userInfo.DiscordId, userInfo.UserKey, userInfo.EMail, userInfo.RemainingTime, userInfo.Subscribed, userInfo.CreateTime, userInfo.State)
	common.SetCommandResult(fmt.Sprintf("%s\n", funName), "Get user details", result, s, i)
}

func SingtonManager(store header.PluginStore) *Manager {
	if nil == gManager {
		gManager = &Manager{
			store: store,
		}
		gManager.setCommand()
		gManager.setFuncs()
		gManager.initCommand()
	}
	return gManager
}
