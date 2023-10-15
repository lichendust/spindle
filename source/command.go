/*
	Spindle
	A static site generator
	Copyright (C) 2022-2023 Harley Denham

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package main

import "os"
import "path/filepath"

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