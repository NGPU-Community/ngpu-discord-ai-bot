package process

import (
	"time"

	log4plus "github.com/nGPU/common/log4go"
	"github.com/nGPU/discordBot/aiModule"
	"github.com/nGPU/discordBot/configure"
	"github.com/nGPU/discordBot/db"
	"github.com/nGPU/discordBot/implementation"
	"github.com/nGPU/discordBot/plugins"
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
	implementation.SingtonStableDiffusion()
	// implementation.SingtonFengshui()
	aiModule.SingtonAnthropic()
}

func SingtonProcess() *Process {
	if gProcess == nil {
		gProcess = &Process{
			outChannel: make(chan bool),
		}
		configure.SingtonConfigure()
		db.SingtonUserDB()
		db.SingtonPaymentDB()
		db.SingtonAPITasksDB()

		StartPoll()
		plugins.SingtonDiscord()
	}
	return gProcess
}
