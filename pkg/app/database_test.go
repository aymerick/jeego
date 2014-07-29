package app

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TempFilename() string {
	randBytes := make([]byte, 16)
	rand.Read(randBytes)
	return filepath.Join(os.TempDir(), "jeego-test"+hex.EncodeToString(randBytes)+".db")
}

func newTestDatabase(t *testing.T, dbFilename string) *Database {
	db, err := LoadDatabase(dbFilename)
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}

	// be sure write queries are synchronous
	db.SetSync(true)

	return db
}

func destroyTestDatabase(db *Database) {
	db.close()
	os.Remove(db.filePath)
}

func Test_InsertNode(t *testing.T) {
	dbFilename := TempFilename()

	db := newTestDatabase(t, dbFilename)
	defer destroyTestDatabase(db)

	assert.Equal(t, len(db.nodes), 0)

	db.InsertNode(2, JEENODE_THLM_NODE)

	assert.Equal(t, len(db.nodes), 1)

	// reopen database
	db.close()
	db2 := newTestDatabase(t, dbFilename)
	defer destroyTestDatabase(db2)

	assert.Equal(t, len(db.nodes), 1)
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
	dbFilename := TempFilename()

	db := newTestDatabase(t, dbFilename)
	defer destroyTestDatabase(db)

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

	// reopen database
	db.close()
	db2 := newTestDatabase(t, dbFilename)
	defer destroyTestDatabase(db2)

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

func Test_InsertNodeLogs(t *testing.T) {
	dbFilename := TempFilename()

	db := newTestDatabase(t, dbFilename)
	defer destroyTestDatabase(db)

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

	db.insertNodeLogs()

	// reopen database
	db.close()
	db2 := newTestDatabase(t, dbFilename)
	defer destroyTestDatabase(db2)

	nodeLogs := db2.nodeLogs(node2)
	assert.Equal(t, len(nodeLogs), 1)

	nodeLog := nodeLogs[0]
	assert.Equal(t, nodeLog.Temperature, float64(21.3))
	assert.Equal(t, nodeLog.Humidity, uint8(74))
	assert.Equal(t, nodeLog.Light, uint8(61))
	assert.Equal(t, nodeLog.Motion, true)
	assert.Equal(t, nodeLog.LowBattery, false)

	nodeLogs = db2.nodeLogs(node3)
	assert.Equal(t, len(nodeLogs), 1)

	nodeLog = nodeLogs[0]
	assert.Equal(t, nodeLog.Temperature, float64(19.4))
	assert.Equal(t, nodeLog.Vcc, 3096)
}
