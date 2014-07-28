package app

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const TEST_DB_PATH = "./jeego-test.db"

func newDatabase() *Database {
	db, _ := LoadDatabase(TEST_DB_PATH)
	return db
}

func destroyDatabase(db *Database) {
	if db != nil {
		db.close()
	}
	os.Remove(TEST_DB_PATH)
}

func Test_InsertNode(t *testing.T) {
	destroyDatabase(nil)
	db := newDatabase()
	defer destroyDatabase(db)

	assert.Equal(t, len(db.nodes), 0)

	db.InsertNode(2, JEENODE_THLM_NODE)

	db2 := newDatabase()
	defer destroyDatabase(db2)

	assert.Equal(t, len(db2.nodes), 1)
}

func Test_UpdateNodeQuery(t *testing.T) {
	node := &Node{
		Id:          2,
		Kind:        JEENODE_THLM_NODE,
		UpdatedAt:   time.Now().UTC(),
		LastSeenAt:  time.Now().UTC(),
		Name:        "test",
		Temperature: float64(21.3),
		Humidity:    uint8(74),
		Light:       uint8(61),
		Motion:      true,
		LowBattery:  false,
	}

	expected_query := "UPDATE nodes SET updated_at = ?, last_seen_at = ?, name = ?, temperature = ?, humidity = ?, light = ?, motion = ?, lowbat = ?, vcc = NULL WHERE id = ?"
	expected_args := []interface{}{node.UpdatedAt.Unix(), node.LastSeenAt.Unix(), node.Name, float64(21.3), uint8(74), uint8(61), true, false, node.Id}

	dbQuery := updateNodeQuery(node)

	assert.Equal(t, dbQuery.query, expected_query)
	assert.True(t, reflect.DeepEqual(dbQuery.args, expected_args))
}

func Test_UpdateNode(t *testing.T) {
	destroyDatabase(nil)
	db := newDatabase()
	defer destroyDatabase(db)

	assert.Equal(t, len(db.nodes), 0)

	node2 := db.InsertNode(2, JEENODE_THLM_NODE)
	node3 := db.InsertNode(3, TINYTX_TH_NODE)

	node2.Temperature = float64(21.3)
	node2.Humidity = uint8(74)
	node2.Light = uint8(61)
	node2.Motion = true
	node2.LowBattery = false

	db.UpdateNode(node2)

	node3.Temperature = float64(19.4)
	node3.Vcc = 3096

	db.UpdateNode(node3)

	db2 := newDatabase()
	defer destroyDatabase(db2)

	assert.Equal(t, len(db2.nodes), 2)

	node2 = db2.NodeForId(2)
	assert.Equal(t, node2.Temperature, float64(21.3))
	assert.Equal(t, node2.Humidity, uint8(74))
	assert.Equal(t, node2.Light, uint8(61))
	assert.Equal(t, node2.Motion, true)
	assert.Equal(t, node2.LowBattery, false)

	node3 = db2.NodeForId(3)
	assert.Equal(t, node3.Temperature, float64(19.4))
	assert.Equal(t, node3.Vcc, uint(3096))
}

func Test_InsertLogQuery(t *testing.T) {
	node := &Node{
		Id:          2,
		Kind:        JEENODE_THLM_NODE,
		UpdatedAt:   time.Now().UTC(),
		LastSeenAt:  time.Now().UTC(),
		Temperature: float64(21.3),
		Humidity:    uint8(74),
		Light:       uint8(61),
		Motion:      true,
		LowBattery:  false,
	}

	expected_query := "INSERT INTO logs(node_id, at, temperature, humidity, light, motion, lowbat) VALUES(?, ?, ?, ?, ?, ?, ?)"
	expected_args := []interface{}{node.Id, node.LastSeenAt.Unix(), float64(21.3), uint8(74), uint8(61), true, false}

	dbQuery := insertLogQuery(node)

	assert.Equal(t, dbQuery.query, expected_query)
	assert.True(t, reflect.DeepEqual(dbQuery.args, expected_args))
}

func Test_InsertLogs(t *testing.T) {
	destroyDatabase(nil)
	db := newDatabase()
	defer destroyDatabase(db)

	assert.Equal(t, len(db.nodes), 0)

	node2 := db.InsertNode(2, JEENODE_THLM_NODE)
	node3 := db.InsertNode(3, TINYTX_TH_NODE)

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
