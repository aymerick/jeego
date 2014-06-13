package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
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
// @todo Handle pagination (cf. http://emberjs.com/guides/models/handling-metadata/)
func wrapHandlerNodes(jeego *Jeego, meth string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		addAccessControlHeaders(w, meth)

		nodes := jeego.database.nodes
		result := make([]interface{}, len(nodes))

		for index, node := range nodes {
			result[index] = node.toJsonifableMap()
		}

		respondsWithJSON(w, map[string]interface{}{"nodes": result})
	}
}

// GET /api/nodes/:id
func wrapHandlerNode(jeego *Jeego, meth string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		addAccessControlHeaders(w, meth)

		// parse node id
		nodeId, err := strconv.Atoi(req.URL.Query().Get(":id"))
		if err != nil {
			respondsWithError(w, http.StatusBadRequest, err)
		} else {
			// get node
			node := jeego.database.nodeForId(nodeId)
			if node != nil {
				respondsWithJSON(w, map[string]interface{}{"node": node.toJsonifableMap()})
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

		// parse JSON
		var nodeJSON NodeJSON
		err := json.NewDecoder(req.Body).Decode(&nodeJSON)
		if err != nil {
			log.Error(fmt.Sprintf("Failed to parse JSON: %v", err))
			respondsWithError(w, http.StatusBadRequest, err)
		} else {
			// parse node id
			nodeId, err := strconv.Atoi(req.URL.Query().Get(":id"))
			if err != nil {
				log.Error("Failed to get node id")
				respondsWithError(w, http.StatusBadRequest, err)
			} else {
				// get node
				node := jeego.database.nodeForId(nodeId)
				if node == nil {
					respondsWithError(w, http.StatusNotFound, fmt.Errorf("Node %d not found", nodeId))
				} else {
					// update node
					node.Name = nodeJSON.Node.Name

					jeego.database.updateNode(node)

					respondsWithJSON(w, map[string]interface{}{"node": node.toJsonifableMap()})
				}
			}
		}
	}
}

// GET /api/nodes/:id/temperatures
func wrapHandlerNodeTemperatures(jeego *Jeego, meth string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		addAccessControlHeaders(w, meth)

		// parse node id
		nodeId, err := strconv.Atoi(req.URL.Query().Get(":id"))
		if err != nil {
			respondsWithError(w, http.StatusBadRequest, err)
		} else {
			// get node
			node := jeego.database.nodeForId(nodeId)
			if node != nil {
				// get logs
				serie := node.temperaturesSerie(jeego.database.nodeLogs(node))

				respondsWithJSON(w, map[string]interface{}{"temperatures": serie})
			} else {
				respondsWithError(w, http.StatusNotFound, fmt.Errorf("Node %d not found", nodeId))
			}
		}
	}
}

// GET /api/nodes/:id/logs
func wrapHandlerNodeLogs(jeego *Jeego, meth string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		addAccessControlHeaders(w, meth)

		// parse node id
		nodeId, err := strconv.Atoi(req.URL.Query().Get(":id"))
		if err != nil {
			respondsWithError(w, http.StatusBadRequest, err)
		} else {
			// get node
			node := jeego.database.nodeForId(nodeId)
			if node != nil {
				// get logs
				nodeLogs := jeego.database.nodeLogs(node)
				result := make([]interface{}, len(nodeLogs))

				for index, nodeLog := range nodeLogs {
					result[index] = nodeLog.toJsonifableMap(node)
				}

				respondsWithJSON(w, map[string]interface{}{"logs": result})
			} else {
				respondsWithError(w, http.StatusNotFound, fmt.Errorf("Node %d not found", nodeId))
			}
		}
	}
}

func setupWebApp(jeego *Jeego) string {
	rootDirPath := os.Getenv("HOME")
	if rootDirPath == "" {
		rootDirPath = "/tmp"
	}

	dirPath := rootDirPath + "/jeego-web-dist"
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		log.Info("Cloning jeego-web-dist repo into: %s", dirPath)

		cmd := exec.Command("git", "clone", "git@github.com:aymerick/jeego-web-dist.git", dirPath)
		out, err := cmd.Output()
		if err != nil {
			log.Critical(string(out)) // @todo wat ?!
			panic(log.Critical(err))
		}
	} else {
		log.Info("Found jeego-web-dist repo at: %s", dirPath)
	}

	return dirPath
}

func runWebServer(jeego *Jeego) {
	app_path := jeego.config.WebAppPath
	if app_path == "" {
		app_path = setupWebApp(jeego)
	} else {
		if _, err := os.Stat(app_path); os.IsNotExist(err) {
			panic(log.Critical("Web app path specified in conf file does NOT exists: %s", app_path))
		}

		log.Info("Using web app at: %s", app_path)
	}

	go func() {
		log.Info("Starting web server on port %d", jeego.config.WebServerPort)

		addr := fmt.Sprintf(":%d", jeego.config.WebServerPort)

		mux := pat.New()

		// API endpoints

		nodesMeth := "OPTIONS, GET"
		mux.Options("/api/nodes", wrapHandlerOptions(jeego, nodesMeth))
		mux.Get("/api/nodes", wrapHandlerNodes(jeego, nodesMeth))

		nodeMeth := "OPTIONS, GET, PUT"
		mux.Options("/api/nodes/:id", wrapHandlerOptions(jeego, nodeMeth))
		mux.Get("/api/nodes/:id", wrapHandlerNode(jeego, nodeMeth))
		mux.Put("/api/nodes/:id", wrapHandlerUpdateNode(jeego, nodeMeth))

		nodeTempMeth := "OPTIONS, GET"
		mux.Options("/api/nodes/:id/temperatures", wrapHandlerOptions(jeego, nodeTempMeth))
		mux.Get("/api/nodes/:id/temperatures", wrapHandlerNodeTemperatures(jeego, nodeTempMeth))

		nodeLogsMeth := "OPTIONS, GET"
		mux.Options("/api/nodes/:id/logs", wrapHandlerOptions(jeego, nodeLogsMeth))
		mux.Get("/api/nodes/:id/logs", wrapHandlerNodeLogs(jeego, nodeLogsMeth))

		http.Handle("/api/", mux)

		// Web App files

		http.Handle("/", http.FileServer(http.Dir(app_path)))

		err := http.ListenAndServe(addr, nil)
		if err != nil {
			panic(log.Critical(err))
		}
	}()
}
