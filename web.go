package main

import (
	"github.com/codegangsta/martini"
	"log"
)

func runWebServer(config *Config, nodes map[byte]INode) {
	m := martini.Classic()

	m.Get("/", func() string {
		result := ""

		for _, node := range nodes {
			result += node.textData()
		}

		return result
	})

	log.Printf("Starting web server runloop")
	m.Run()
}
