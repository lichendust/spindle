package main

import (
	"os"
	"path/filepath"
)

const (
	VERSION uint8 = iota
	HELP
	INIT
	BUILD
	SERVE
)

func command_init(config *config) {
	if config.output_path != "" {
		make_dir(config.output_path)
		os.Chdir(config.output_path)
	}

	make_dir(template_path)
	make_dir(partial_path)
	make_dir(script_path)
	make_dir(source_path)

	write_file(filepath.Join(config_path, "main" + extension),  main_template)
	write_file(filepath.Join(source_path, "index" + extension), index_template)
}