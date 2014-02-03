package main

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
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

// Database
type Database struct {
	driver     *sql.DB
	nodes      []*Node
	logsTicker *time.Ticker
}

// Database value
type DBValue struct {
	name  string
	value interface{}
}

// Setup a new database connection and load nodes
func loadDatabase() (*Database, error) {
	// open
	sqlDriver, err := sql.Open("sqlite3", "./jeego.db")
	if err != nil {
		return nil, err
	}

	db := Database{driver: sqlDriver}

	// create tables if necessary
	db.createTables()

	// load nodes
	db.loadNodes()

	return &db, nil
}

// Create tables
func (db *Database) createTables() {
	schemas := [2]string{NODES_SCHEMA, LOGS_SCHEMA}

	for _, schema := range schemas {
		_, err := db.driver.Exec(schema)
		if err != nil {
			log.Panic(errors.New(fmt.Sprintf("Failed to create SQL table %q: %s", err, schema)))
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
		log.Fatal(err)
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
			node.Vcc = int(vcc.Int64)
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
	_, err := db.driver.Exec("INSERT INTO nodes(id, kind) VALUES(?, ?)", id, kind)
	if err != nil {
		log.Fatal(err)
	}

	return node
}

// Update node
func (db *Database) updateNode(node *Node) {
	dbValues := node.dbValues()
	if len(dbValues) > 0 {
		node.UpdatedAt = time.Now().UTC()

		args := make([]interface{}, 0)

		query := "UPDATE nodes SET updated_at = ?"
		args = append(args, node.UpdatedAt.Unix())

		for _, dbValue := range dbValues {
			query += fmt.Sprintf(", %s = ?", dbValue.name)
			args = append(args, dbValue.value)
		}

		query += " WHERE id = ?"
		args = append(args, node.Id)

		_, err := db.driver.Exec(query, args...)
		if err != nil {
			log.Panic(err)
		}
	}
}

// Insert log for given node
func (db *Database) insertLog(node *Node) {
	dbValues := node.dbValues()
	if len(dbValues) > 0 {
		args := make([]interface{}, 0)

		query := "INSERT INTO logs(node_id, at"
		args = append(args, node.Id)
		args = append(args, time.Now().UTC().Unix())

		for _, dbValue := range dbValues {
			query += fmt.Sprintf(", %s", dbValue.name)
			args = append(args, dbValue.value)
		}

		query += ") VALUES(?, ?"
		for i := 0; i < len(dbValues); i++ {
			query += ", ?"
		}
		query += ")"

		// log.Printf("[insertLog] %s with %v", query, args)

		_, err := db.driver.Exec(query, args...)
		if err != nil {
			log.Panic(err)
		}
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

			db.trimLogs(history);
		}
	}()
}

// Delete old logs
func (db *Database) trimLogs(history time.Duration) {
	trimTo := time.Now().UTC().Add(-history).Unix()

	query := "DELETE FROM logs WHERE (at < ?)"

	// log.Printf("[trimLogs] %s with %v", query, trimTo)

	_, err := db.driver.Exec(query, trimTo)
	if err != nil {
		log.Panic(err)
	}
}
