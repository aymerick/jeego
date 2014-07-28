package ws_hub

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

type WsConnection struct {
	ws *websocket.Conn

	send chan []byte
}

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

/**
 * Hub
 */

type WsHub struct {
	// Registered connections.
	connections map[*WsConnection]bool

	// Inbound messages from the connections.
	broadcast chan []byte

	// Register requests from the connections.
	register chan *WsConnection

	// Unregister requests from connections.
	unregister chan *WsConnection
}

func New() *WsHub {
	return &WsHub{
		broadcast:   make(chan []byte),
		register:    make(chan *WsConnection),
		unregister:  make(chan *WsConnection),
		connections: make(map[*WsConnection]bool),
	}
}

func Run() *WsHub {
	result := New()

	go result.run()

	return result
}

func (hub *WsHub) RegisterConn(wsConn *websocket.Conn) *WsConnection {
	conn := &WsConnection{send: make(chan []byte, 256), ws: wsConn}
	hub.register <- conn

	return conn
}

func (hub *WsHub) UnregisterConn(conn *WsConnection) {
	hub.unregister <- conn
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

func (hub *WsHub) SendMsg(msg []byte) {
	hub.broadcast <- msg
}
