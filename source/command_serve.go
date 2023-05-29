/*
	Spindle
	A static site generator
	Copyright (C) 2022-2023 Harley Denham

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package main

import (
	"time"
	"runtime"
	"strings"
	"os/exec"
	"net/http"
	// "path/filepath"

	"github.com/gorilla/websocket"
)

const SERVE_PORT     = ":3011"

const SPINDLE_PREFIX = "/_spindle/"
const RELOAD_ADDRESS = SPINDLE_PREFIX + "reload"
const MANUAL_ADDRESS = SPINDLE_PREFIX + "manual/"

const TIME_WRITE_WAIT  = 10 * time.Second
const TIME_PONG_WAIT   = 60 * time.Second
const TIME_PING_PERIOD = (TIME_PONG_WAIT * 9) / 10

func open_browser(port string) {
	const url = "http://localhost" + SERVE_PORT

	var err error

	switch runtime.GOOS {
	case "linux":   err = exec.Command("xdg-open", url).Start()
	case "darwin":  err = exec.Command("open", url).Start()
	case "windows": err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	}

	if err != nil {
		panic(err)
	}

	println(SPINDLE)
	println("\n   ", url)
}

func command_serve(spindle *Spindle) {
	the_server := http.NewServeMux()

	spindle.finder_cache = make(map[string]*File, 64)
	spindle.gen_pages    = make(map[string]*Gen_Page, 32)
	spindle.gen_images   = make(map[uint32]*Image,   32)

	spindle.templates = load_support_directory(spindle, TEMPLATE, TEMPLATE_PATH)
	spindle.partials  = load_support_directory(spindle, PARTIAL,  PARTIAL_PATH)

	if spindle.errors.has_errors() {
		println(spindle.errors.render_term_errors())
		return
	}

	if data, ok := load_file_tree(spindle); ok {
		spindle.file_tree = data
	}

	// websocket hub
	the_hub := &Client_Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}

	go the_hub.run()

	// server components
	the_server.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		found_file, ok := find_file_hash(spindle.file_tree, new_hash(r.URL.Path))

		if !ok {
			if gen, ok := spindle.gen_pages[r.URL.Path]; ok {
				if page, ok := load_page_from_file(spindle, gen.file); ok {
					page.file        = gen.file
					page.import_cond = gen.import_cond
					page.import_hash = gen.import_hash

					assembled := render_syntax_tree(spindle, page)

					if spindle.errors.has_errors() {
						assembled = spindle.errors.render_html_page()
						spindle.errors.reset()
					}

					w.WriteHeader(http.StatusOK)
					w.Header().Add("Cache-Control", "no-cache")
					w.Write([]byte(assembled))
					return
				}
			}

			w.WriteHeader(http.StatusNotFound)
			w.Header().Add("Cache-Control", "no-cache")
			w.Write([]byte(ERROR_PAGE_NOT_FOUND))
			return
		}

		if found_file.file_type == MARKUP {
			page, ok := load_page_from_file(spindle, found_file)
			if ok {
				assembled := render_syntax_tree(spindle, page)

				if spindle.errors.has_errors() {
					assembled = spindle.errors.render_html_page()
					spindle.errors.reset()
				}

				w.WriteHeader(http.StatusOK)
				w.Header().Add("Cache-Control", "no-cache")
				w.Write([]byte(assembled))
				return

			} else {
				w.WriteHeader(http.StatusNotFound)
				return
			}
		}

		w.Header().Add("Cache-Control", "no-cache")
		http.ServeFile(w, r, found_file.path)
	})

	// socket reloader
	the_server.HandleFunc(RELOAD_ADDRESS, func(w http.ResponseWriter, r *http.Request) {
		register_client(the_hub, w, r)
	})

	// built-in manual server
	the_server.HandleFunc(MANUAL_ADDRESS, func(w http.ResponseWriter, r *http.Request) {
		request := r.URL.Path[len(MANUAL_ADDRESS):]
		content := manual_content(request)

		// manually set MIME types for the manual
		// because ServeFile isn't here to save us
		if strings.HasSuffix(request, ".css") {
			w.Header().Set("Content-Type", "text/css")
		} else if strings.HasSuffix(request, ".js") {
			w.Header().Set("Content-Type", "text/js")
		}

		w.Header().Add("Cache-Control", "no-cache")
		w.Write([]byte(content))
	})

	// start server
	go func() {
		err := http.ListenAndServe(SERVE_PORT, the_server)
		if err != nil {
			panic(err)
		}
	}()

	open_browser(SERVE_PORT)

	// monitor files for changes
	last_run := time.Now()

	for range time.Tick(time.Second) {
		if folder_has_changes(SOURCE_PATH, last_run) {
			if data, ok := load_file_tree(spindle); ok {
				spindle.file_tree = data

				// @todo gen_page cache expiry in server
				/*for x := range spindle.gen_pages {
					delete(spindle.gen_pages, x)
				}*/

				send_reload(the_hub)
			}

			last_run = time.Now()

		} else if folder_has_changes(TEMPLATE_PATH, last_run) {
			spindle.templates = load_support_directory(spindle, TEMPLATE, TEMPLATE_PATH)
			last_run = time.Now()

			send_reload(the_hub)

		} else if folder_has_changes(PARTIAL_PATH, last_run) {
			spindle.partials = load_support_directory(spindle, PARTIAL,  PARTIAL_PATH)
			last_run = time.Now()

			send_reload(the_hub)
		}
	}
}

/*func serve_public(args []string) {
	the_server := http.NewServeMux()

	the_server.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		w.Header().Add("Cache-Control", "no-cache")

		if path == "/" {
			path = filepath.Join(public_path, "index.html")
		} else {
			path = filepath.Join(public_path, path)

			if is_dir(path) {
				path = filepath.Join(path, "index.html")
			} else if filepath.Ext(path) == "" {
				path += ".html"
			}
		}

		http.ServeFile(w, r, path)
	})

	go func() {
		err := http.ListenAndServe(check_port, the_server)

		if err != nil {
			panic(err)
		}
	}()

	// print_server_info(check_port)
	open_browser("/", check_port)

	for range time.Tick(time.Second * 2) {}
}*/

// it's very possible to do all this with
// golang's own websocket, but for now this
// works fine

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

const RELOAD_SCRIPT = `<script type='text/javascript'>function spindle_reload() {
	var socket = new WebSocket("ws://" + window.location.host + "` + RELOAD_ADDRESS + `");
	socket.onclose = function(evt) {
		setTimeout(() => spindle_reload(), 2000);
	};
	socket.onmessage = function(evt) {
		location.reload();
	};
};
spindle_reload()</script>`

type Client_Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

func (h *Client_Hub) run() {
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

type Client struct {
	socket  *websocket.Conn
	send    chan []byte
}

func (c *Client) read_pump(the_hub *Client_Hub) {
	defer func() {
		the_hub.unregister <- c
		c.socket.Close()
	}()

	for {
		_, _, err := c.socket.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				println("reload socket: unexpected closure") // @todo
			}
			break
		}
	}
}

func (c *Client) write_pump() {
	ticker := time.NewTicker(TIME_PING_PERIOD)

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

			c.socket.SetWriteDeadline(time.Now().Add(TIME_WRITE_WAIT))

			w, err := c.socket.NextWriter(websocket.TextMessage)

			if err != nil {
				return
			}

			w.Write(message)

			n := len(c.send)

			for i := 0; i < n; i += 1 {
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

func (c *Client) write(mt int, payload []byte) error {
	c.socket.SetWriteDeadline(time.Now().Add(TIME_WRITE_WAIT))
	return c.socket.WriteMessage(mt, payload)
}

func register_client(the_hub *Client_Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		panic(err)
	}

	the_client := &Client{
		socket: conn,
		send:   make(chan []byte, 256),
	}

	the_hub.register <- the_client

	go the_client.write_pump()
	the_client.read_pump(the_hub)
}

func send_reload(the_hub *Client_Hub) {
	the_hub.broadcast <- []byte("reload")
}