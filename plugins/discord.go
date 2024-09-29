package plugins

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/nGPU/bot/configure"
	"github.com/nGPU/bot/db"
	"github.com/nGPU/bot/header"
	"github.com/nGPU/bot/plugins/discord"
	log4plus "github.com/nGPU/common/log4go"
)

type Discord struct {
	roots              *x509.CertPool
	rootPEM            []byte
	dg                 *discordgo.Session
	msgSession         *discordgo.Session
	msgChannelId       *discordgo.MessageCreate
	msgFuncs           map[string]header.DiscordMsgFunction
	registeredCommands map[int]*discordgo.ApplicationCommand
	commandLines       map[string]string
	commands           []*discordgo.ApplicationCommand
}

var gSession *discordgo.Session
var gDiscord *Discord

func (w *Discord) Version() string {
	return "Discord Version 1.0.0"
}

func (w *Discord) RegisteredCommand(cmd *discordgo.ApplicationCommand) {
	w.commands = append(w.commands, cmd)
}

func (w *Discord) SetDiscordFunction(funName string, fun header.DiscordMsgFunction) {
	w.msgFuncs[strings.ToLower(funName)] = fun
}

func (w *Discord) SetTelegramFunction(funName string, fun header.TelegramMsgFunction) {
}

func (w *Discord) AddCommandLine(commands []*header.CommandLine) {
	for _, v := range commands {
		cmd := ":point_right: " + v.Command
		content := fmt.Sprintf("%s\n%s", v.En, v.Zh)
		w.commandLines[cmd] = content
	}
}

func (w *Discord) GetCommandLine() ([]string, []string) {
	var keys []string
	var values []string
	for k, v := range w.commandLines {
		keys = append(keys, k)
		values = append(values, v)
	}
	return keys, values
}

func (w *Discord) CheckBTCAddress(btcAddress string) bool {
	funName := "CheckBTCAddress"
	err, exist := w.getBRC20(btcAddress, w.roots)
	if err != nil {
		log4plus.Error(fmt.Sprintf("%s getBRC20 Failed err=[%s]", funName, err.Error()))
		return false
	}
	return exist
}

func (w *Discord) GetUserBase(discordId string) (error, bool, header.ResponseUserInfo) {
	funName := "GetUserBase"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s discordId=[%s] consumption time=%d(s)", funName, discordId, time.Now().Unix()-now)
	}()
	err, userInfo := db.SingtonUserDB().GetUser(discordId)
	if err != nil {
		return err, false, header.ResponseUserInfo{}
	}
	return nil, true, userInfo
}

func (w *Discord) getBRC20(BtcAddress string, roots *x509.CertPool) (error, bool) {
	funName := "getBRC20"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s BtcAddress=[%s] consumption time=%d(s)", funName, BtcAddress, time.Now().Unix()-now)
	}()
	tr := &http.Transport{
		TLSClientConfig:     &tls.Config{RootCAs: roots},
		TLSHandshakeTimeout: 30 * time.Second,
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Duration(30) * time.Second,
	}
	defer client.CloseIdleConnections()

	var reqUri = fmt.Sprintf("%s?address=%s", configure.SingtonConfigure().Interfaces.BTC.BtcAmtUri, BtcAddress)
	req, err := http.NewRequest("GET", reqUri, nil)
	log4plus.Info("%s http.NewRequest url=[%s]", funName, reqUri)

	response, err := client.Do(req)
	if err != nil {
		log4plus.Error("%s client.Do Failed err=[%s] url=[%s]", funName, err.Error(), reqUri)
		return err, false
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusOK {
		log4plus.Info("%s StatusCode is http.StatusOK response=[%+v]", funName, *response)
		body, errRes := ioutil.ReadAll(response.Body)
		if errRes != nil {
			log4plus.Error("%s ReadAll Failed err=[%s] url=[%s]", funName, errRes.Error(), reqUri)
			return errRes, false
		}
		brcResponse := new(struct {
			Exists   bool     `json:"exists"`   //是否存在指定的铭文
			Names    []string `json:"names"`    //名称
			Balances []int64  `json:"balances"` //数量
		})
		if err := json.Unmarshal(body, &brcResponse); err != nil {
			log4plus.Error("%s ReadAll Failed err=[%s] url=[%s]", funName, err.Error(), reqUri)
			return err, false
		}
		log4plus.Info("%s Result=[%+v]", funName, *brcResponse)
		return nil, brcResponse.Exists

	} else if response.StatusCode == 400 {
		log4plus.Info("%s StatusCode=[%d] BtcAddress=[%s]", funName, response.StatusCode, BtcAddress)
		return nil, false
	}
	log4plus.Error("%s response=[%+v]", funName, *response)
	errString := fmt.Sprintf("%s StatusCode=[%d]", funName, response.StatusCode)
	return errors.New(errString), false
}

func (w *Discord) responseMsg(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) error {
	funName := "responseMsg"
	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
		},
	}
	err := s.InteractionRespond(i.Interaction, response)
	if err != nil {
		log4plus.Error("%s InteractionRespond Failed er=[%s]", funName, err.Error())
		return err
	}
	return nil
}

func (w *Discord) discordInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	funName := "discordInteractionCreate"
	if i.Type == discordgo.InteractionApplicationCommand {
		log4plus.Info("---->>>>%s i.ApplicationCommandData()=[%+v] CommandName=[%s]", funName, i.ApplicationCommandData(), strings.ToLower(i.ApplicationCommandData().Name))

		for key, _ := range w.msgFuncs {
			log4plus.Info("---->>>>%s key=[%s]", funName, key)
		}
		if msgFun, ok := w.msgFuncs[strings.ToLower(i.ApplicationCommandData().Name)]; ok {
			msgFun(s, i)
		}
	}
}

func (w *Discord) discordGuildMemberAdd(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	funName := "discordGuildMemberAdd"
	welcomeMessage := fmt.Sprintf("Welcome to the Ais server, %s! We're glad to have you here.", m.User.Mention())
	_, err := s.ChannelMessageSend(configure.SingtonConfigure().Token.Discord.WelcomeChannelId, welcomeMessage)
	if err != nil {
		log4plus.Info("%s ChannelMessageSend Failed err=[%s]", funName, err)
	}
}

func (w *Discord) SetSession(s *discordgo.Session, m *discordgo.MessageCreate) {
	w.msgChannelId = m
	w.msgSession = s
}

func (w *Discord) WaitingTaskId(taskId string) (error, string) {
	funName := "WaitingTaskId"
	// waiting
	maxTimeout := int64(5 * 60 * 60)
	for {
		time.Sleep(time.Duration(5) * time.Second)
		maxTimeout = maxTimeout - 10
		if maxTimeout <= 0 {
			return errors.New("task timeout"), ""
		}
		err, taskInformation := db.SingtonAPITasksDB().GetTaskId(taskId)
		if err != nil {
			return err, ""
		}
		if taskInformation.State == header.FinishState {
			return nil, taskInformation.Response
		} else if taskInformation.State == header.ErrorState {
			type ResponseResult struct {
				ResultCode interface{} `json:"result_code"`
			}
			var result ResponseResult
			json.Unmarshal([]byte(taskInformation.Response), &result)
			if result.ResultCode == nil {
				errString := fmt.Sprintf("%s result.ResultCode is nil", funName)
				log4plus.Error(errString)
				return errors.New(errString), ""
			}
			errString := fmt.Sprintf("%s Ai Module Result Code is [%d]", funName, int(result.ResultCode.(float64)))
			log4plus.Error(errString)
			return errors.New(errString), ""
		}
	}
}

func (w *Discord) getUrlExtension(urlString string) string {
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

func (w *Discord) DownloadFile(url string) (err error, newUrl string, localPath string) {
	funName := "DownloadFile"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s url=%s consumption time=%d(s)", funName, url, time.Now().Unix()-now)
	}()

	response, err := http.Get(url)
	if err != nil {
		log4plus.Error("%s Get Failed err=%s", funName, err.Error())
		return err, "", ""
	}
	defer response.Body.Close()

	fileExt := w.getUrlExtension(url)
	fileName := fmt.Sprintf("%s%s.%s", time.Now().Format("20060102150405"), fmt.Sprintf("%06d", time.Now().Nanosecond()/1e3), fileExt)
	localPath = fmt.Sprintf("%s/%s", configure.SingtonConfigure().Resource.ResourcePath, fileName)
	file, err := os.Create(localPath)
	if err != nil {
		log4plus.Error("%s Create Failed err=%s", funName, err.Error())
		return err, "", ""
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		log4plus.Error("%s Copy Failed err=%s", funName, err.Error())
		return err, "", ""
	}
	newUrl = fmt.Sprintf("%s/%s", configure.SingtonConfigure().Resource.Domain, fileName)
	return nil, newUrl, localPath
}

func (w *Discord) Init() {
	funName := "Init"
	flag.Parse()
	botString := fmt.Sprintf("Bot %s", configure.SingtonConfigure().Token.Discord.Token)
	log4plus.Info("%s botString=[%s]", funName, botString)
	tmpSession, err := discordgo.New(botString)
	if err != nil {
		log4plus.Error("%s discordgo.New Failed err=[%s]", funName, err.Error())
		return
	}

	//添加对消息的监听
	tmpSession.AddHandler(w.discordInteractionCreate)
	tmpSession.AddHandler(w.discordGuildMemberAdd)

	//在本例中，我们只关心接收消息事件。
	tmpSession.Identify.Intents = discordgo.IntentsGuildMessages
	//打开discord
	if err = tmpSession.Open(); err != nil {
		log4plus.Error("%s tmpSession.Open Failed err=[%s]", funName, err.Error())
		return
	}
	gSession = tmpSession

	//获取服务器的所有频道
	// channelID, err := discordgo.GuildAllChannelsID(configure.SingtonConfigure().Token.DiscrodGuildId)
	// if err != nil {
	// 	log4plus.Error("%s GuildAllChannelsID Failed err=[%s]", funName, err.Error())
	// 	return
	// }
	// log4plus.Info("%s GuildAllChannelsID DiscrodGuildId=[%s] channelID=[%s]", funName, configure.SingtonConfigure().Token.DiscrodGuildId, channelID)

	//先删除之前注册的指令
	// log4plus.Info("%s Deleteing commands...", funName)
	// for _, command := range w.registeredCommands {
	// 	//err := gSession.ApplicationCommandDelete(gSession.State.User.ID, configure.SingtonConfigure().Token.DiscrodGuildId, command.ID)
	// 	err := gSession.ApplicationCommandDelete(gSession.State.User.ID, "", command.ID)
	// 	if err != nil {
	// 		log4plus.Error("Cannot delete command.Name=[%s] err=[%s]", command.Name, err.Error())
	// 	}
	// }

	//注册指令信息
	log4plus.Info("%s Registered commands...", funName)
	for i, command := range w.commands {
		log4plus.Info("%s command.Name=[%s] ", funName, command.Name)
		// cmd, err := gSession.ApplicationCommandCreate(gSession.State.User.ID, configure.SingtonConfigure().Token.DiscrodGuildId, v)
		cmd, err := gSession.ApplicationCommandCreate(gSession.State.User.ID, "", command)
		if err != nil {
			log4plus.Error("%s ApplicationCommandCreate Failed v.Name=[%s] err=[%s]", funName, command.Name, err.Error())
			return
		}
		// log4plus.Info("%s v.Name=[%s] cmd=[%+v]", funName, command.Name, cmd)
		w.registeredCommands[i] = cmd
	}

	//等待程序结束
	log4plus.Info("%s Discord Bot is now running. Press CTRL+C to exit.", funName)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	//关闭Discord会话
	gSession.Close()
	//删除注册的指令
	log4plus.Info("Removing commands...")
	for _, command := range w.registeredCommands {
		// err := gSession.ApplicationCommandDelete(gSession.State.User.ID, configure.SingtonConfigure().Token.DiscrodGuildId, command.ID)
		err := gSession.ApplicationCommandDelete(gSession.State.User.ID, "", command.ID)
		if err != nil {
			log4plus.Error("Cannot delete command.Name=[%s] err=[%s]", command.Name, err.Error())
		}
	}
}

func SingtonDiscord() *Discord {
	if nil == gDiscord {
		if strings.Trim(configure.SingtonConfigure().Token.Discord.Token, " ") == "" {
			log4plus.Error("<<<<---->>>> Discord Token is empty")
			return nil
		}
		log4plus.Info("Discord Token is [%s]", configure.SingtonConfigure().Token.Discord.Token)

		gDiscord = &Discord{
			msgChannelId:       nil,
			msgSession:         nil,
			msgFuncs:           make(map[string]header.DiscordMsgFunction),
			registeredCommands: make(map[int]*discordgo.ApplicationCommand),
			commandLines:       make(map[string]string),
		}

		log4plus.Info("discord.SingtonFaceFusion(gDiscord)")
		discord.SingtonFaceFusion(gDiscord)

		log4plus.Info("discord.SingtonSadTalker(gDiscord)")
		discord.SingtonSadTalker(gDiscord)

		log4plus.Info("discord.SingtonBlip(gDiscord)")
		discord.SingtonBlip(gDiscord)

		log4plus.Info("discord.SingtonRemoveBG(gDiscord)")
		discord.SingtonRemoveBG(gDiscord)

		log4plus.Info("discord.SingtonStableDiffusion(gDiscord)")
		discord.SingtonStableDiffusion(gDiscord)

		log4plus.Info("discord.SingtonFengshui(gDiscord)")
		discord.SingtonFengshui(gDiscord)

		// log4plus.Info("discord.SingtonManager(gDiscord)")
		// discord.SingtonManager(gDiscord)

		log4plus.Info("go gDiscord.Init()")
		go gDiscord.Init()
	}
	return gDiscord
}
