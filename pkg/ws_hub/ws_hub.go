package ws_hub

import (
	log "code.google.com/p/log4go"
	"github.com/gorilla/websocket"
)

/**
 * Reference: https://gist.github.com/garyburd/1316852
 */

// WebSocket connection
type WsConnection struct {
	ws *websocket.Conn

	send chan []byte
}

// Write loop
func (conn *WsConnection) WriterLoop() {
	for message := range conn.send {
		err := conn.ws.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Error(err)
			break
		}
	}
	conn.ws.Close()
}

// WebSocket connections hub
type WsHub struct {
	// Registered connections
	connections map[*WsConnection]bool

	// Message broadcasting channel
	broadcast chan []byte

	// Connection registration channel
	register chan *WsConnection

	// Connection unregistration channel
	unregister chan *WsConnection
}

// Instanciates a new WebSocket hub
func New() *WsHub {
	return &WsHub{
		broadcast:   make(chan []byte),
		register:    make(chan *WsConnection),
		unregister:  make(chan *WsConnection),
		connections: make(map[*WsConnection]bool),
	}
}

// Instanciates a new WebSocket hub and run it
func Run() *WsHub {
	result := New()

	go result.run()

	return result
}

// Run hub
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

// Register a new WebSocket connection
func (hub *WsHub) RegisterConn(wsConn *websocket.Conn) *WsConnection {
	conn := &WsConnection{send: make(chan []byte, 256), ws: wsConn}
	hub.register <- conn

	return conn
}

// Unregister a WebSocket connection
func (hub *WsHub) UnregisterConn(conn *WsConnection) {
	hub.unregister <- conn
}

// Broadcast message to all registered WebSocket connections
func (hub *WsHub) SendMsg(msg []byte) {
	hub.broadcast <- msg
}
