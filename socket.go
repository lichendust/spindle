package main

import (
	"fmt"
	"time"
	"net/http"

	"github.com/gorilla/websocket"
)

// it's very possible to do all this with
// golang's own websocket, but for now this
// works fine

// @todo
// clean this up, there's a lot of mess
// between this file and server.go, which
// makes for a confusing read

// also need to handle errors better, with
// concise warnings to the user where
// important and careful curation of
// redundant messages where unimportant

const client_socket = `<script type='text/javascript'>function spindle_reload(address) {
	var socket = new WebSocket("ws://" + window.location.host + "/__spindle");
	socket.onclose = function(evt) {
		setTimeout(() => spindle_reload(), 2000);
	};
	socket.onmessage = function(evt) {
		location.reload();
	};
};
spindle_reload()</script>`

const (
	time_write_wait  = 10 * time.Second
	time_pong_wait   = 60 * time.Second
	time_ping_period = (time_pong_wait * 9) / 10
)

var upgrader = websocket.Upgrader {
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func send_reload(the_hub *client_hub) {
	the_hub.broadcast <- []byte("reload")
}

func reload_socket(the_hub *client_hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		panic(err)
	}

	the_client := &client {
		socket: conn,
		send:   make(chan []byte, 256),
	}

	the_hub.register <- the_client

	go the_client.write_pump()
	the_client.read_pump(the_hub)
}

type client_hub struct {
	clients map[*client]bool

	broadcast  chan []byte

	register   chan *client
	unregister chan *client
}

func (h *client_hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true

		case client := <-h.unregister:
			if ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}

		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

type client struct {
	socket  *websocket.Conn
	send    chan []byte
}

func (c *client) read_pump(the_hub *client_hub) {
	defer func() {
		the_hub.unregister <- c
		c.socket.Close()
	}()

	for {
		_, _, err := c.socket.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				fmt.Println("reload socket: unexpected closure") // @todo
			}
			break
		}
	}
}

func (c *client) write_pump() {
	ticker := time.NewTicker(time_ping_period)

	defer func() {
		ticker.Stop()
		c.socket.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}

			c.socket.SetWriteDeadline(time.Now().Add(time_write_wait))

			w, err := c.socket.NextWriter(websocket.TextMessage)

			if err != nil {
				return
			}

			w.Write(message)

			n := len(c.send)

			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func (c *client) write(mt int, payload []byte) error {
	c.socket.SetWriteDeadline(time.Now().Add(time_write_wait))
	return c.socket.WriteMessage(mt, payload)
}