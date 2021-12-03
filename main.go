package main

import (
	"os"
	"fmt"
	"strings"
)

var config *global_config

type global_config struct {
	vars map[string]string

	build_mode bool

	image_rewrite_extensions bool
	image_ext map[string]string

	exclude_list []string

	serve_port string
	check_port string
}

func load_config(build bool) bool {
	raw_text, ok := load_file("config/config.x")

	if !ok {
		return false
	}

	config = &global_config {
		build_mode: build,
		image_ext:  make(map[string]string, 4),
	}

	data := markup_parser(raw_text)

	config.vars = merge_maps(data.vars, tag_defaults)

	if _, ok := config.vars["domain"]; !ok {
		fmt.Println(`"domain" missing from config.x`)
		return false
	}

	for key, v := range config.vars {
		if strings.HasPrefix(key, "image_ext.") {
			config.image_rewrite_extensions = true
			config.image_ext[key[10:]] = v
			delete(config.vars, key)
		}
	}

	if x, ok := config.vars["exclude"]; ok {
		config.exclude_list = strings.Fields(x)
		delete(config.vars, "exclude")
	}

	if v, ok := config.vars["serve_port"]; ok {
		if !is_all_numbers(v) {
			panic(sprint(`config.serve_port — invalid port number %s`, v)) // @error
		}

		config.serve_port = ":" + v
		delete(config.vars, "serve_port")
	} else {
		config.serve_port = ":3011" // default port
	}

	if v, ok := config.vars["check_port"]; ok {
		if !is_all_numbers(v) {
			panic(sprint(`config.check_port — invalid port number %s`, v)) // @error
		}

		config.check_port = ":" + v
		delete(config.vars, "check_port")
	} else {
		config.check_port = ":3022" // default port
	}

	if !config.build_mode {
		config.image_rewrite_extensions = false // no conversions in server mode
	}

	return true
}

func console_print(base string, args ...interface{}) {
	fmt.Println(fmt.Sprintf(base, args...))
}

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Println(help_message)
		return
	}

	switch args[0] {
	case "new":
		make_blank_project()

	case "build":
		build_project(args[1:])

	case "serve":
		serve_source()

	case "check":
		serve_public(args[1:])

	default:
		fmt.Println(help_message)
	}
}