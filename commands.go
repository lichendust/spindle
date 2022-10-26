package main

import "os"
import "path/filepath"

const (
	VERSION uint8 = iota
	HELP
	INIT
	BUILD
	SERVE
)

func command_init(config *config) {
	if config.path != "" {
		make_dir(config.path)
		os.Chdir(config.path)
	}

	make_dir(template_path)
	make_dir(partial_path)
	make_dir(script_path)
	make_dir(source_path)

	write_file(filepath.Join(config_path, "main.x"),  main_template)
	write_file(filepath.Join(source_path, "index.x"), index_template)
}

func command_build(config *config) {
	spindle := &spindle{}
	spindle.config = config

	new_hash("default")

	config.domain = "https://qxoko.io/"
	config.default_path_type = ABSOLUTE

	if data, ok := load_file_tree(); ok {
		spindle.file_tree = data
	}

	spindle.templates    = load_all_templates()
	spindle.finder_cache = make(map[string]*disk_object, 64)

	page, ok := load_page("source/index.x")

	if !ok {
		panic("failed to load page")
	}

	assembled := render_syntax_tree(spindle, page)

	if validate_html(assembled) {
		_println("GOOD!")
	} else {
		_println("BAD!")
	}

	_println(assembled)
}