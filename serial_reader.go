package main

import (
	"github.com/tarm/goserial"
	"io"
	"log"
)

const (
	LF_CHAR = 10
)

// Serial port reader
type SerialReader struct {
	io.ReadWriteCloser
}

// Instanciate a serial port reader
func newSerialReader(port string, baud int) *SerialReader {
	conf := &serial.Config{Name: port, Baud: baud}
	ser, err := serial.OpenPort(conf)
	if err != nil {
		log.Panic(err)
	}

	return &SerialReader{ser}
}

// Read a line from serial port
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
