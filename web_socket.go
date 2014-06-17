package main

import (
	log "code.google.com/p/log4go"
	"github.com/gorilla/websocket"
)

/**
 * Reference: https://gist.github.com/garyburd/1316852
 */

/**
 * Connection
 */

type wsConnection struct {
	ws *websocket.Conn

	send chan []byte
}

func (conn *wsConnection) writer() {
	for message := range conn.send {
		err := conn.ws.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Error(err)
			break
		}
	}
	conn.ws.Close()
}

/**
 * Hub
 */

type WsHub struct {
	// Registered connections.
	connections map[*wsConnection]bool

	// Inbound messages from the connections.
	broadcast chan []byte

	// Register requests from the connections.
	register chan *wsConnection

	// Unregister requests from connections.
	unregister chan *wsConnection
}

func newWsHub() *WsHub {
	return &WsHub{
		broadcast:   make(chan []byte),
		register:    make(chan *wsConnection),
		unregister:  make(chan *wsConnection),
		connections: make(map[*wsConnection]bool),
	}
}

func runWsHub() *WsHub {
	result := newWsHub()

	go result.run()

	return result
}

func (hub *WsHub) run() {
	for {
		select {
		case conn := <-hub.register:
			hub.connections[conn] = true
		case conn := <-hub.unregister:
			delete(hub.connections, conn)
			close(conn.send)
		case m := <-hub.broadcast:
			for conn := range hub.connections {
				select {
				case conn.send <- m:
				default:
					delete(hub.connections, conn)
					close(conn.send)
				}
			}
		}
	}
}

func (hub *WsHub) sendMsg(msg []byte) {
	hub.broadcast <- msg
}
