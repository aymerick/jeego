package main

import (
	"github.com/tarm/goserial"
	"io"
	"log"
	"runtime"
)

const (
	PORT_BAUD = 57600
	LF_CHAR   = 10
)

type SerialReader struct {
	io.ReadWriteCloser
}

func NewSerialReader() *SerialReader {
	defaultPorts := map[string]string {
		"darwin": "/dev/tty.usbserial-A1014IM4",
		"linux":  "/dev/ttyUSB0",
	}

	// @todo Settingfy port
	port := defaultPorts[runtime.GOOS]
	if port == "" {
		log.Panic("Unsupported plateform")
	}

	conf := &serial.Config{Name: port, Baud: PORT_BAUD}
	ser, err := serial.OpenPort(conf)
	if err != nil {
		log.Panic(err)
	}

	return &SerialReader{ser}
}

func (ser *SerialReader) readLine() string {
	result := make([]byte, 0)
	lastRead := make([]byte, 1)

	// read byte by byte until the Line Feed character
	for lastRead[0] != LF_CHAR {
		n, err := ser.Read(lastRead)
		if (err != nil) || (n != 1) {
			log.Panic(err)
		}

		result = append(result, lastRead[0])
	}

	return string(result)
}
