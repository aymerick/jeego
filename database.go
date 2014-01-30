package main

import (
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
    "log"
    "errors"
    "fmt"
    "time"
)

var schema = `
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

// Database
type Database struct {
    driver *sql.DB
    nodes  []*Node
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
    db.createTables();

    // load nodes
    db.loadNodes();

    return &db, nil
}

// Create tables
func (db *Database) createTables() {
    _, err := db.driver.Exec(schema)
    if err != nil {
        log.Panic(errors.New(fmt.Sprintf("Failed to create SQL tables %q: %s", err, schema)))
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
        // fetch node fields
        var (
            id           int
            kind         int
            updated_at   int64
            name         string
            domoticz_idx string
            temperature  float32
            humidity     uint8
            light        uint8
            motion       bool
            lowbat       bool
            vcc          int
        )

        rows.Scan(&id, &kind, &updated_at, &name, &domoticz_idx, &temperature, &humidity, &light, &motion, &lowbat, &vcc)

        // init node
        node := &Node{
            Id:          id,
            Kind:        kind,
            UpdatedAt:   time.Unix(updated_at, 0),
            Name:        name,
            DomoticzIdx: domoticz_idx,
            Temperature: temperature,
            Humidity:    humidity,
            Light:       light,
            Motion:      motion,
            LowBattery:  lowbat,
            Vcc:         vcc,
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
    node := &Node{ Id: id, Kind: kind }

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
