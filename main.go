package main

import (
	log "code.google.com/p/log4go"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	LOG_PERIOD  = 5 // in minutes
	LOG_HISTORY = 2 // in days
)

// Jeego
type Jeego struct {
	config   *Config
	database *Database
}

func newJeego() *Jeego {
	// load config
	config, err := loadConfig()
	if err != nil {
		panic(log.Critical(err))
	}

	// setup logging
	setupLogging(config.LogLevel, config.LogFile)

	log.Debug("Jeego config: %+v", config)

	// load database
	database, err := loadDatabase(config.DatabasePath)
	if err != nil {
		panic(log.Critical(err))
	}

	log.Info("Jeego database loaded with %d nodes: %v", len(database.nodes), config.DatabasePath)

	// debug
	for _, node := range database.nodes {
		node.logDebug(node.textData())
	}

	return &Jeego{
		config:   config,
		database: database,
	}
}

// Borrowed from InfluxDB
func setupLogging(loggingLevel, logFile string) {
	level := log.WARNING
	switch loggingLevel {
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

	if logFile == "stdout" {
		flw := log.NewConsoleLogWriter()
		log.AddFilter("stdout", level, flw)

	} else {
		logFileDir := filepath.Dir(logFile)
		os.MkdirAll(logFileDir, 0744)

		flw := log.NewFileLogWriter(logFile, false)
		log.AddFilter("file", level, flw)

		flw.SetFormat("[%D %T] [%L] %M")
		flw.SetRotate(true)
		flw.SetRotateSize(0)
		flw.SetRotateLines(0)
		flw.SetRotateDaily(true)
	}

	log.Info("Logging to file: %s", logFile)
}

func main() {
	// init Jeego
	jeego := newJeego()

	log.Info("Jeego - Target OS/Arch: %s %s", runtime.GOOS, runtime.GOARCH)
	log.Info("Built with Go Version: %s", runtime.Version())

	// save nodes values to database every 5mn
	jeego.database.startLogsTicker(time.Minute * LOG_PERIOD, time.Hour * 24 * LOG_HISTORY)

	// @todo Save nodes values to database every day

	// start web server
	runWebServer(jeego)

	// start handler
	handlerChan := runRf12demoHandler(jeego)

	// serial reader
	sr := newSerialReader(jeego.config.SerialPort, jeego.config.SerialBaud)

	log.Info("Reading on serial port: %+v", jeego.config.SerialPort)

	// loop forever
	for {
		// read a line and trim it
		line := strings.Trim(sr.readLine(), " \n\r")
		if line != "" {
			log.Debug("Received: %s", line)

			// send line to handler
			handlerChan <- line
		}
	}
}
