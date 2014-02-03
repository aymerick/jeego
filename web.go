package main

import (
	log "code.google.com/p/log4go"
	"github.com/codegangsta/martini"
)

func runWebServer(database *Database) {
	m := martini.Classic()

	m.Get("/", func() string {
		result := ""

		for _, node := range database.nodes {
			result += node.textData()
		}

		return result
	})

	log.Info("Starting web server runloop")
	m.Run()
}
