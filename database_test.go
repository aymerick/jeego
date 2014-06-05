package main

import (
	"github.com/stretchr/testify/assert"
	"os"
	"reflect"
	"testing"
	"time"
)

const TEST_DB_PATH = "./jeego-test.db"

func newDatabase() *Database {
	db, _ := loadDatabase(TEST_DB_PATH)
	return db
}

func destroyDatabase(db *Database) {
	if db != nil {
		db.close()
	}
	os.Remove(TEST_DB_PATH)
}


func Test_InsertNodeQuery(t *testing.T) {
	query, args := insertNodeQuery(2, JEENODE_THLM_NODE)

	assert.Equal(t, query, "INSERT INTO nodes(id, kind) VALUES(?, ?)")
	assert.True(t, reflect.DeepEqual(args, []interface{}{ 2, JEENODE_THLM_NODE }))
}

func Test_InsertNode(t *testing.T) {
	destroyDatabase(nil)
	db := newDatabase()
	defer destroyDatabase(db)

	assert.Equal(t, len(db.nodes), 0)

	db.insertNode(2, JEENODE_THLM_NODE)

	db2 := newDatabase()
	defer destroyDatabase(db2)

	assert.Equal(t, len(db2.nodes), 1)
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

func Test_UpdateNode(t *testing.T) {
	destroyDatabase(nil)
	db := newDatabase()
	defer destroyDatabase(db)

	assert.Equal(t, len(db.nodes), 0)

	node2 := db.insertNode(2, JEENODE_THLM_NODE)
	node3 := db.insertNode(3, TINYTX_TH_NODE)

	node2.Temperature = float64(21.3)
	node2.Humidity = uint8(74)
	node2.Light = uint8(61)
	node2.Motion = true
	node2.LowBattery = false

	db.updateNode(node2)

	node3.Temperature = float64(19.4)
	node3.Vcc = 3096

	db.updateNode(node3)

	db2 := newDatabase()
	defer destroyDatabase(db2)

	assert.Equal(t, len(db2.nodes), 2)

	node2 = db2.nodeForId(2)
	assert.Equal(t, node2.Temperature, float64(21.3))
	assert.Equal(t, node2.Humidity,    uint8(74))
	assert.Equal(t, node2.Light,       uint8(61))
	assert.Equal(t, node2.Motion,      true)
	assert.Equal(t, node2.LowBattery,  false)

	node3 = db2.nodeForId(3)
	assert.Equal(t, node3.Temperature, float64(19.4))
	assert.Equal(t, node3.Vcc,         uint(3096))
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

func Test_InsertLogs(t *testing.T) {
	destroyDatabase(nil)
	db := newDatabase()
	defer destroyDatabase(db)

	assert.Equal(t, len(db.nodes), 0)

	node2 := db.insertNode(2, JEENODE_THLM_NODE)
	node3 := db.insertNode(3, TINYTX_TH_NODE)

	node2.Temperature = float64(21.3)
	node2.Humidity = uint8(74)
	node2.Light = uint8(61)
	node2.Motion = true
	node2.LowBattery = false

	node3.Temperature = float64(19.4)
	node3.Vcc = 3096

	db.insertLogs()
	// @todo Check that logs are correctly inserted in database
}

func Test_TrimLogsQuery(t *testing.T) {
	history := time.Minute

	expected_query := "DELETE FROM logs WHERE (at < ?)"
	expected_args  := []interface{}{ time.Now().UTC().Add(-history).Unix() }

	query, args := trimLogsQuery(history)

	assert.Equal(t, query, expected_query)
	assert.True(t, reflect.DeepEqual(args, expected_args))
}