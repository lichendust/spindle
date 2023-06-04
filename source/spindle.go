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

const VERSION = "v0.4.2"
const SPINDLE = "Spindle " + VERSION

var spindle *Spindle

type Spindle struct {
	server_mode bool

	errors    *Error_Handler
	file_tree *File

	Config

	has_webp bool

	finder_cache map[string]*File

	pages        map[string]*Page
	templates    map[uint32]*Support_Markup
	partials     map[uint32]*Support_Markup

	gen_pages    map[string]*Gen_Page
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
		println(apply_color(HELP_TEXT))
		return
	case COMMAND_VERSION:
		println(SPINDLE)
		return
	case COMMAND_INIT:
		command_init(&config)
		return
	}

	spindle = new(Spindle)

	spindle.Config = config
	spindle.errors = new_error_handler()

	switch config.command {
	case COMMAND_BUILD:
		command_build()
	case COMMAND_SERVE:
		spindle.server_mode = true
		command_serve()
	}
}