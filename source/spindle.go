package main

import "fmt"

const VERSION = "v0.4.0"
const SPINDLE = "Spindle " + VERSION

type spindle struct {
	server_mode  bool

	errors       *error_handler
	file_tree    *disk_object

	config

	has_webp     bool

	pages        map[string]*page_object
	templates    map[uint32]*template_object
	partials     map[uint32]*partial_object

	finder_cache map[string]*disk_object

	gen_pages    map[string]*page_object
	gen_images   map[uint32]*gen_image
}

func main() {
	config, ok := get_arguments()

	if !ok {
		return // @error
	}

	switch config.command {
	case COMMAND_HELP:
		fmt.Println(SPINDLE)
		return
	case COMMAND_VERSION:
		fmt.Println(SPINDLE)
		return
	case COMMAND_INIT:
		command_init(&config)
		return
	}

	spindle := spindle{}
	spindle.config = config
	spindle.errors = new_error_handler()

	switch config.command {
	case COMMAND_BUILD:
		command_build(&spindle)
	case COMMAND_SERVE:
		spindle.server_mode = true
		command_serve(&spindle)
	}
}