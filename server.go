package main

import (
	"fmt"
	"time"
	"strings"
	"runtime"
	"os/exec"
	"net/http"
	"path/filepath"
)

func serve_source(args []string) {
	the_server := http.NewServeMux()

	// websocket hub
	the_hub := &client_hub {
		clients:    make(map[*client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *client),
		unregister: make(chan *client),
	}

	go the_hub.run()

	// server components
	the_server.HandleFunc("/", resource_finder)
	the_server.HandleFunc("/__spindle", func(w http.ResponseWriter, r *http.Request) {
		reload_socket(the_hub, w, r)
	})

	// start server
	go func() {
		err := http.ListenAndServe(config.main_port, the_server)

		if err != nil {
			panic(err)
		}
	}()

	// print server startup message to user
	print_server_info(config.main_port)
	open_browser("/", config.main_port)

	// monitor files for changes
	last_run := time.Unix(0,0)

	for range time.Tick(time.Second / 2) {
		if file_has_changes("config/config.x", last_run) {
			expire_plates_cache()
			if !load_config(false) {
				fmt.Println("error in config.x, stopping server")
				break
			}
			send_reload(the_hub)
		}

		if directory_has_changes("config/chunks", last_run) {
			expire_chunks_cache()
			send_reload(the_hub)
		}

		if directory_has_changes("config/plates", last_run) {
			expire_plates_cache()
			send_reload(the_hub)
		}

		if directory_has_changes("source", last_run) {
			send_reload(the_hub)
		}

		last_run = time.Now()
	}
}

func serve_public(public_dir string) {
	the_server := http.NewServeMux()

	the_server.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		w.Header().Add("Cache-Control", "no-cache")

		if path == "/" {
			path = filepath.Join(public_dir, "index.html")
		} else {
			path = filepath.Join(public_dir, path)

			if is_dir(path) {
				path = filepath.Join(path, "index.html")
			} else if filepath.Ext(path) == "" {
				path += ".html"
			}
		}

		http.ServeFile(w, r, path)
	})

	go func() {
		err := http.ListenAndServe(config.test_port, the_server)

		if err != nil {
			panic(err)
		}
	}()

	print_server_info(config.test_port)
	open_browser("/", config.test_port)

	for range time.Tick(time.Second * 2) {}
}

func resource_finder(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	w.Header().Add("Cache-Control", "no-cache")

	if strings.HasSuffix(path, ".html") {
		path = path[:len(path) - 5]
	}

	is_markup := false

	if path == "/" {
		is_markup = true
		path = "source/index.x"
	} else {
		path = filepath.Join("source", path)

		if is_dir(path) {
			is_markup = true
			path = filepath.Join(path, "index.x")
		}

		if filepath.Ext(path) == "" {
			is_markup = true
			path += ".x"
		}
	}

	if is_markup {
		page_obj, ok := load_page(path)

		if ok {
			out_text := markup_render(page_obj)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(out_text))
			return
		}

		w.WriteHeader(http.StatusNotFound)
		return
	}

	ext := filepath.Ext(path)

	if ext == ".png" || ext == ".jpg" || ext == ".jpeg" {
		if n := strings.IndexRune(path, '@'); n > -1 {
			if is_file(path) {
				http.ServeFile(w, r, path)
			} else {
				http.ServeFile(w, r, path[:n] + ext)
			}
			return
		}
	}

	http.ServeFile(w, r, path)
}

func print_server_info(port string) {
	// @todo get actual network interfaces and print 'em
	fmt.Printf("spindle server\n\n    localhost%s\n\n", port)
}

func open_browser(path, port string) {
	url := sprint("http://localhost%s/%s", port, path)

	var err error

	switch runtime.GOOS {
	case "linux":   err = exec.Command("xdg-open", url).Start()
	case "darwin":  err = exec.Command("open", url).Start()
	case "windows": err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	}

	if err != nil {
		panic(err)
	}
}