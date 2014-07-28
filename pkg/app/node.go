package app

import (
	"fmt"
	"math"
	"reflect"
	"time"

	log "code.google.com/p/log4go"
)

// node kinds
const (
	JEENODE_THLM_NODE = iota + 1 // Jeenode: Temperature Humidity Light Motion
	JEENODE_THL_NODE             // Jeenode: Temperature Humidity Light
	TINYTX_T_NODE                //  TinyTX: Temperature
	TINYTX_TH_NODE               //  TinyTX: Temperature Humidity
	TINYTX_TL_NODE               //  TinyTX: Temperature Light
)

// sensors kinds
const (
	TEMP_SENSOR   = iota // Temperature
	HUMI_SENSOR          // Humidity
	LIGHT_SENSOR         // Light
	MOTION_SENSOR        // Motion
	LOWBAT_SENSOR        // Low Battery
	VCC_SENSOR           // Supply voltage
)

// base to compute Device ID
const DOMOTICZ_DEVICE_ID_BASE = 2000

type Sensor uint

var AllSensors []Sensor
var SensorsForNodeKind map[int][]Sensor
var BitsNbForSensor map[Sensor]int

// Node
type Node struct {
	Id          int       `json:"id"`
	Kind        int       `json:"kind"`
	UpdatedAt   time.Time `json:"updated_at"`
	LastSeenAt  time.Time `json:"last_seen_at"`
	Name        string    `json:"name"`
	DomoticzIdx string    `json:"domoticz_idx"`

	// sensors
	Temperature float64 `json:"temperature"`
	Humidity    uint8   `json:"humidity"`
	Light       uint8   `json:"light"`
	Motion      bool    `json:"motion"`
	LowBattery  bool    `json:"low_battery"`
	Vcc         uint    `json:"vcc"`
}

func init() {
	AllSensors = []Sensor{TEMP_SENSOR, HUMI_SENSOR, LIGHT_SENSOR, MOTION_SENSOR, LOWBAT_SENSOR, VCC_SENSOR}

	SensorsForNodeKind = map[int][]Sensor{
		JEENODE_THLM_NODE: {TEMP_SENSOR, HUMI_SENSOR, LIGHT_SENSOR, MOTION_SENSOR, LOWBAT_SENSOR},
		JEENODE_THL_NODE:  {TEMP_SENSOR, HUMI_SENSOR, LIGHT_SENSOR, LOWBAT_SENSOR},
		TINYTX_T_NODE:     {TEMP_SENSOR, VCC_SENSOR},
		TINYTX_TH_NODE:    {TEMP_SENSOR, HUMI_SENSOR, VCC_SENSOR},
		TINYTX_TL_NODE:    {TEMP_SENSOR, LIGHT_SENSOR, VCC_SENSOR},
	}

	BitsNbForSensor = map[Sensor]int{
		TEMP_SENSOR:   10, // [10 bits] Temperature: -512..+512 (tenths)
		HUMI_SENSOR:   7,  //  [7 bits] Humidity: 0..100
		LIGHT_SENSOR:  8,  //  [8 bits] Light: 0..255
		MOTION_SENSOR: 1,  //   [1 bit] Motion: 0..1
		LOWBAT_SENSOR: 1,  //   [1 bit] Low Battery: 0..1
		VCC_SENSOR:    12, // [12 bits] Supply voltage: 0..4095 mV
	}
}

// log formatted debug message
func (node *Node) LogDebug(msg string) {
	nodeName := node.Name
	if nodeName == "" {
		nodeName = "Unnamed"
	}

	log.Debug("[node %d][%s] %s", node.Id, nodeName, msg)
}

// log formatted warn message
func (node *Node) LogWarn(msg string) {
	nodeName := node.Name
	if nodeName == "" {
		nodeName = "Unnamed"
	}

	log.Warn("[node %d][%s] %s", node.Id, nodeName, msg)
}

// return all sensors
func (node *Node) sensors() []Sensor {
	return SensorsForNodeKind[node.Kind]
}

// return all absent sensors
func (node *Node) absentSensors() []Sensor {
	result := make([]Sensor, 0)

	for _, sensor := range AllSensors {
		if !node.haveSensor(sensor) {
			result = append(result, sensor)
		}
	}

	return result
}

// check if node have given sensor
func (node *Node) haveSensor(sensor Sensor) bool {
	for _, nodeSensor := range node.sensors() {
		if nodeSensor == sensor {
			return true
		}
	}

	return false
}

// reset sensors values
func (node *Node) ResetSensors() {
	node.Temperature = float64(0)
	node.Humidity = uint8(0)
	node.Light = uint8(0)
	node.Motion = false
	node.LowBattery = false
	node.Vcc = 0
}

// handle incoming node data
func (node *Node) HandleData(data []byte) {
	if node.sensors() != nil {
		expectedLength := node.expectedDataLength()
		if len(data) == expectedLength {
			sensorsData := node.parseData(data)

			for sensor, value := range sensorsData {
				node.setSensorRawValue(sensor, value)
			}
		} else {
			log.Error(fmt.Sprintf("Unexpected data length: %v / Expected: %d", data, expectedLength))
		}
	} else {
		log.Error(fmt.Sprintf("Unsupported node kind: %d", node.Kind))
	}
}

// returns expected node data length
func (node *Node) expectedDataLength() int {
	bitsNb := 0

	for _, sensor := range node.sensors() {
		bitsNb += BitsNbForSensor[sensor]
	}

	result := bitsNb / 8

	if (bitsNb % 8) != 0 {
		result += 1
	}

	return result
}

// parse incoming node data
func (node *Node) parseData(data []byte) map[Sensor]uint64 {
	result := make(map[Sensor]uint64)

	curByte := 0
	curBytePos := 0

	value := uint64(0)
	sensorBitsNb := 0
	totalBitsShift := 0
	bytesNeeded := 0

	for _, sensor := range AllSensors {
		if node.haveSensor(sensor) {
			value = 0

			sensorBitsNb = BitsNbForSensor[sensor]
			if sensorBitsNb > 64 {
				panic(log.Critical("Wow, sensor %d needs %d bits, really ?", sensor, sensorBitsNb))
			}

			totalBitsShift = curBytePos + sensorBitsNb

			bytesNeeded = totalBitsShift / 8
			if (totalBitsShift % 8) != 0 {
				bytesNeeded += 1
			}

			for i := 0; i < bytesNeeded; i++ {
				value += uint64(data[curByte+i]) << uint(8*i)
			}

			value = (value >> uint(curBytePos)) & ((1 << uint(sensorBitsNb)) - 1)

			result[sensor] = value

			curByte += (bytesNeeded - 1)
			curBytePos = totalBitsShift % 8
		}
	}

	return result
}

// set given sensor value
func (node *Node) setSensorRawValue(sensor Sensor, value uint64) {
	switch sensor {
	case TEMP_SENSOR:
		node.Temperature = node.computeTemperatureValue(value)

	case HUMI_SENSOR:
		node.Humidity = node.computeHumidityValue(value)

	case LIGHT_SENSOR:
		node.Light = node.computeLightValue(value)

	case MOTION_SENSOR:
		node.Motion = node.computeMotionValue(value)

	case LOWBAT_SENSOR:
		node.LowBattery = node.computeLowbatValue(value)

	case VCC_SENSOR:
		node.Vcc = node.computeVccValue(value)

	default:
		panic(log.Critical("Unknown sensor: %d", sensor))
	}
}

// compite temperature value from raw data
func (node *Node) computeTemperatureValue(value uint64) float64 {
	result := int64(value)

	if result > 512 {
		// negative value
		result -= 1024
	}

	return math.Ceil(float64(result)) / 10
}

// compite humidity value from raw data
func (node *Node) computeHumidityValue(value uint64) uint8 {
	return uint8(value)
}

// compite light value from raw data
func (node *Node) computeLightValue(value uint64) uint8 {
	return uint8((value * 100) / 255)
}

// compite motion value from raw data
func (node *Node) computeMotionValue(value uint64) bool {
	return (value != 0)
}

// compite low battery value from raw data
func (node *Node) computeLowbatValue(value uint64) bool {
	return (value != 0)
}

// compite vcc value from raw data
func (node *Node) computeVccValue(value uint64) uint {
	return uint(value)
}

// get sensor value
func (node *Node) sensorValue(sensor Sensor) interface{} {
	var result interface{}

	switch sensor {
	case TEMP_SENSOR:
		result = node.Temperature

	case HUMI_SENSOR:
		result = node.Humidity

	case LIGHT_SENSOR:
		result = node.Light

	case MOTION_SENSOR:
		result = node.Motion

	case LOWBAT_SENSOR:
		result = node.LowBattery

	case VCC_SENSOR:
		result = node.Vcc

	default:
		panic(log.Critical("Unknown sensor: %d", sensor))
	}

	return result
}

// text to display for debugging
func (node *Node) TextData() string {
	result := ""

	for _, sensor := range node.sensors() {
		if result != "" {
			result += " | "
		}

		result += fmt.Sprintf("%s: %v", ColNameForSensor[sensor], node.sensorValue(sensor))
	}

	return result
}

// cf. http://stackoverflow.com/a/17323212
func (node *Node) toJsonifableMap() map[string]interface{} {
	result := make(map[string]interface{})

	result[node.jsonFieldName("Id")] = node.Id
	result[node.jsonFieldName("Kind")] = node.Kind
	result[node.jsonFieldName("UpdatedAt")] = node.UpdatedAt
	result[node.jsonFieldName("LastSeenAt")] = node.LastSeenAt
	result[node.jsonFieldName("Name")] = node.Name

	if node.DomoticzIdx != "" {
		result[node.jsonFieldName("DomoticzIdx")] = node.DomoticzIdx
	}

	for _, sensor := range AllSensors {
		if node.haveSensor(sensor) {
			switch sensor {
			case TEMP_SENSOR:
				result[node.jsonFieldName("Temperature")] = node.Temperature

			case HUMI_SENSOR:
				result[node.jsonFieldName("Humidity")] = node.Humidity

			case LIGHT_SENSOR:
				result[node.jsonFieldName("Light")] = node.Light

			case MOTION_SENSOR:
				result[node.jsonFieldName("Motion")] = node.Motion

			case LOWBAT_SENSOR:
				result[node.jsonFieldName("LowBattery")] = node.LowBattery

			case VCC_SENSOR:
				result[node.jsonFieldName("Vcc")] = node.Vcc

			default:
				panic(log.Critical("Unknown sensor: %d", sensor))
			}
		}
	}

	// this emberjs convention for async relationships retrieval
	// @todo Move that to web.go
	result["links"] = map[string]interface{}{"logs": fmt.Sprintf("/api/nodes/%d/logs", node.Id)}

	return result
}

// get json field name for given struct field name
func (node *Node) jsonFieldName(fieldName string) string {
	rt := reflect.TypeOf(*node)
	field, _ := rt.FieldByName(fieldName)
	return field.Tag.Get("json")
}

// query parameters part to send to domoticz
func (node *Node) DomoticzParams(hardwareId string) string {
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

func (node *Node) temperaturesSerie(nodeLogs []*NodeLog) [][]string {
	result := make([][]string, len(nodeLogs))

	for index, nodeLog := range nodeLogs {
		serieData := []string{nodeLog.At.Format(time.RFC3339), fmt.Sprintf("%v", nodeLog.Temperature)}
		result[index] = serieData
	}

	return result
}
