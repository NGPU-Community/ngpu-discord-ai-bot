package process

import (
	"time"

	"github.com/nGPU/bot/aiModule"
	"github.com/nGPU/bot/db"
	"github.com/nGPU/bot/implementation"
	"github.com/nGPU/bot/plugins"
	log4plus "github.com/nGPU/common/log4go"
)

type Process struct {
	outChannel chan bool
}

var gProcess *Process

func (p *Process) pollMonitor() {
	for {
		time.Sleep(time.Duration(5) * time.Second)
		log4plus.Info("****************************")
	}
}

func StopProcess() {
	if gProcess != nil {
		gProcess.outChannel <- true
	}
}

func StartPoll() {
	implementation.SingtonFaceFusion()
	implementation.SingtonSadTalker()
	implementation.SingtonBlip()
	implementation.SingtonChatTTS()
	implementation.SingtonLlm()
	implementation.SingtonTxt2Img()
	implementation.SingtonStableDiffusion()
	// implementation.SingtonFengshui()
	aiModule.SingtonAnthropic()
}

func SingtonProcess() *Process {
	if gProcess == nil {
		gProcess = &Process{
			outChannel: make(chan bool),
		}

		db.SingtonUserDB()
		db.SingtonPaymentDB()
		db.SingtonAPITasksDB()

		StartPoll()
		plugins.SingtonDiscord()
		plugins.SingtonTelegram()
	}
	return gProcess
}
