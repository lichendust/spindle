package main

const VERSION = "v0.4.1"
const SPINDLE = "Spindle " + VERSION

type spindle struct {
	server_mode  bool

	errors       *Error_Handler
	file_tree    *File

	config

	has_webp     bool

	finder_cache map[string]*File

	pages        map[string]*Page
	templates    map[uint32]*Support_Markup
	partials     map[uint32]*Support_Markup

	gen_pages    map[string]*Page
	gen_images   map[uint32]*Image
}

func main() {
	config, ok := get_arguments()
	if !ok {
		return // @error
	}

	switch config.command {
	case COMMAND_HELP:
		println(SPINDLE)
		return
	case COMMAND_VERSION:
		println(SPINDLE)
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