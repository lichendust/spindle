package main

import (
	"time"
	"runtime"
	"os/exec"
	"net/http"
	// "path/filepath"

	"github.com/gorilla/websocket"
)

const serve_port     = ":3011"
const reload_address = "/_spindle/reload"

func open_browser(port string) {
	const url = "http://localhost" + serve_port

	var err error

	switch runtime.GOOS {
	case "linux":   err = exec.Command("xdg-open", url).Start()
	case "darwin":  err = exec.Command("open", url).Start()
	case "windows": err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	}

	if err != nil {
		panic(err)
	}

	_println(title)
	_println("\n   ", url)
}

func command_serve(spindle *spindle) {
	the_server := http.NewServeMux()

	spindle.finder_cache = make(map[string]*disk_object, 64)
	spindle.gen_images   = make(map[uint32]*gen_image, 32)
	spindle.gen_pages    = make(map[string]*gen_page, 32)

	spindle.partials  = load_all_partials(spindle)
	spindle.templates = load_all_templates(spindle)

	if spindle.errors.has_errors() {
		_println(spindle.errors.render_term_errors())
		return
	}

	if data, ok := load_file_tree(spindle); ok {
		spindle.file_tree = data
	}

	// websocket hub
	the_hub := &client_hub {
		clients:    make(map[*client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *client),
		unregister: make(chan *client),
	}

	go the_hub.run()

	// server components
	the_server.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		found_file, ok := find_file_hash(spindle.file_tree, new_hash(r.URL.Path))
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			w.Header().Add("Cache-Control", "no-cache")
			w.Write([]byte(t_error_page_not_found))
			return
		}

		if found_file.file_type == MARKUP {
			page, ok := load_page(spindle, found_file.path)
			if ok {
				assembled := render_syntax_tree(spindle, page, 0)

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
	the_server.HandleFunc(reload_address, func(w http.ResponseWriter, r *http.Request) {
		register_client(the_hub, w, r)
	})

	// start server
	go func() {
		err := http.ListenAndServe(serve_port, the_server)
		if err != nil {
			panic(err)
		}
	}()

	open_browser(serve_port)

	// monitor files for changes
	last_run := time.Now()

	for range time.Tick(time.Second) {
		if folder_has_changes(source_path, last_run) {
			if data, ok := load_file_tree(spindle); ok {
				spindle.file_tree = data
				send_reload(the_hub)
			}
			last_run = time.Now()
		} else if folder_has_changes(template_path, last_run) {
			spindle.templates = load_all_templates(spindle)
			last_run = time.Now()
			send_reload(the_hub)
		} else if folder_has_changes(partial_path, last_run) {
			spindle.partials = load_all_partials(spindle)
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

const (
	time_write_wait  = 10 * time.Second
	time_pong_wait   = 60 * time.Second
	time_ping_period = (time_pong_wait * 9) / 10
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

const reload_script = `<script type='text/javascript'>function spindle_reload() {
	var socket = new WebSocket("ws://" + window.location.host + "` + reload_address + `");
	socket.onclose = function(evt) {
		setTimeout(() => spindle_reload(), 2000);
	};
	socket.onmessage = function(evt) {
		location.reload();
	};
};
spindle_reload()</script>`

type client_hub struct {
	clients    map[*client]bool
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
				_println("reload socket: unexpected closure") // @todo
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

func register_client(the_hub *client_hub, w http.ResponseWriter, r *http.Request) {
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

func send_reload(the_hub *client_hub) {
	the_hub.broadcast <- []byte("reload")
}