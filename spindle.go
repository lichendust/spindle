package main

//go:generate stringer -type=ast_type,ast_modifier,file_type -output=parser_string.go

const title = "Spindle 0.4.0"

type spindle struct {
	config       *config
	server_mode  bool

	file_tree    *disk_object
	templates    map[uint32]*template_object
	finder_cache map[string]*disk_object
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