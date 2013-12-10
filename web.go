package main

import (
	"github.com/codegangsta/martini"
	"log"
)

func runWebServer(config *Config) {
	m := martini.Classic()

	m.Get("/", func() string {
		result := ""

		for _, node := range config.Nodes {
			result += node.textData()
		}

		return result
	})

	log.Printf("Starting web server runloop")
	m.Run()
}
