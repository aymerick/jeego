package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	log "code.google.com/p/log4go"
	"github.com/bmizerany/pat"
)

// helper
func respondsWithError404(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	io.WriteString(w, "404: Not Found")
}

// helper
func respondsWithError400(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	io.WriteString(w, "400: Bad Request")
}

// helper
func respondsWithJSON(w http.ResponseWriter, data map[string]interface{}) {
	response, err := json.Marshal(data)
	if err != nil {
		panic(log.Critical(err))
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

func addAccessControlHeaders(w http.ResponseWriter, meth string) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", meth)
	w.Header().Set("Access-Control-Allow-Headers", "Origin, Accept, Content-Type")
}

// GET /api/nodes
func wrapHandlerNodes(jeego *Jeego) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		addAccessControlHeaders(w, "GET")

		respondsWithJSON(w, map[string]interface{}{"nodes": jeego.database.nodes})
	}
}

// GET /api/nodes/:id
func wrapHandlerNode(jeego *Jeego) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		addAccessControlHeaders(w, "GET")

		nodeId, err := strconv.Atoi(req.URL.Query().Get(":id"))
		if err != nil {
			respondsWithError400(w)
		} else {
			node := jeego.database.nodeForId(nodeId)
			if node != nil {
				respondsWithJSON(w, map[string]interface{}{"node": node})
			} else {
				respondsWithError404(w)
			}
		}
	}
}

func runWebServer(jeego *Jeego) {
	go func() {
		log.Info("Starting web server on port %d", jeego.config.WebServerPort)

		addr := fmt.Sprintf(":%d", jeego.config.WebServerPort)

		mux := pat.New()
		mux.Get("/api/nodes", wrapHandlerNodes(jeego))
		mux.Get("/api/nodes/:id", wrapHandlerNode(jeego))

		http.Handle("/", mux)
		err := http.ListenAndServe(addr, nil)
		if err != nil {
			panic(log.Critical(err))
		}
	}()
}
