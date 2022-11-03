package main

//go:generate stringer -type=ast_type,ast_modifier,file_type -output=parser_string.go

const title = "Spindle 0.4.0"

type spindle struct {
	config       *config
	server_mode  bool

	errors       *error_handler
	file_tree    *disk_object

	templates    map[uint32]*template_object
	partials     map[uint32]*partial_object

	finder_cache map[string]*disk_object

	generated_images map[uint32]*generated_image
}

func main() {
	config, ok := get_arguments()

	if !ok {
		return // @error
	}

	switch config.command {
	case INIT:
		command_init(config)
	case BUILD:
		command_build(config)
	}
}