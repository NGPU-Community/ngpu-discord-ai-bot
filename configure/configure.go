package configure

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	log4plus "github.com/nGPU/common/log4go"
)

type MysqlConfig struct {
	MysqlIp        string `json:"Ip"`       //Mysql的IP地址
	MysqlPort      int    `json:"Port"`     //Mysql的端口
	MysqlDBName    string `json:"DBName"`   //DBName
	MysqlDBCharset string `json:"Charset"`  //数据库字符集
	UserName       string `json:"UserName"` //用户明
	Password       string `json:"Password"` //用户密码
}

type DiscordConfig struct {
	Token            string `json:"token"`
	GuildId          string `json:"guildId"`
	WelcomeChannelId string `json:"welcomeChannelId"`
}

type TelegramConfig struct {
	Token string `json:"token"`
}

type TokenConfig struct {
	Discord  DiscordConfig  `json:"discord"`
	Telegram TelegramConfig `json:"telegram"`
}

type ResourceConfig struct {
	Listen       string `json:"listen"`
	Domain       string `json:"domain"`
	ResourcePath string `json:"resourcePath"`
}

type StripeConfig struct {
	Listen     string `json:"listen"`
	PaymentUrl string `json:"paymentUrl"`
	PrivateKey string `json:"privateKey"`
}

type ManagerConfig struct {
	RegisterUri string `json:"register"`
	UserBaseUri string `json:"getUser"`
}

type BtcConfig struct {
	BtcAmtUri string `json:"btcAmt"`
}

/*************/
type ChannelBase struct {
	ChannelId   string `json:"channelId"`
	ChannelName string `json:"channelName"`
}

type TwitterBase struct {
	Key       string `json:"key"`
	Host      string `json:"host"`
	ChannelId string `json:"channelId"`
}

type TwitterUrls struct {
	Details   string `json:"details"`
	Following string `json:"following"`
}

type TwitterConfig struct {
	Base TwitterBase `json:"base"`
	Urls TwitterUrls `json:"urls"`
}

/*************/
type ChatTTSUrls struct {
	GenerateAudio MethodBase `json:"generateAudio"`
	CustomAudio   MethodBase `json:"customAudio"`
}

type ChatTTSConfig struct {
	Base ChannelBase `json:"base"`
	Urls ChatTTSUrls `json:"urls"`
}

/*************/
type MethodBase struct {
	MethodMode int    `json:"methodMode"`
	MethodUrl  string `json:"methodUrl"`
	GetTask    string `json:"getTask"`
	Comment    string `json:"comment"`
}

type FaceFusionUrls struct {
	BlackWhite2Color MethodBase `json:"blackWhite2Color"`
	LipSyncer        MethodBase `json:"lipSyncer"`
	FaceSwap         MethodBase `json:"faceSwap"`
	FrameEnhance     MethodBase `json:"frameEnhance"`
}

type FaceFusionConfig struct {
	Base ChannelBase    `json:"base"`
	Urls FaceFusionUrls `json:"urls"`
}

/*************/
type SadTalkerUrls struct {
	SadTalker MethodBase `json:"sadTalker"`
}

type SadTalkerConfig struct {
	Base ChannelBase   `json:"base"`
	Urls SadTalkerUrls `json:"urls"`
}

/*************/
type BlipUrls struct {
	Blip MethodBase `json:blip"`
}

type BlipConfig struct {
	Base ChannelBase `json:"base"`
	Urls BlipUrls    `json:"urls"`
}

/*************/
type StableDiffusionUrls struct {
	StableDiffusion MethodBase `json:stableDiffusion"`
}

type StableDiffusionConfig struct {
	Base ChannelBase         `json:"base"`
	Urls StableDiffusionUrls `json:"urls"`
}

/*************/
type LLMUrls struct {
	LLM MethodBase `json:llm"`
}

type LLMConfig struct {
	Base ChannelBase `json:"base"`
	Urls LLMUrls     `json:"urls"`
}

/*************/
type RemoveBGUrls struct {
	RemoveBG MethodBase `json:removeBG"`
}

type RemoveBGConfig struct {
	Base ChannelBase  `json:"base"`
	Urls RemoveBGUrls `json:"urls"`
}

/*************/

type ReplaceBGUrls struct {
	ReplaceBG MethodBase `json:replaceBG"`
}

type ReplaceBGConfig struct {
	Base ChannelBase   `json:"base"`
	Urls ReplaceBGUrls `json:"urls"`
}

type InterfacesConfig struct {
	Manager         ManagerConfig         `json:"manager"`
	BTC             BtcConfig             `json:"btc"`
	FaceFusion      FaceFusionConfig      `json:"faceFusion"`
	SadTalker       SadTalkerConfig       `json:"sadTalker"`
	StableDiffusion StableDiffusionConfig `json:"stableDiffusion"`
	Blip            BlipConfig            `json:"blip"`
	LLm             LLMConfig             `json:"llm"`
	RemoveBG        RemoveBGConfig        `json:"removeBG"`
	ReplaceBG       ReplaceBGConfig       `json:"replaceBG"`
}

type ApplicationConfig struct {
	Name    string `json:"name"`
	Comment string `json:"comment"`
}

/*************/
type AiModule struct {
	ApiKey string `json:apiKey"`
}

type AiModuleConfig struct {
	Anthropic AiModule `json:"anthropic"`
	ChatGPT   AiModule `json:"chatGPT"`
}

type Config struct {
	Application ApplicationConfig `json:"application"`
	Token       TokenConfig       `json:"token"`
	Stripe      StripeConfig      `json:"stripe"`
	Resource    ResourceConfig    `json:"resource"`
	Mysql       MysqlConfig       `json:"mysql"`
	Interfaces  InterfacesConfig  `json:"interfaces"`
	AiModules   AiModuleConfig    `json:"aiModules"`
}

type ConfigureManager struct {
	config Config
}

func GetJsonFileName() string {
	// return fmt.Sprintf("%s%s", time.Now().Format("20060102_15_04_05"), fmt.Sprintf("_%06d", time.Now().Nanosecond()/1e3))
	return fmt.Sprintf("configCheck.json")
}

var configureManager *ConfigureManager

func (u *ConfigureManager) getConfig() error {
	funName := "getConfig"
	log4plus.Info("%s ---->>>>", funName)
	data, err := ioutil.ReadFile("./config.json")
	if err != nil {
		log4plus.Error("%s ReadFile error=[%+v]", funName, err)
		return err
	}
	log4plus.Info("%s data=[%s]", funName, string(data))
	err = json.Unmarshal(data, &u.config)
	if err != nil {
		log4plus.Error("%s json.Unmarshal error=[%+v]", funName, err)
		return err
	}

	fileName := fmt.Sprintf("%s", GetJsonFileName())
	file, err := os.Create(fileName)
	if err != nil {
		log4plus.Error("%s Create error=[%+v]", funName, err)
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(u.config)
	if err != nil {
		log4plus.Error("%s Encode error=[%+v]", funName, err)
		return err
	}

	log4plus.Info("%s success---->>>>", funName)
	return nil
}

func SingtonConfigure() Config {
	if configureManager == nil {
		configureManager = &ConfigureManager{}
		configureManager.getConfig()
	}
	return configureManager.config
}
