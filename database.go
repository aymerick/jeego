package main

import (
	log "code.google.com/p/log4go"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

const NODES_SCHEMA = `
CREATE TABLE IF NOT EXISTS nodes (
    id INTEGER NOT NULL PRIMARY KEY,
    kind INTEGER NOT NULL,
    updated_at INTEGER,
    name TEXT,
    domoticz_idx TEXT,
    temperature REAL,
    humidity INTEGER,
    light INTEGER,
    motion INTEGER,
    lowbat INTEGER,
    vcc INTEGER
);
`

const LOGS_SCHEMA = `
CREATE TABLE IF NOT EXISTS logs (
	node_id INTEGER NOT NULL,
	at INTEGER NOT NULL,
	temperature REAL,
	humidity INTEGER,
	light INTEGER,
	motion INTEGER,
	lowbat INTEGER,
	vcc INTEGER
);
`

var ColNameForSensor map[Sensor]string

// Database
type Database struct {
	driver      *sql.DB
	queryWriter chan *DatabaseQuery
	nodes       []*Node
	logsTicker  *time.Ticker
}

// Database Query
type DatabaseQuery struct {
	query string
	args  []interface{}
}

// Init
func init() {
	ColNameForSensor = map[Sensor]string{
		TEMP_SENSOR:   "temperature",
		HUMI_SENSOR:   "humidity",
		LIGHT_SENSOR:  "light",
		MOTION_SENSOR: "motion",
		LOWBAT_SENSOR: "lowbat",
		VCC_SENSOR:    "vcc",
	}
}

// Setup a new database connection and load nodes
func loadDatabase(databasePath string) (*Database, error) {
	// open
	sqlDriver, err := sql.Open("sqlite3", databasePath)
	if err != nil {
		return nil, err
	}

	db := Database{driver: sqlDriver}

	// create tables if necessary
	db.createTables()

	// load nodes
	db.loadNodes()

	// run query writer
	db.runQueryWriter()

	return &db, nil
}

// Start RF12demo handler
func (db *Database) runQueryWriter() {
	inputChan := make(chan *DatabaseQuery)

	go func() {
		var dbQuery *DatabaseQuery

		// loop forever
		for {
			dbQuery = <-inputChan

			// log.Debug("Exec DB write query: %s", dbQuery.query)

			_, err := db.driver.Exec(dbQuery.query, dbQuery.args...)
			if err != nil {
				panic(log.Critical(err))
			}
		}
	}()

	db.queryWriter = inputChan
}

func (db *Database) writeQuery(dbQuery *DatabaseQuery) {
	db.queryWriter <- dbQuery
}

func (db *Database) close() {
	db.driver.Close()
}

// Create tables
func (db *Database) createTables() {
	schemas := [2]string{NODES_SCHEMA, LOGS_SCHEMA}

	for _, schema := range schemas {
		_, err := db.driver.Exec(schema)
		if err != nil {
			panic(log.Critical("Failed to create SQL table %q: %s", err, schema))
		}
	}
}

// Load all nodes
func (db *Database) loadNodes() {
	// reset nodes
	db.nodes = make([]*Node, 0)

	// fetch nodes from db
	rows, err := db.driver.Query("SELECT * FROM nodes")
	if err != nil {
		panic(log.Critical(err))
	}
	defer rows.Close()

	for rows.Next() {
		var node *Node

		// fetch node fields
		var (
			id           int
			kind         int
			updated_at   int64
			name         sql.NullString
			domoticz_idx sql.NullString
			temperature  sql.NullFloat64
			humidity     sql.NullInt64
			light        sql.NullInt64
			motion       sql.NullBool
			lowbat       sql.NullBool
			vcc          sql.NullInt64
		)

		// @todo Use github.com/russross/meddler ?
		rows.Scan(&id, &kind, &updated_at, &name, &domoticz_idx, &temperature, &humidity, &light, &motion, &lowbat, &vcc)

		// init node
		node = &Node{
			Id:        id,
			Kind:      kind,
			UpdatedAt: time.Unix(updated_at, 0),
		}

		if name.Valid {
			node.Name = name.String
		}

		if domoticz_idx.Valid {
			node.DomoticzIdx = domoticz_idx.String
		}

		if temperature.Valid {
			node.Temperature = float64(temperature.Float64)
		}

		if humidity.Valid {
			node.Humidity = uint8(humidity.Int64)
		}

		if light.Valid {
			node.Light = uint8(light.Int64)
		}

		if motion.Valid {
			node.Motion = motion.Bool
		}

		if lowbat.Valid {
			node.LowBattery = lowbat.Bool
		}

		if vcc.Valid {
			node.Vcc = uint(vcc.Int64)
		}

		// add node to list
		db.nodes = append(db.nodes, node)
	}

	rows.Close()
}

// Get a node
func (db *Database) nodeForId(id int) *Node {
	for _, node := range db.nodes {
		if node.Id == id {
			return node
		}
	}

	return nil
}

// Insert a new node
func (db *Database) insertNode(id int, kind int) *Node {
	// init node
	node := &Node{Id: id, Kind: kind}

	// add node to list
	db.nodes = append(db.nodes, node)

	// persist in database
	db.writeQuery(&DatabaseQuery{
		query: "INSERT INTO nodes(id, kind) VALUES(?, ?)",
		args:  []interface{}{ id, kind },
	})

	return node
}

func updateNodeQuery(node *Node) *DatabaseQuery {
	args := make([]interface{}, 0)

	query := "UPDATE nodes SET updated_at = ?"
	args = append(args, node.UpdatedAt.Unix())

	// set sensors values
	for _, sensor := range node.sensors() {
		colName := ColNameForSensor[sensor]
		if colName != "" {
			value := node.sensorValue(sensor)

			query += fmt.Sprintf(", %s = ?", colName)
			args = append(args, value)
		}
	}

	// set NULL for absent sensors
	for _, sensor := range node.absentSensors() {
		colName := ColNameForSensor[sensor]
		if colName != "" {
			query += fmt.Sprintf(", %s = NULL", colName)
		}
	}

	query += " WHERE id = ?"
	args = append(args, node.Id)

	return &DatabaseQuery{ query: query, args: args }
}

// Update node
func (db *Database) updateNode(node *Node) {
	if len(node.sensors()) > 0 {
		node.UpdatedAt = time.Now().UTC()

		// persist in database
		db.writeQuery(updateNodeQuery(node))
	}
}

func insertLogQuery(node *Node) *DatabaseQuery {
	args := make([]interface{}, 0)

	query := "INSERT INTO logs(node_id, at"
	args = append(args, node.Id)
	args = append(args, time.Now().UTC().Unix())

	nbSensors := 0

	for _, sensor := range node.sensors() {
		colName := ColNameForSensor[sensor]
		if colName != "" {
			query += fmt.Sprintf(", %s", colName)
			args = append(args, node.sensorValue(sensor))

			nbSensors += 1
		}
	}

	query += ") VALUES(?, ?"
	for i := 0; i < nbSensors; i++ {
		query += ", ?"
	}
	query += ")"

	return &DatabaseQuery{ query: query, args: args }
}

// Insert log for given node
func (db *Database) insertLog(node *Node) {
	if len(node.sensors()) > 0 {
		// persist in database
		db.writeQuery(insertLogQuery(node))
	}
}

// Insert logs for all nodes
func (db *Database) insertLogs() {
	for _, node := range db.nodes {
		db.insertLog(node)
	}
}

// Add a log entry every 5 minutes
func (db *Database) startLogsTicker(period time.Duration, history time.Duration) {
	db.logsTicker = time.NewTicker(period)

	// do it right now
	db.insertLogs()

	go func() {
		for _ = range db.logsTicker.C {
			db.insertLogs()

			db.trimLogs(history)
		}
	}()
}

// Delete old logs
func (db *Database) trimLogs(history time.Duration) {
	// persist in database
	db.writeQuery(&DatabaseQuery{
		query: "DELETE FROM logs WHERE (at < ?)",
		args:  []interface{}{ time.Now().UTC().Add(-history).Unix() },
	})
}
