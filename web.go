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

type NodeJSON struct {
	Node Node `json:"node"`
}

// helper
func respondsWithError(w http.ResponseWriter, status int, err error) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	io.WriteString(w, err.Error())
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

// OPTIONS ...
func wrapHandlerOptions(jeego *Jeego, meth string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		addAccessControlHeaders(w, meth)
	}
}

// GET /api/nodes
func wrapHandlerNodes(jeego *Jeego, meth string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		addAccessControlHeaders(w, meth)

		// @todo Handle pagination (cf. http://emberjs.com/guides/models/handling-metadata/)
		respondsWithJSON(w, map[string]interface{}{"nodes": jeego.database.nodes})
	}
}

// GET /api/nodes/:id
func wrapHandlerNode(jeego *Jeego, meth string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		addAccessControlHeaders(w, meth)

		nodeId, err := strconv.Atoi(req.URL.Query().Get(":id"))
		if err != nil {
			respondsWithError(w, http.StatusBadRequest, err)
		} else {
			node := jeego.database.nodeForId(nodeId)
			if node != nil {
				respondsWithJSON(w, map[string]interface{}{"node": node})
			} else {
				respondsWithError(w, http.StatusNotFound, fmt.Errorf("Node %d not found", nodeId))
			}
		}
	}
}

// PUT /api/nodes/:id
func wrapHandlerUpdateNode(jeego *Jeego, meth string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		addAccessControlHeaders(w, meth)

		// Parse the incoming kitten from the request body
		var nodeJSON NodeJSON
		err := json.NewDecoder(req.Body).Decode(&nodeJSON)
		if err != nil {
			log.Error(fmt.Sprintf("Failed to parse JSON: %v", err))
			respondsWithError(w, http.StatusBadRequest, err)
		} else {
			nodeId, err := strconv.Atoi(req.URL.Query().Get(":id"))
			if err != nil {
				log.Error("Failed to get node id")
				respondsWithError(w, http.StatusBadRequest, err)
			} else {
				node := jeego.database.nodeForId(nodeId)
				if node == nil {
					respondsWithError(w, http.StatusNotFound, fmt.Errorf("Node %d not found", nodeId))
				} else {
					// update node
					node.Name = nodeJSON.Node.Name

					jeego.database.updateNode(node)

					respondsWithJSON(w, map[string]interface{}{"node": node})
				}
			}
		}
	}
}

func runWebServer(jeego *Jeego) {
	go func() {
		log.Info("Starting web server on port %d", jeego.config.WebServerPort)

		addr := fmt.Sprintf(":%d", jeego.config.WebServerPort)

		mux := pat.New()

		nodesMeth := "OPTIONS, GET"
		mux.Options("/api/nodes", wrapHandlerOptions(jeego, nodesMeth))
		mux.Get("/api/nodes", wrapHandlerNodes(jeego, nodesMeth))

		nodeMeth := "OPTIONS, GET, PUT"
		mux.Options("/api/nodes/:id", wrapHandlerOptions(jeego, nodeMeth))
		mux.Get("/api/nodes/:id", wrapHandlerNode(jeego, nodeMeth))
		mux.Put("/api/nodes/:id", wrapHandlerUpdateNode(jeego, nodeMeth))

		http.Handle("/", mux)
		err := http.ListenAndServe(addr, nil)
		if err != nil {
			panic(log.Critical(err))
		}
	}()
}
