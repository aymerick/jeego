package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
)

func main() {
	log.SetOutput(os.Stderr)

	log.Printf("Jeego - Target OS/Arch: %s %s", runtime.GOOS, runtime.GOARCH)
	log.Printf("Built with Go Version: %s", runtime.Version())

	// load config
	config, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: \n\n%s\n", err)
		return
	}

	log.Printf("Jeego config: %+v", config)

	// instanciate nodes
	nodes := instanciateNodes(config)

	// serial reader
	sr := NewSerialReader(config.SerialPort, config.SerialBaud)

	// loop forever
	for {
		// read a line and trim it
		line := strings.Trim(sr.readLine(), " \n\r")
		if line != "" {
			log.Printf("Received: %s", line)

			// parse node id and data
			nodeId, data, err := parseLine(line)
			if err == nil {
				// get node
				node := nodes[nodeId]
				if node != nil {
					// handle data
					node.handleData(data)

					// debug
					node.dumpData()
				} else {
					// unknown node
					log.Printf("Ignoring unknown node: %v", nodeId)
				}
			}
		}
	}
}

// Instanciate nodes map from config file
func instanciateNodes(config *Config) map[byte]INode {
	result := make(map[byte]INode)

	for _, nodeConfig := range config.Nodes {
		var node INode = nil

		switch nodeConfig.Kind {
		case "roomNode":
			node = &RoomNode{Node: Node{nodeConfig.Id, nodeConfig.Name}}
		case "thlNode":
			node = &ThlNode{Node: Node{nodeConfig.Id, nodeConfig.Name}}
		default:
			log.Panic(fmt.Printf("Unsupported node kind: %s", nodeConfig.Kind))
		}

		result[nodeConfig.Id] = node
	}

	return result
}

// Parse a line received from central node
//
// Example of line generated by RF12Demo sketch that received a packet from a standard roomNode:
//
//       OK 2 156 149 213 0
//          ^ -------------
//     header      ^
//             data bytes
//
//
// header:
//
//      0   0   0   0   0   0   1   0
//      ^   ^   ^   -----------------
//     CTL DST ACK          ^
//                       node id => 2
func parseLine(line string) (nodeId byte, data []byte, err error) {
	// split line
	dataStrArray := strings.Split(line, " ")

	// parse status
	if dataStrArray[0] == "OK" {
		// parse node id
		nodeId = byteFromString(dataStrArray[1]) & 0x1f

		// parse data
		data = make([]byte, len(dataStrArray)-2)

		for index, dataStr := range dataStrArray {
			if index > 1 {
				data[index-2] = byteFromString(dataStr)
			}
		}
	} else {
		err = errors.New("Garbage received")
	}

	return
}

// helper
func byteFromString(val string) byte {
	i, err := strconv.ParseUint(val, 10, 8)
	if err != nil {
		log.Panic(err)
	}

	return byte(i)
}
