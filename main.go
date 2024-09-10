package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/nGPU/common"
	log4plus "github.com/nGPU/common/log4go"
	"github.com/nGPU/discordBot/configure"
	"github.com/nGPU/discordBot/process"
	"github.com/nGPU/discordBot/web"
)

const (
	DiscordVersion = "2.0.0"
)

type Flags struct {
	Help    bool
	Version bool
}

func (f *Flags) Init() {
	flag.BoolVar(&f.Help, "h", false, "help")
	flag.BoolVar(&f.Version, "v", false, "show version")
}

func (f *Flags) Check() (needReturn bool) {
	flag.Parse()
	if f.Help {
		flag.Usage()
		needReturn = true
	} else if f.Version {
		verString := configure.SingtonConfigure().Application.Comment + " Version: " + DiscordVersion + "\r\n"
		fmt.Println(verString)
		needReturn = true
	}
	return needReturn
}

var flags *Flags = &Flags{}

func init() {
	flags.Init()
}

func getExeName() string {
	ret := ""
	ex, err := os.Executable()
	if err == nil {
		ret = filepath.Base(ex)
	}
	return ret
}

func setLog() {
	logJson := "log.json"
	set := false
	if bExist := common.PathExist(logJson); bExist {
		if err := log4plus.SetupLogWithConf(logJson); err == nil {
			set = true
		}
	}
	if !set {
		fileWriter := log4plus.NewFileWriter()
		exeName := getExeName()
		fileWriter.SetPathPattern("./log/" + exeName + "-%Y%M%D.log")
		log4plus.Register(fileWriter)
		log4plus.SetLevel(log4plus.DEBUG)
	}
}

func main() {

	//write crash dump file
	defer func() {
		if r := recover(); r != nil {
			dumpFile := fmt.Sprintf("crashDump_%s_%s.log", time.Now().Format("20060102_15_04_05"), fmt.Sprintf("%06d", time.Now().Nanosecond()/1e3))
			f, err := os.Create(dumpFile)
			if err != nil {
				log4plus.Error("Failed to create crash dump file err=[%s]", err.Error())
				return
			}
			defer f.Close()
			f.Write(debug.Stack())
			log4plus.Info("Crash dump saved to", dumpFile)
		}
	}()

	needReturn := flags.Check()
	if needReturn {
		return
	}
	setLog()
	defer log4plus.Close()
	if strings.Trim(configure.SingtonConfigure().Token.Discord.Token, " ") == "" {
		log4plus.Error("<<<<---->>>> Discord Token is empty")
		return
	}
	log4plus.Info("Discord Token is [%s]", configure.SingtonConfigure().Token.Discord.Token)
	web.SingtonWeb()
	process.SingtonProcess()
	for {
		time.Sleep(time.Duration(10) * time.Second)
	}
}
