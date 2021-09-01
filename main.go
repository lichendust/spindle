package main

import (
	"os"
	"fmt"
	"strconv"
)

var config *global_config

type global_config struct {
	vars map[string]string

	build_mode bool

	image_resize      int
	image_quality     string
	image_make_webp   bool
}

func load_config() bool {
	raw_text, ok := load_file("config/config.x")

	if !ok {
		return false
	}

	config = &global_config {}

	data := markup_parser(raw_text)

	if x, ok := data.vars["image_resize"]; ok {
		n, err := strconv.Atoi(x)

		if err != nil {
			console_print(`"image_resize" in config.x: "%s" is not a number`, x)
			return false
		}

		config.image_resize = n
		delete(data.vars, "image_resize")
	}

	if x, ok := data.vars["image_quality"]; ok {
		config.image_quality = x
		delete(data.vars, "image_quality")
	} else {
		config.image_quality = "100"
	}

	if _, ok := data.vars["image_make_webp"]; ok {
		config.image_make_webp = true
		delete(data.vars, "image_make_webp")
	}

	config.vars = merge_maps(data.vars, tag_defaults)

	if _, ok := config.vars["domain"]; !ok {
		fmt.Println(`"domain" missing from config.x`)
		return false
	}

	if !config.build_mode {
		config.image_make_webp = false // no conversions in server mode
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

	if args[0] == "new" {
		make_blank_project()
		return
	}

	if !load_config() {
		fmt.Println("not a spindle project!")
		return
	}

	switch args[0] {
	case "build":
		config.build_mode = true
		build_project(args[1:])

	case "serve":
		serve_project(args[1:])

	case "test":
		serve_test()

	default:
		fmt.Println(help_message)
	}
}