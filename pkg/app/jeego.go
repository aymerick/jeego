package app

import (
	"os"
	"path/filepath"
	"time"

	log "code.google.com/p/log4go"

	"github.com/aymerick/jeego/pkg/config"
	"github.com/aymerick/jeego/pkg/domoticz"
	"github.com/aymerick/jeego/pkg/ws_hub"
)

const (
	LOG_PERIOD  = 5 // in minutes
	LOG_HISTORY = 2 // in days
)

// Jeego
type Jeego struct {
	Config   *config.Config
	Database *Database
	wsHub    *ws_hub.WsHub
	Domoticz *domoticz.Domoticz
}

func NewJeego() *Jeego {
	return &Jeego{}
}

func (jeego *Jeego) LoadConfig() {
	var err error

	// load config
	jeego.Config, err = config.Load()
	if err != nil {
		panic(log.Critical(err))
	}
}

func (jeego *Jeego) DumpConfig() {
	log.Debug("Jeego config: %+v", jeego.Config)
}

func (jeego *Jeego) SetupLogging() {
	level := log.WARNING
	switch jeego.Config.LogLevel {
	case "debug":
		level = log.DEBUG
	case "info":
		level = log.INFO
	case "error":
		level = log.ERROR
	}

	for _, filter := range log.Global {
		filter.Level = level
	}

	if jeego.Config.LogFile == "stdout" {
		flw := log.NewConsoleLogWriter()
		log.AddFilter("stdout", level, flw)

	} else {
		logFileDir := filepath.Dir(jeego.Config.LogFile)
		os.MkdirAll(logFileDir, 0744)

		flw := log.NewFileLogWriter(jeego.Config.LogFile, false)
		log.AddFilter("file", level, flw)

		flw.SetFormat("[%D %T] [%L] %M")
		flw.SetRotate(true)
		flw.SetRotateSize(0)
		flw.SetRotateLines(0)
		flw.SetRotateDaily(true)
	}

	log.Info("Logging to file: %s", jeego.Config.LogFile)
}

func (jeego *Jeego) SetupDatabase() {
	var err error

	// load database
	jeego.Database, err = LoadDatabase(jeego.Config.DatabasePath)
	if err != nil {
		panic(log.Critical(err))
	}

	log.Info("Jeego database loaded with %d nodes: %v", len(jeego.Database.nodes), jeego.Config.DatabasePath)

	// debug
	for _, node := range jeego.Database.nodes {
		node.LogDebug(node.TextData())
	}
}

// Add a log entry every 5 minutes
func (jeego *Jeego) RunNodeLogsTicker() {
	logsTicker := time.NewTicker(time.Minute * LOG_PERIOD)

	// do it right now
	jeego.Database.insertNodeLogs()

	go func() {
		for _ = range logsTicker.C {
			// insert logs
			jeego.Database.insertNodeLogs()

			// @todo send to websocket clients
			for _, node := range jeego.Database.nodes {
				jeego.wsHub.SendMsg([]byte(node.TextData()))
			}

			// trim old logs
			jeego.Database.trimNodeLogs(time.Hour * 24 * LOG_HISTORY)
		}
	}()
}

// Setup domoticz remote
func (jeego *Jeego) SetupDomoticz() {
	if jeego.Config.DomoticzHost != "" {
		jeego.Domoticz = &domoticz.Domoticz{
			Host:       jeego.Config.DomoticzHost,
			Port:       jeego.Config.DomoticzPort,
			HardwareId: jeego.Config.DomoticzHardwareId,
		}
	}
}

// Start Websocket Hub
func (jeego *Jeego) StartWsHub() {
	jeego.wsHub = ws_hub.Run()
}

// Start Web Server
func (jeego *Jeego) StartWebServer() {
	RunWebServer(jeego)
}

// Start RF12demo handler
func (jeego *Jeego) StartRf12demo() chan string {
	return RunRf12demo(jeego)
}
