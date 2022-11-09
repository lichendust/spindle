package main

import "fmt"

const title = "Spindle 0.4.0"

type spindle struct {
	server_mode  bool

	errors       *error_handler
	file_tree    *disk_object

	config

	pages        map[string]*page_object
	templates    map[uint32]*template_object
	partials     map[uint32]*partial_object

	finder_cache map[string]*disk_object

	gen_pages    map[string]*gen_page
	gen_images   map[uint32]*gen_image
}

func main() {
	config, ok := get_arguments()

	if !ok {
		return // @error
	}

	switch config.command {
	case HELP:
		fmt.Println(title)
		return
	case VERSION:
		fmt.Println(title)
		return
	case INIT:
		command_init(&config)
		return
	}

	spindle := spindle{}
	spindle.config = config
	spindle.errors = new_error_handler()

	switch config.command {
	case BUILD:
		command_build(&spindle)
	case SERVE:
		spindle.server_mode = true
		command_serve(&spindle)
	}
}