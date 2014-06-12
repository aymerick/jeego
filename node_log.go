package main

import (
	"reflect"
	"time"

	log "code.google.com/p/log4go"
)

type NodeLog struct {
	Id          int       `json:"id"`
	NodeId      int       `json:"node_id"`
	At          time.Time `json:"at"`
	Temperature float64   `json:"temperature"`
	Humidity    uint8     `json:"humidity"`
	Light       uint8     `json:"light"`
	Motion      bool      `json:"motion"`
	LowBattery  bool      `json:"low_battery"`
	Vcc         uint      `json:"vcc"`
}

// cf. http://stackoverflow.com/a/17323212
func (nodeLog *NodeLog) toJsonifableMap(node *Node) map[string]interface{} {
	result := make(map[string]interface{})

	result[nodeLog.jsonFieldName("Id")] = nodeLog.Id
	result[nodeLog.jsonFieldName("NodeId")] = nodeLog.NodeId
	result[nodeLog.jsonFieldName("At")] = nodeLog.At.UTC()

	for _, sensor := range AllSensors {
		if node.haveSensor(sensor) {
			switch sensor {
			case TEMP_SENSOR:
				result[nodeLog.jsonFieldName("Temperature")] = nodeLog.Temperature

			case HUMI_SENSOR:
				result[nodeLog.jsonFieldName("Humidity")] = nodeLog.Humidity

			case LIGHT_SENSOR:
				result[nodeLog.jsonFieldName("Light")] = nodeLog.Light

			case MOTION_SENSOR:
				result[nodeLog.jsonFieldName("Motion")] = nodeLog.Motion

			case LOWBAT_SENSOR:
				result[nodeLog.jsonFieldName("LowBattery")] = nodeLog.LowBattery

			case VCC_SENSOR:
				result[nodeLog.jsonFieldName("Vcc")] = nodeLog.Vcc

			default:
				panic(log.Critical("Unknown sensor: %d", sensor))
			}
		}
	}

	return result
}

// get json field name for given struct field name
func (nodeLog *NodeLog) jsonFieldName(fieldName string) string {
	rt := reflect.TypeOf(*nodeLog)
	field, _ := rt.FieldByName(fieldName)
	return field.Tag.Get("json")
}
