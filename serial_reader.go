package main

import (
	log "code.google.com/p/log4go"
	"github.com/chimera/rs232"
	"io"
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
	options := rs232.Options{ BitRate: uint32(baud), DataBits: 8, StopBits: 1 }
	ser, err := rs232.Open(port, options)

	if err != nil {
		panic(log.Critical(err))
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
			panic(log.Critical(err))
		}

		result = append(result, lastRead[0])
	}

	return string(result)
}
