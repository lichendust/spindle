package main

import (
	"os"
	"fmt"
	"strconv"
)

var (
	config *global_config
	console_handler *console
)

type global_config struct {
	vars map[string]string

	image_target        int
	image_jpeg_quality  int
}

func load_config() bool {
	raw_text, ok := load_file("config/config.x")

	if !ok {
		return false
	}

	config = &global_config {}

	data := markup_parser(raw_text)

	if x, ok := data.vars["image_target"]; ok {
		n, err := strconv.Atoi(x)

		if err != nil {
			console_handler.print(`"image_target" in config.x: "%s" is not a number`, x)
			return false
		}

		config.image_target = n
		delete(data.vars, "image_target")
	}

	if config.image_target > 0 {
		config.image_jpeg_quality = 85 // @todo
	}

	config.vars = merge_maps(data.vars, tag_defaults)

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