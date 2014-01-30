package main

import (
	"github.com/codegangsta/martini"
	"log"
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

	log.Printf("Starting web server runloop")
	m.Run()
}
