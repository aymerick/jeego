package main

import (
	"runtime"
	"strings"

	log "code.google.com/p/log4go"
)

func main() {
	// init Jeego
	jeego := newJeego()

	jeego.loadConfig()
	jeego.setupLogging()

	log.Info("Jeego - Target OS/Arch: %s %s", runtime.GOOS, runtime.GOARCH)
	log.Info("Built with Go Version: %s", runtime.Version())

	// debug
	jeego.dumpConfig()

	jeego.setupDatabase()

	// save nodes values to database every 5mn
	jeego.runNodeLogsTicker()

	// start websocket hub
	jeego.wsHub = runWsHub()

	// start web server
	runWebServer(jeego)

	// start RF12 handler
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
