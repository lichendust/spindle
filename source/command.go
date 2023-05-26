package main

import (
	"os"
	"path/filepath"
)

const (
	COMMAND_VERSION uint8 = iota
	COMMAND_HELP
	COMMAND_INIT
	COMMAND_BUILD
	COMMAND_SERVE
)

func command_init(config *Config) {
	if config.output_path != "" {
		make_dir(config.output_path)
		os.Chdir(config.output_path)
	}

	make_dir(TEMPLATE_PATH)
	make_dir(PARTIAL_PATH)
	make_dir(SCRIPT_PATH)
	make_dir(SOURCE_PATH)

	write_file(filepath.Join(TEMPLATE_PATH, "main" + EXTENSION),  MAIN_TEMPLATE)
	write_file(filepath.Join(SOURCE_PATH, "index" + EXTENSION), INDEX_TEMPLATE)
	write_file(CONFIG_FILE_PATH, CONFIG_TEMPLATE)
}