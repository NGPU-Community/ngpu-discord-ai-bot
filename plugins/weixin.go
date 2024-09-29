package plugins

import (
	"crypto/x509"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/mdp/qrterminal/v3"
	log4plus "github.com/nGPU/common/log4go"
	"github.com/wechaty/go-wechaty/wechaty"
	wp "github.com/wechaty/go-wechaty/wechaty-puppet"
	"github.com/wechaty/go-wechaty/wechaty-puppet/filebox"
	"github.com/wechaty/go-wechaty/wechaty-puppet/schemas"
	"github.com/wechaty/go-wechaty/wechaty/user"
)

type Weixin struct {
	roots   *x509.CertPool
	rootPEM []byte
	bot     *wechaty.Wechaty
}

var gWeixin *Weixin

func onMessage(ctx *wechaty.Context, message *user.Message) {
	funName := "onMessage"

	log4plus.Info("%s message=[%+v]", funName, *message)
	if message.Self() {
		log4plus.Error("%s Message discarded because its outgoing", funName)
		return
	}

	if message.Age() > 2*60*time.Second {
		log4plus.Error("%s Message discarded because its TOO OLD(than 2 minutes)", funName)
		return
	}

	if message.Type() != schemas.MessageTypeText || message.Text() != "#ding" {
		log4plus.Error("%s Message discarded because it does not match #ding", funName)
		return
	}

	// 1. reply text 'dong'
	_, err := message.Say("dong")
	if err != nil {
		log4plus.Error("%s err=[%s]", funName, err)
		return
	}
	log4plus.Info("%s REPLY with text: dong", funName)

	// 2. reply image(qrcode image)
	fileBox := filebox.FromUrl("https://wechaty.github.io/wechaty/images/bot-qr-code.png")
	_, err = message.Say(fileBox)
	if err != nil {
		log4plus.Error("%s err=[%s]", funName, err)
		return
	}
	log4plus.Info("%s REPLY with image: %s", funName, fileBox)

	// 3. reply url link
	urlLink := user.NewUrlLink(&schemas.UrlLinkPayload{
		Description:  "Go Wechaty is a Conversational SDK for Chatbot Makers Written in Go",
		ThumbnailUrl: "https://wechaty.js.org/img/icon.png",
		Title:        "wechaty/go-wechaty",
		Url:          "https://github.com/wechaty/go-wechaty",
	})
	_, err = message.Say(urlLink)
	if err != nil {
		log4plus.Error("%s err=[%s]", funName, err)
		return
	}
	log4plus.Info("%s REPLY with urlLink: %s", funName, urlLink)
}

func onScan(ctx *wechaty.Context, qrCode string, status schemas.ScanStatus, data string) {
	if status == schemas.ScanStatusWaiting || status == schemas.ScanStatusTimeout {
		qrterminal.GenerateHalfBlock(qrCode, qrterminal.L, os.Stdout)

		qrcodeImageUrl := fmt.Sprintf("https://wechaty.js.org/qrcode/%s", url.QueryEscape(qrCode))
		fmt.Printf("onScan: %s - %s\n", status, qrcodeImageUrl)
		return
	}
	fmt.Printf("onScan: %s\n", status)
}

func (w *Weixin) Version() string {
	return "Weixin Version 1.0.0"
}

func (w *Weixin) Init() {
	funName := "Init"
	w.bot = wechaty.NewWechaty(wechaty.WithPuppetOption(wp.Option{
		Token: "",
	}))

	w.bot.OnScan(onScan).OnLogin(func(ctx *wechaty.Context, user *user.ContactSelf) {
		log4plus.Info("%s User %s logined", funName, user.Name())

	}).OnMessage(onMessage).OnLogout(func(ctx *wechaty.Context, user *user.ContactSelf, reason string) {
		log4plus.Info("%s User %s logouted: %s", funName, reason)

	})
	w.bot.DaemonStart()
}

func SingtonWeixin() *Weixin {
	if nil == gWeixin {
		gWeixin = &Weixin{}

		go gWeixin.Init()
	}
	return gWeixin
}
