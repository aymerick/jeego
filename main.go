package main

import (
	"log"
	"os"
	"runtime"
	"strings"
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
		log.Panic(err)
	}

	log.Printf("Jeego config: %+v", config)

	// load database
	database, err := loadDatabase()
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Jeego database loaded")

	return &Jeego{
		config: config,
		database: database,
	}
}

func main() {
	log.SetOutput(os.Stderr)

	log.Printf("Jeego - Target OS/Arch: %s %s", runtime.GOOS, runtime.GOARCH)
	log.Printf("Built with Go Version: %s", runtime.Version())

	jeego := newJeego()

	// @todo Save nodes values to database every 5mn
	// @todo Save nodes values to database every day

	// start handler
	handlerChan := runRf12demoHandler(jeego)

	// serial reader
	sr := newSerialReader(jeego.config.SerialPort, jeego.config.SerialBaud)

	log.Printf("Reading on serial port: %+v", jeego.config.SerialPort)

	// loop forever
	for {
		// read a line and trim it
		line := strings.Trim(sr.readLine(), " \n\r")
		if line != "" {
			if jeego.config.Debug {
				log.Printf("Received: %s", line)
			}

			// send line to handler
			handlerChan <- line
		}
	}
}
