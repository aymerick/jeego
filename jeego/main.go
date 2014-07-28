package main

import (
	"runtime"
	"strings"

	"github.com/aymerick/jeego/pkg/app"
	"github.com/aymerick/jeego/pkg/rf12demo"
	"github.com/aymerick/jeego/pkg/serial_reader"

	log "code.google.com/p/log4go"
)

func main() {
	// init Jeego Server
	jeego := app.NewJeego()

	jeego.LoadConfig()
	jeego.SetupLogging()

	log.Info("Jeego - Target OS/Arch: %s %s", runtime.GOOS, runtime.GOARCH)
	log.Info("Built with Go Version: %s", runtime.Version())

	// debug
	jeego.DumpConfig()

	jeego.SetupDatabase()

	// save nodes values to database every 5mn
	jeego.RunNodeLogsTicker()

	// start websocket hub
	jeego.StartWsHub()

	// start web server
	jeego.StartWebServer()

	// start RF12 handler
	handlerChan := rf12demo.Run(jeego)

	// serial reader
	sr := serial_reader.New(jeego.Config.SerialPort, jeego.Config.SerialBaud)

	log.Info("Reading on serial port: %+v", jeego.Config.SerialPort)

	// loop forever
	for {
		// read a line and trim it
		line := strings.Trim(sr.ReadLine(), " \n\r")
		if line != "" {
			log.Debug("Received: %s", line)

			// send line to handler
			handlerChan <- line
		}
	}
}
