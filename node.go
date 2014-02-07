package main

import (
	log "code.google.com/p/log4go"
	"fmt"
	"math"
	"strconv"
	"time"
)

// node kinds
const (
	JEE_ROOM_NODE = iota + 1
	JEE_THL_NODE
	TINY_TEMP_NODE
)

// sensors kinds
const (
	TEMP_SENSOR   = 1 << iota
	HUMI_SENSOR   = 1 << iota
	LIGHT_SENSOR  = 1 << iota
	MOTION_SENSOR = 1 << iota
	LOWBAT_SENSOR = 1 << iota
	VCC_SENSOR    = 1 << iota
)

type Sensor uint

var SensorsForKind map[int]Sensor

// Node
type Node struct {
	Id          int       `json:"id"`
	Kind        int       `json:"kind"`
	UpdatedAt   time.Time `json:"updated_at"`
	Name        string    `json:"name"`
	DomoticzIdx string    `json:"domoticz_idx,omitempty"`

	// sensors
	Temperature float64   `json:"temperature,omitempty"`
	Humidity    uint8     `json:"humidity,omitempty"`
	Light       uint8     `json:"light,omitempty"`
	Motion      bool      `json:"motion,omitempty"`
	LowBattery  bool      `json:"low_battery,omitempty"`
	Vcc         int       `json:"vcc,omitempty"`
}

func init() {
	SensorsForKind = map[int]Sensor{
		JEE_ROOM_NODE:  TEMP_SENSOR + HUMI_SENSOR + LIGHT_SENSOR + LOWBAT_SENSOR + MOTION_SENSOR,
		JEE_THL_NODE:   TEMP_SENSOR + HUMI_SENSOR + LIGHT_SENSOR + LOWBAT_SENSOR,
		TINY_TEMP_NODE: TEMP_SENSOR + VCC_SENSOR,
	}
}

// log formatted debug message
func (node *Node) logDebug(msg string) {
	nodeName := node.Name
	if nodeName == "" {
		nodeName = "Unnamed"
	}

	log.Debug("[node %d][%s] %s", node.Id, nodeName, msg)
}

func (node *Node) handleData(data []byte) {
	switch node.Kind {
	case JEE_ROOM_NODE:
		node.handleJeeRoomNodeData(data)
	case JEE_THL_NODE:
		node.handleJeeThlNodeData(data)
	case TINY_TEMP_NODE:
		node.handleTinyTempNodeData(data)
	default:
		log.Error(fmt.Sprintf("Unsupported node kind: %d;", node.Kind))
	}
}

func (node *Node) haveSensor(sensor Sensor) bool {
	return ((SensorsForKind[node.Kind] & sensor) != 0)
}

//
// Example of data bytes decoding: 156 149 213 0
//
//             156 => 1 0 0 1 1 1 0 0
//             149 => 1 0 0 1 0 1 0 1
//             213 => 1 1 0 1 0 1 0 1
//               0 => 0 0 0 0 0 0 0 0
//
//             light: 1 0 0 1 1 1 0 0 => 156 * 100 / 255 = 61
//             moved:               1 => true
//          humidity: 1 0 0 1 0 1 0   => 74
//       temperature: 1 1 0 1 0 1 0 1 => 213 / 10 = 21.3
//                                0 0
//       low battery:           0     => false
//        <not used>: 0 0 0 0 0
//
// References:
//   - http://jeelabs.org/2011/06/09/rf12-packet-format-and-design/
//   - http://jeelabs.org/2011/01/14/nodes-addresses-and-interference/
//   - http://jeelabs.org/2010/12/07/binary-packet-decoding/
//   - http://jeelabs.org/2010/12/08/binary-packet-decoding-–-part-2/
//   - http://jeelabs.org/2013/09/05/decoding-bit-fields/
//   - http://jeelabs.org/2013/09/06/decoding-bit-fields-part-2/
//
func (node *Node) handleJeeRoomNodeData(data []byte) {
	if len(data) == 4 {
		var temperature int = ((256 * (int(data[3]) & 3)) + int(data[2]))
		if temperature > 512 {
			// negative value
			temperature = temperature - 1024
		}

		node.Light = uint8((int(data[0]) * 100) / 255)
		node.Motion = ((data[1] & 1) == 1)
		node.Humidity = data[1] >> 1
		node.Temperature = math.Ceil(float64(temperature)) / 10
		node.LowBattery = (((data[3] >> 2) & 1) == 1)
	}
}

//
// Example of data bytes decoding: 184 100 210 0
//
//             184 => 1 0 1 1 1 0 0 0
//             100 => 0 1 1 0 0 1 0 0
//             210 => 1 1 0 1 0 0 1 0
//               0 => 0 0 0 0 0 0 0 0
//
//             light: 1 0 1 1 1 0 0 0 => 184 * 100 / 255 = 72
//       low battery:               0 => true
//          humidity: 0 1 1 0 0 1 0   => 50
//       temperature: 1 1 0 1 0 0 1 0
//                                0 0 => 210 / 10 = 21.0
//        <not used>: 0 0 0 0 0 0
//
func (node *Node) handleJeeThlNodeData(data []byte) {
	if len(data) == 4 {
		var temperature int = ((256 * (int(data[3]) & 3)) + int(data[2]))
		if temperature > 512 {
			// negative value
			temperature = temperature - 1024
		}

		node.Light = uint8((int(data[0]) * 100) / 255)
		node.LowBattery = ((data[1] & 1) == 1)
		node.Humidity = data[1] >> 1
		node.Temperature = math.Ceil(float64(temperature)) / 10
	}
}

//
// Example of data bytes decoding: 92 44 17
//
//              92 => 0 1 0 1 1 1 0 0
//              44 => 0 0 1 0 1 1 0 0
//              17 => 0 0 0 1 0 0 0 1
//
//               vcc: 0 1 0 1 1 1 0 0
//                            1 1 0 0 => 3164 mv
//       temperature: 0 0 1 0
//                        0 1 0 0 0 1 => 274 / 10 = 27.4
//        <not used>: 0 0
//
func (node *Node) handleTinyTempNodeData(data []byte) {
	if len(data) == 3 {
		var vcc int = (int(data[1]&0x0F) << 8) + int(data[0])
		var temperature int = (int(data[1]&0xF0) >> 4) + (int(data[2]&0x3F) << 4)
		if temperature > 512 {
			// negative value
			temperature = temperature - 1024
		}

		node.Vcc = vcc
		node.Temperature = float64(temperature) / 10
	}
}

// Text to display for debugging
func (node *Node) textData() string {
	result := ""

	if node.haveSensor(TEMP_SENSOR) {
		result += "Temperature: " + strconv.FormatFloat(float64(node.Temperature), 'f', 1, 64)
	}

	if node.haveSensor(HUMI_SENSOR) {
		if result != "" {
			result += " | "
		}

		result += "Humidity: " + strconv.Itoa(int(node.Humidity))
	}

	if node.haveSensor(LIGHT_SENSOR) {
		if result != "" {
			result += " | "
		}

		result += "Light: " + strconv.Itoa(int(node.Light))
	}

	if node.haveSensor(MOTION_SENSOR) {
		if result != "" {
			result += " | "
		}

		result += "Motion: " + strconv.FormatBool(node.Motion)
	}

	if node.haveSensor(LOWBAT_SENSOR) {
		if result != "" {
			result += " | "
		}

		result += "LowBattery: " + strconv.FormatBool(node.LowBattery)
	}

	if node.haveSensor(VCC_SENSOR) {
		if result != "" {
			result += " | "
		}

		result += "Vcc: " + strconv.Itoa(node.Vcc)
	}

	return result
}

// Values to insert in database
func (node *Node) dbValues() []*DBValue {
	result := make([]*DBValue, 0)

	if node.haveSensor(TEMP_SENSOR) {
		result = append(result, &DBValue{name: "temperature", value: node.Temperature})
	}

	if node.haveSensor(HUMI_SENSOR) {
		result = append(result, &DBValue{name: "humidity", value: node.Humidity})
	}

	if node.haveSensor(LIGHT_SENSOR) {
		result = append(result, &DBValue{name: "light", value: node.Light})
	}

	if node.haveSensor(MOTION_SENSOR) {
		result = append(result, &DBValue{name: "motion", value: node.Motion})
	}

	if node.haveSensor(LOWBAT_SENSOR) {
		result = append(result, &DBValue{name: "lowbat", value: node.LowBattery})
	}

	if node.haveSensor(VCC_SENSOR) {
		result = append(result, &DBValue{name: "vcc", value: node.Vcc})
	}

	return result
}

// Query parameters part to send to domoticz
func (node *Node) domoticzParams(hardwareId string) string {
	result := ""

	isPushable := (node.DomoticzIdx != "") || (hardwareId != "")
	haveSensor := node.haveSensor(TEMP_SENSOR) || node.haveSensor(HUMI_SENSOR)

	if isPushable && haveSensor {
		if node.DomoticzIdx != "" {
			result += fmt.Sprintf("idx=%s&nvalue=0&svalue=", node.DomoticzIdx)
		} else {
			hid := hardwareId
			did := DOMOTICZ_DEVICE_ID_BASE + node.Id
			dunit := 1 // ??
			dsubtype := 1

			// pTypeTEMP_HUM 0x50 (temperature)
			dtype := 80

			if node.haveSensor(TEMP_SENSOR) && node.haveSensor(HUMI_SENSOR) {
				// pTypeTEMP_HUM 0x52 (temperature+humidity)
				dtype = 82
			} else if node.haveSensor(HUMI_SENSOR) {
				// pTypeTEMP_HUM 0x51 (humidity)
				dtype = 81
			}

			result += fmt.Sprintf("hid=%s&did=%d&dunit=%d&dtype=%d&dsubtype=%d&nvalue=0&svalue=", hid, did, dunit, dtype, dsubtype)
		}

		if node.haveSensor(TEMP_SENSOR) {
			result += fmt.Sprintf("%.1f;", node.Temperature)
		}

		if node.haveSensor(HUMI_SENSOR) {
			result += fmt.Sprintf("%d;", node.Humidity)
		}

		result += "0"
	}

	return result
}
