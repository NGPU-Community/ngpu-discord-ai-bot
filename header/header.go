package header

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	tele "gopkg.in/telebot.v3"
)

// const (
// 	DiscordApplication = "nGPU"
// 	DiscordName        = "Discord Bot for nGPU"
// 	TwitterUrl         = "https://twitter.com/chain_fxh7622/status"
// )

const (
	Base               = iota + 300
	JsonParseError     = Base + 1
	CheckApiKeyError   = Base + 2
	ApiKeyTimeoutError = Base + 3
	ParamsIsNilError   = Base + 4
	TaskNotFoundError  = Base + 5

	BlipBase             = Base + 10
	ChatTTSBase          = Base + 20
	FaceFusionBase       = Base + 30
	LLMBase              = Base + 40
	SadTalkerBase        = Base + 50
	StableDiffustionBase = Base + 60
	RemoveBase           = Base + 70
	FengshuiBase         = Base + 80
)

const (
	TaskStateBase     = iota + 0
	ErrorState        = TaskStateBase - 1
	InitState         = TaskStateBase + 1
	RunningState      = TaskStateBase + 2
	FinishState       = TaskStateBase + 3
	IntermediateState = TaskStateBase + 4 //the state when the task has been submitted and returned but is still processing subsequent tasks
)

type ResponseCheck struct {
	ResultCode int64  `json:"result_code"`
	Msg        string `json:"msg"`
}

type CommandLine struct {
	Command string `json:"command"`
	En      string `json:"en"`
	Zh      string `json:"zh"`
}

type TaskInfo struct {
	TaskId        string `json:"taskId"`
	State         int    `json:"state"`
	RequestTime   string `json:"requesttime"`
	Method        string `json:"method"`
	Response      string `json:"response"`
	RecordDursion int64  `json:"recordDursion"`
}

type QueryTaskResponse struct {
	ReturnCode  int64  `json:"result_code"`
	ReturnMsg   string `json:"msg"`
	ResultSize  int    `json:"result_size"`
	TaskDursion int64  `json:"task_dursion"`
	Data        string `json:"data"`
}

type RequestData struct {
	BTCAddr string          `json:"btc_address"`
	Data    json.RawMessage `json:"data"`
}

/*Discord Plugin*/
type DiscordMsgFunction func(s *discordgo.Session, i *discordgo.InteractionCreate)
type DiscordPluginStore interface {
	RegisteredCommand(cmd *discordgo.ApplicationCommand)
	SetDiscordFunction(funName string, fun DiscordMsgFunction)
	AddCommandLine(commands []*CommandLine)
	GetCommandLine() ([]string, []string)
	CheckBTCAddress(btcAddress string) bool
	GetUserBase(discordId string) (error, bool, ResponseUserInfo)
	WaitingTaskId(taskId string) (error, string)
	DownloadFile(url string) (err error, newUrl string, localPath string)
}

/*Telegram Plugin*/

type Step struct {
	StepId      int             `json:"stepId"`
	RequestTime time.Time       `json:"request"`
	Data        json.RawMessage `json:"data"`
}

type UserStep struct {
	TelegramId   int64  `json:"telegramId"`
	FunctionName string `json:"functionName"`
	MaxStep      int    `json:"maxStep"`
	MessageId    int    `json:"messageId"`
	Steps        []Step `json:"steps"`
}

type TelegramMsgFunction func(c tele.Context) error
type TelegramInlineFunction func(c tele.Context, data string) error
type TelegramPluginStore interface {
	GetBotObject() *tele.Bot
	SetTelegramFunction(major, minor, showTxt string, fun TelegramMsgFunction)
	SetInlineFunction(major, data string, fun TelegramInlineFunction)
	ClearInlineFunction()
	CheckBTCAddress(btcAddress string) bool
	WaitingTaskId(taskId string) (error, string)
	SaveMediaFile(fileId string) (error, string)
	DownloadFile(url string) (err error, newUrl string, localPath string)

	AddUser(user *UserStep) error
	DeleteUser(telegramId int64)
	FindUser(telegramId int64) *UserStep
}

var (
	HomeButton = discordgo.Button{
		Emoji: &discordgo.ComponentEmoji{
			Name: "🌐",
		},
		Label: "nGPU Home",
		Style: discordgo.LinkButton,
		URL:   fmt.Sprintf("https://www.ngpu.ai"),
	}

	MonthPrice         = "10"
	monthLable         = fmt.Sprintf("subscription $%s/month", MonthPrice)
	MonthPaymentButton = discordgo.Button{
		Emoji: &discordgo.ComponentEmoji{
			Name: "📜",
		},
		Label: monthLable,
		Style: discordgo.LinkButton,
	}

	YearPrice         = "100"
	yearLable         = fmt.Sprintf("subscription $%s/year", YearPrice)
	YearPaymentButton = discordgo.Button{
		Emoji: &discordgo.ComponentEmoji{
			Name: "🟨",
		},
		Label: yearLable,
		Style: discordgo.LinkButton,
	}
)

func MakePaymentUrl(paymentUrl string, discordId string, price string, eMail string) string {
	url := fmt.Sprintf("%s/user/payment?discordId=%s&paymentMode=0&paymentPrice=%s&Email=%s", paymentUrl, discordId, price, eMail)
	return url
}

func GetTaskId() string {
	return fmt.Sprintf("%s%s", time.Now().Format("20060102_15_04_05"), fmt.Sprintf("_%06d", time.Now().Nanosecond()/1e3))
}
