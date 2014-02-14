package main

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)


func Test_InsertNodeQuery(t *testing.T) {
	query, args := insertNodeQuery(2, JEENODE_THLM_NODE)

	assert.Equal(t, query, "INSERT INTO nodes(id, kind) VALUES(?, ?)")
	assert.True(t, reflect.DeepEqual(args, []interface{}{ 2, JEENODE_THLM_NODE }))
}

func Test_UpdateNodeQuery(t *testing.T) {
	node := &Node{
		Id: 2,
		Kind: JEENODE_THLM_NODE,
		UpdatedAt: time.Now().UTC(),
		Temperature: float64(21.3),
		Humidity: uint8(74),
		Light: uint8(61),
		Motion: true,
		LowBattery: false,
	}

	expected_query := "UPDATE nodes SET updated_at = ?, temperature = ?, humidity = ?, light = ?, motion = ?, lowbat = ?, vcc = NULL WHERE id = ?"
	expected_args  := []interface{}{ node.UpdatedAt.Unix(), float64(21.3), uint8(74), uint8(61), true, false, node.Id }

	query, args := updateNodeQuery(node)

	assert.Equal(t, query, expected_query)
	assert.True(t, reflect.DeepEqual(args, expected_args))
}

func Test_InsertLogQuery(t *testing.T) {
	node := &Node{
		Id: 2,
		Kind: JEENODE_THLM_NODE,
		UpdatedAt: time.Now().UTC(),
		Temperature: float64(21.3),
		Humidity: uint8(74),
		Light: uint8(61),
		Motion: true,
		LowBattery: false,
	}

	expected_query := "INSERT INTO logs(node_id, at, temperature, humidity, light, motion, lowbat) VALUES(?, ?, ?, ?, ?, ?, ?)"
	expected_args  := []interface{}{ node.Id, node.UpdatedAt.Unix(), float64(21.3), uint8(74), uint8(61), true, false }

	query, args := insertLogQuery(node)

	assert.Equal(t, query, expected_query)
	assert.True(t, reflect.DeepEqual(args, expected_args))
}
