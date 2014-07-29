package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	log "code.google.com/p/log4go"
)

// Reference: https://github.com/mitchellh/packer

// Serial port:
//   - Jeelink on Mac: /dev/tty.usbserial-A1014IM4
//   - Jeelink on Raspberry: /dev/ttyUSB0
//   - Jeenode on Raspberry FTDI: /dev/ttyAMA0 (cf. http://jeelabs.org/2012/09/20/serial-hookup-jeenode-to-raspberry-pi/)
const defaultConfig = `
{
	"serial_port": "/dev/ttyUSB0",
	"serial_baud": 57600,
	"domoticz_port": 8080,
	"log_level": "info",
	"log_file": "stdout",
	"database_path": "./jeego.db",
	"web_server_port": 3000
}
`

// Jeego configuration
type Config struct {
	SerialPort         string `json:"serial_port"`
	SerialBaud         int    `json:"serial_baud"`
	DomoticzHost       string `json:"domoticz_host"`
	DomoticzPort       int    `json:"domoticz_port"`
	DomoticzHardwareId string `json:"domoticz_hardware_id"`
	LogLevel           string `json:"log_level"`
	LogFile            string `json:"log_file"`
	DatabasePath       string `json:"database_path"`
	WebServerPort      int    `json:"web_server_port"`
	WebAppPath         string `json:"web_app_path"`
	Rf12demoLogFile    string `json:"rf12demo_log_file"`
}

// Load config from conf file
func Load() (*Config, error) {
	var config Config
	if err := decodeConfig(bytes.NewBufferString(defaultConfig), &config); err != nil {
		return nil, err
	}

	mustExist := true
	configFilePath := os.Getenv("JEEGO_CONFIG")
	if configFilePath == "" {
		var err error
		configFilePath, err = configFile()
		mustExist = false

		if err != nil {
			log.Error("Error detecting default config file path: %s", err)
		}
	}

	if configFilePath == "" {
		return &config, nil
	}

	log.Debug("Attempting to open config file: %s", configFilePath)
	f, err := os.Open(configFilePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}

		if mustExist {
			return nil, err
		}

		log.Debug("File doesn't exist, but doesn't need to. Ignoring.")
		return &config, nil
	}
	defer f.Close()

	if err := decodeConfig(f, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func decodeConfig(r io.Reader, c *Config) error {
	decoder := json.NewDecoder(r)
	return decoder.Decode(c)
}

func configFile() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, ".jeego.json"), nil
}

func configDir() (string, error) {
	// First prefer the HOME environmental variable
	if home := os.Getenv("HOME"); home != "" {
		log.Debug("Detected home directory from env var: %s", home)
		return home, nil
	}

	// If that fails, try the shell
	var stdout bytes.Buffer
	cmd := exec.Command("sh", "-c", "eval echo ~$USER")
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", err
	}

	result := strings.TrimSpace(stdout.String())
	if result == "" {
		return "", errors.New("blank output")
	}

	return result, nil
}
