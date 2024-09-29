package common

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/nGPU/bot/configure"
	log4plus "github.com/nGPU/common/log4go"
)

func GetColor() int {
	// 设置红色和蓝色通道的值，绿色通道的值为0
	red := 128  // 可以根据需要调整红色通道的值
	blue := 255 // 可以根据需要调整蓝色通道的值
	green := 0  // 绿色通道的值为0
	purple := (red << 16) | (green << 8) | blue
	return purple
}

func GetRedColor() int {
	// 设置红色和蓝色通道的值，绿色通道的值为0
	red := 255 // 可以根据需要调整红色通道的值
	blue := 0  // 可以根据需要调整蓝色通道的值
	green := 0 // 绿色通道的值为0
	purple := (red << 16) | (green << 8) | blue
	return purple
}

func GetGreenColor() int {
	// 设置红色和蓝色通道的值，绿色通道的值为0
	red := 0     // 可以根据需要调整红色通道的值
	blue := 0    // 可以根据需要调整蓝色通道的值
	green := 255 // 绿色通道的值为0
	purple := (red << 16) | (green << 8) | blue
	return purple
}

func SetCommandErrorResult(title, description, content string, s *discordgo.Session, i *discordgo.InteractionCreate) {
	funName := "setCommandErrorResult"

	// 构建交互回应数据
	data := &discordgo.InteractionResponseData{
		Content: title,
		Embeds: []*discordgo.MessageEmbed{&discordgo.MessageEmbed{
			Title:       description,
			Description: content,
			Color:       GetRedColor(),
		}},
	}
	log4plus.Info("%s Concatenate response data title=[%s] description=[%s] content=[%s]", funName, title, description, content)

	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: data,
	}
	if err := s.InteractionRespond(i.Interaction, response); err != nil {
		log4plus.Error("Embed responseMsg Failed err=%s", err.Error())
	}
	log4plus.Info("%s InteractionRespond Success", funName)
}

func SetCommandResult(title, description, content string, s *discordgo.Session, i *discordgo.InteractionCreate) {
	funName := "setCommandResult"

	// 构建交互回应数据
	data := &discordgo.InteractionResponseData{
		Content: title,
		Embeds: []*discordgo.MessageEmbed{&discordgo.MessageEmbed{
			Title:       description,
			Description: content,
			Color:       GetColor(),
		}},
	}
	log4plus.Info("%s Concatenate response data title=[%s] description=[%s] content=[%s]", funName, title, description, content)

	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: data,
	}
	if err := s.InteractionRespond(i.Interaction, response); err != nil {
		log4plus.Error("Embed responseMsg Failed err=%s", err.Error())
	}
	log4plus.Info("%s InteractionRespond Success", funName)
}

// 回答视频正确，分享视频到推特
func AnswerVideoSuccessWithButton(cmd string, txt string, shareContent string, userID string, s *discordgo.Session, i *discordgo.InteractionCreate, roots *x509.CertPool) string {
	funName := "answerVideoSuccessWithButton"
	content := fmt.Sprintf("%s%s", cmd, txt)
	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	}); err != nil {
		log4plus.Error("%s Failed userId=%s txt=%s err=%s", funName, userID, txt, err.Error())
		if _, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: content,
		}); err != nil {
			log4plus.Error("%s FollowupMessageCreate Failed err=%s", funName, err.Error())
		}
		return ""
	}
	return content
}

// 设置应答信息
func SetFirstPrivate(txt string, userID string, s *discordgo.Session, i *discordgo.InteractionCreate) string {
	funName := "setFirstPrivate"
	content := fmt.Sprintf("%s", txt)
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:   discordgo.MessageFlagsEphemeral,
			Content: content,
		},
	}); err != nil {
		log4plus.Error("%s Failed userId=%s txt=%s err=%s", funName, userID, txt, err.Error())
		if _, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: content,
		}); err != nil {
			log4plus.Error("%s FollowupMessageCreate Failed err=%s", funName, err.Error())
		}
		return ""
	}
	return content
}

// 设置应答信息
func SetAnswerSuccessPrivate(cmd string, txt string, userID string, s *discordgo.Session, i *discordgo.InteractionCreate) string {
	funName := "setAnswerSuccessPrivate"
	content := fmt.Sprintf("%s%s", cmd, txt)
	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	}); err != nil {
		log4plus.Error("%s Failed userId=%s txt=%s err=%s", funName, userID, txt, err.Error())
		if _, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: content,
		}); err != nil {
			log4plus.Error("%s FollowupMessageCreate Failed err=%s", funName, err.Error())
		}
		return ""
	}
	return content
}

// 设置应答信息
func SetLastPrivate(txt string, userID string, s *discordgo.Session, i *discordgo.InteractionCreate) string {
	funName := "setLastPrivate"
	content := fmt.Sprintf("%s", txt)
	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	}); err != nil {
		log4plus.Error("%s Failed userId=%s txt=%s err=%s", funName, userID, txt, err.Error())
		if _, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: content,
		}); err != nil {
			log4plus.Error("%s FollowupMessageCreate Failed err=%s", funName, err.Error())
		}
		return ""
	}
	return content
}

func GetFileExtension(urlString string) string {
	funName := "getFileExtension"
	parsedURL, err := url.Parse(urlString)
	if err != nil {
		// 处理URL解析错误
		log4plus.Error("%s Failed err=%s", funName, err.Error())
		return ""
	}
	// 从路径中获取文件名
	fileName := path.Base(parsedURL.Path)
	// 获取文件后缀名
	fileExtension := strings.TrimPrefix(path.Ext(fileName), ".")
	log4plus.Info("%s fileExtension=%s", funName, fileExtension)
	return fileExtension
}

func DownloadFile(url string) (error, string) {
	funName := "downloadFile"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s url=%s consumption time=%d(s)", funName, url, time.Now().Unix()-now)
	}()
	//下载文件
	response, err := http.Get(url)
	if err != nil {
		log4plus.Error("%s Get Failed err=%s", funName, err.Error())
		return err, ""
	}
	defer response.Body.Close()
	//创建文件
	fileExt := GetFileExtension(url)
	fileName := fmt.Sprintf("%s.%s", time.Now().Format("20060102150405000"), fileExt)
	filePath := fmt.Sprintf("%s/%s", configure.SingtonConfigure().Resource.ResourcePath, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		log4plus.Error("%s Create Failed err=%s", funName, err.Error())
		return err, ""
	}
	defer file.Close()
	//文件拷贝
	_, err = io.Copy(file, response.Body)
	if err != nil {
		log4plus.Error("%s Copy Failed err=%s", funName, err.Error())
		return err, ""
	}
	//生成新的Url地址
	newUrl := fmt.Sprintf("%s/%s", configure.SingtonConfigure().Resource.Domain, fileName)
	log4plus.Info("%s newUrl=%s", funName, newUrl)
	return nil, newUrl
}

func DownloadLocalFile(url string) (error, string) {
	funName := "downloadLocalFile"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s url=%s consumption time=%d(s)", funName, url, time.Now().Unix()-now)
	}()
	//下载文件
	response, err := http.Get(url)
	if err != nil {
		log4plus.Error("%s Get Failed err=%s", funName, err.Error())
		return err, ""
	}
	defer response.Body.Close()
	//创建文件
	fileExt := GetFileExtension(url)
	fileName := fmt.Sprintf("%s%s.%s", time.Now().Format("20060102150405"), fmt.Sprintf("%06d", time.Now().Nanosecond()/1e3), fileExt)
	filePath := fmt.Sprintf("%s/%s", configure.SingtonConfigure().Resource.ResourcePath, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		log4plus.Error("%s Create Failed err=%s", funName, err.Error())
		return err, ""
	}
	defer file.Close()
	//文件拷贝
	_, err = io.Copy(file, response.Body)
	if err != nil {
		log4plus.Error("%s Copy Failed err=%s", funName, err.Error())
		return err, ""
	}
	//生成新的Url地址
	newUrl := fmt.Sprintf("%s/%s", configure.SingtonConfigure().Resource.Domain, fileName)
	log4plus.Info("%s newUrl=%s filePath=%s", funName, newUrl, filePath)
	return nil, filePath
}

func DownloadCurPathFile(url string) (error, string) {
	funName := "downloadCurPathFile"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s url=%s consumption time=%d(s)", funName, url, time.Now().Unix()-now)
	}()
	//下载文件
	response, err := http.Get(url)
	if err != nil {
		log4plus.Error("%s Get Failed err=%s", funName, err.Error())
		return err, ""
	}
	defer response.Body.Close()
	//创建文件
	fileExt := GetFileExtension(url)
	fileName := fmt.Sprintf("%s%s.%s", time.Now().Format("20060102150405"), fmt.Sprintf("%06d", time.Now().Nanosecond()/1e3), fileExt)
	currentDir, err := os.Getwd()
	if err != nil {
		log4plus.Error("%s Getwd Failed err=%s", funName, err.Error())
		return err, ""
	}
	filePath := fmt.Sprintf("%s/images", currentDir)
	if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
		log4plus.Error("%s MkdirAll Failed err=%s", funName, err.Error())
		return err, ""
	}
	filePath = fmt.Sprintf("%s/%s", filePath, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		log4plus.Error("%s Create Failed err=%s", funName, err.Error())
		return err, ""
	}
	defer file.Close()
	//文件拷贝
	_, err = io.Copy(file, response.Body)
	if err != nil {
		log4plus.Error("%s Copy Failed err=%s", funName, err.Error())
		return err, ""
	}
	//生成新的Url地址
	newUrl := fmt.Sprintf("%s/%s", configure.SingtonConfigure().Resource.Domain, fileName)
	log4plus.Info("%s newUrl=%s filePath=%s", funName, newUrl, filePath)
	return nil, filePath
}

func DownloadUrlFile(url string) (error, string, string) {
	funName := "downloadUrlFile"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s url=%s consumption time=%d(s)", funName, url, time.Now().Unix()-now)
	}()
	//下载文件
	response, err := http.Get(url)
	if err != nil {
		log4plus.Error("%s Get Failed err=%s", funName, err.Error())
		return err, "", ""
	}
	defer response.Body.Close()
	//创建文件
	fileExt := GetFileExtension(url)
	fileName := fmt.Sprintf("%s%s.%s", time.Now().Format("20060102150405"), fmt.Sprintf("%06d", time.Now().Nanosecond()/1e3), fileExt)
	filePath := fmt.Sprintf("%s/%s", configure.SingtonConfigure().Resource.ResourcePath, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		log4plus.Error("%s Create Failed err=%s", funName, err.Error())
		return err, "", ""
	}
	defer file.Close()
	//文件拷贝
	_, err = io.Copy(file, response.Body)
	if err != nil {
		log4plus.Error("%s Copy Failed err=%s", funName, err.Error())
		return err, "", ""
	}
	//生成新的Url地址
	newUrl := fmt.Sprintf("%s/%s", configure.SingtonConfigure().Resource.Domain, fileName)
	log4plus.Info("%s newUrl=%s filePath=%s", funName, newUrl, filePath)
	return nil, filePath, newUrl
}

func GetUserId(s *discordgo.Session, i *discordgo.InteractionCreate) string {
	userID := ""
	if i.Interaction.User != nil {
		userID = i.Interaction.User.ID
	} else if i.Interaction.Member != nil {
		userID = i.Interaction.Member.User.ID
	}
	return userID
}

func GetBrcAmt(btcAddress string, roots *x509.CertPool) (error, bool, []string, []int64) {
	var names []string
	var balances []int64
	funName := "getBrcAmt"
	now := time.Now().Unix()
	defer func() {
		log4plus.Info("%s btcAddress=%s consumption time=%d(s)", funName, btcAddress, time.Now().Unix()-now)
	}()
	tr := &http.Transport{
		TLSClientConfig:     &tls.Config{RootCAs: roots},
		TLSHandshakeTimeout: 10 * time.Minute,
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Duration(10) * time.Minute,
	}
	defer client.CloseIdleConnections()
	var reqUri = fmt.Sprintf("%s?address=%s", configure.SingtonConfigure().Interfaces.BTC.BtcAmtUri, btcAddress)
	req, err := http.NewRequest("GET", reqUri, nil)
	log4plus.Info("%s client.Do(req) uri=%s", funName, reqUri)
	response, err := client.Do(req)
	if err != nil {
		log4plus.Error("%s client.Do Failed err: %s url: %s", funName, err.Error(), reqUri)
		return err, false, names, balances
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusOK {
		log4plus.Info("%s StatusCode is http.StatusOK body=%v", funName, response.Body)
		body, errRes := ioutil.ReadAll(response.Body)
		if errRes != nil {
			log4plus.Error("%s ReadAll Failed err: %s url: %s", funName, errRes.Error(), reqUri)
			return errRes, false, names, balances
		}
		type ResponseAmt struct {
			Exists   bool     `json:"exists"`   //是否存在指定的铭文
			Names    []string `json:"names"`    //名称
			Balances []int64  `json:"balances"` //数量
		}
		var checkAmtResponse ResponseAmt
		if err := json.Unmarshal(body, &checkAmtResponse); err != nil {
			log4plus.Error("%s ReadAll Failed err: %s url: %s", funName, err.Error(), reqUri)
			return err, false, names, balances
		}
		if checkAmtResponse.Exists {
			for index, name := range checkAmtResponse.Names {
				names = append(names, name)
				balances = append(balances, checkAmtResponse.Balances[index])
				log4plus.Info("%s Exists：%t Name：%s Balance=%d", funName, checkAmtResponse.Exists, name, checkAmtResponse.Balances[index])
			}
		}
		return nil, checkAmtResponse.Exists, names, balances
	}
	log4plus.Error("%s StatusCode is %d", funName, response.StatusCode)
	errString := fmt.Sprintf("%s StatusCode is %d", funName, response.StatusCode)
	return errors.New(errString), false, names, balances
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
