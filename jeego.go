package main

import (
	"os"
	"path/filepath"
	"time"

	log "code.google.com/p/log4go"
)

const (
	LOG_PERIOD  = 5 // in minutes
	LOG_HISTORY = 2 // in days
)

// Jeego
type Jeego struct {
	config   *Config
	database *Database
	wsHub    *WsHub
}

func newJeego() *Jeego {
	return &Jeego{}
}

func (jeego *Jeego) loadConfig() {
	var err error

	// load config
	jeego.config, err = loadConfig()
	if err != nil {
		panic(log.Critical(err))
	}
}

func (jeego *Jeego) dumpConfig() {
	log.Debug("Jeego config: %+v", jeego.config)
}

func (jeego *Jeego) setupLogging() {
	level := log.WARNING
	switch jeego.config.LogLevel {
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

	if jeego.config.LogFile == "stdout" {
		flw := log.NewConsoleLogWriter()
		log.AddFilter("stdout", level, flw)

	} else {
		logFileDir := filepath.Dir(jeego.config.LogFile)
		os.MkdirAll(logFileDir, 0744)

		flw := log.NewFileLogWriter(jeego.config.LogFile, false)
		log.AddFilter("file", level, flw)

		flw.SetFormat("[%D %T] [%L] %M")
		flw.SetRotate(true)
		flw.SetRotateSize(0)
		flw.SetRotateLines(0)
		flw.SetRotateDaily(true)
	}

	log.Info("Logging to file: %s", jeego.config.LogFile)
}

func (jeego *Jeego) setupDatabase() {
	var err error

	// load database
	jeego.database, err = loadDatabase(jeego.config.DatabasePath)
	if err != nil {
		panic(log.Critical(err))
	}

	log.Info("Jeego database loaded with %d nodes: %v", len(jeego.database.nodes), jeego.config.DatabasePath)

	// debug
	for _, node := range jeego.database.nodes {
		node.logDebug(node.textData())
	}
}

// Add a log entry every 5 minutes
func (jeego *Jeego) runNodeLogsTicker() {
	logsTicker := time.NewTicker(time.Minute * LOG_PERIOD)

	// do it right now
	jeego.database.insertNodeLogs()

	go func() {
		for _ = range logsTicker.C {
			// insert logs
			jeego.database.insertNodeLogs()

			// @todo send to websocket clients
			for _, node := range jeego.database.nodes {
				jeego.wsHub.sendMsg([]byte(node.textData()))
			}

			// trim old logs
			jeego.database.trimNodeLogs(time.Hour * 24 * LOG_HISTORY)
		}
	}()
}
