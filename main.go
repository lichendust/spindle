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

	main_port string
	test_port string
}

func load_config(build bool) bool {
	raw_text, ok := load_file("config/config.x")

	if !ok {
		return false
	}

	config = &global_config {
		build_mode: build,
	}

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

	if v, ok := data.vars["main_port"]; ok {
		if !is_all_numbers(v) {
			panic(sprint(`config.main_port — invalid port number %s`, v)) // @error
		}

		config.main_port = ":" + v
		delete(data.vars, "main_port")
	} else {
		config.main_port = ":3011" // default port
	}

	if v, ok := data.vars["test_port"]; ok {
		if !is_all_numbers(v) {
			panic(sprint(`config.test_port — invalid port number %s`, v)) // @error
		}

		config.test_port = ":" + v
		delete(data.vars, "test_port")
	} else {
		config.test_port = ":3022" // default port
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

	switch args[0] {
	case "new":
		make_blank_project()

	case "build":
		build_project(args[1:])

	case "serve":
		serve_source()

	case "test":
		serve_public(args[1:]) // @todo rename this command

	default:
		fmt.Println(help_message)
	}
}