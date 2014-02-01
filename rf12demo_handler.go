package main

import (
	"errors"
	"log"
	"strconv"
	"strings"
)

// Start RF12demo handler
func runRf12demoHandler(jeego *Jeego) chan string {
	inputChan := make(chan string, 1)

	go func() {
		var line string

		// loop forever
		for {
			line = <-inputChan

			// parse node infos and data
			nodeId, nodeKind, data, err := parseLine(line)
			if err == nil {
				// @todo log raw data to file

				// get node
				node := jeego.database.nodeForId(nodeId)
				if node == nil {
					// insert new node in database
					node = jeego.database.insertNode(nodeId, nodeKind)

					// debug
					if jeego.config.Debug {
						pringDebugMsgForNode(node, "New node added to database")
					}
				}

				// handle data
				node.handleData(data)

				// debug
				if jeego.config.Debug {
					pringDebugMsgForNode(node, node.textData())
				}

				// update database
				jeego.database.updateNode(node)

				// debug
				if jeego.config.Debug {
					pringDebugMsgForNode(node, "Node updated into database")
				}

				// push to domoticz
				go pushToDomoticz(jeego.config, node)
			}
		}
	}()

	return inputChan
}

// print formatted debug message
func pringDebugMsgForNode(node *Node, msg string) {
	log.Printf("%s <node %d> %s", node.Name, node.Id, msg)
}

// Parse a line received from central node
//
// Example of line generated by RF12Demo sketch that received a packet from a jeeRoomNode:
//
//       OK 2 3 156 149 213 0
//          ^ ^ -------------
//     header |      ^
//            |  data bytes
//            |
//         node kind
//
// header:
//
//      0   0   0   0   0   0   1   0
//      ^   ^   ^   -----------------
//     CTL DST ACK         ^
//                    node id => 2
//
// node kind:
//
//      0   0   0   0   0   0   1   1
//      ^   -------------------------
// reserved            ^
//               node kind => 3
func parseLine(line string) (nodeId int, nodeKind int, data []byte, err error) {
	// split line
	dataStrArray := strings.Split(line, " ")

	// parse status
	if (len(dataStrArray) > 3) && (dataStrArray[0] == "OK") {
		// parse node id
		nodeId = int(byteFromString(dataStrArray[1]) & 0x1f)

		// parse node infos
		nodeInfosByte := byteFromString(dataStrArray[2])

		// check reserved field
		if (nodeInfosByte & 0x80) != 0 {
			log.Printf("Received payload with reserved field set to 1")
		} else {
			// parse node kind
			nodeKind = int(nodeInfosByte & 0x7f)

			// parse data
			data = make([]byte, len(dataStrArray)-3)

			for index, dataStr := range dataStrArray {
				if index > 2 {
					data[index-3] = byteFromString(dataStr)
				}
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
