package main

import (
	log "code.google.com/p/log4go"
	"fmt"
	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
	"net/http"
)

func runWebServer(jeego *Jeego) {
	go func() {
		log.Info("Starting web server on port %d", jeego.config.WebServerPort)

		m := martini.Classic()
		m.Use(render.Renderer())

		// API: nodes list
		m.Get("/api/nodes", func(r render.Render) {
			r.JSON(200, map[string]interface{}{"nodes": jeego.database.nodes})
		})

		addr := fmt.Sprintf(":%d", jeego.config.WebServerPort)
		http.ListenAndServe(addr, m)
	}()
}
