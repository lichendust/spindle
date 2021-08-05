package main

import (
	"os"
	"fmt"
)

var console_handler *console

func load_config() bool {
	raw_text, ok := load_file("config/config.x")

	if !ok {
		return false
	}

	config = markup_parser(raw_text)
	config.vars = merge_maps(config.vars, tag_defaults)

	if _, ok := config.vars["domain"]; !ok {
		fmt.Println(`"domain" missing from config.x`)
		return false
	}

	return true
}

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Println(help_message)
		return
	}

	if args[0] == "new" {
		make_blank_project()
		fmt.Println(new_project_message)
		return
	}

	if !load_config() {
		fmt.Println("not a spindle project!")
		return
	}

	console_handler = &console {
		scrollback: make([]string, 0, 10),
	}

	switch args[0] {
	case "build":
		build_project(args[1:])

	case "serve":
		serve_project(args[1:])

	default:
		fmt.Println(help_message)
	}
}