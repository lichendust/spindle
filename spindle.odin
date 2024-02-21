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

import "core:os"
import "core:fmt"
import "core:strings"
import "core:unicode"
import "core:runtime"
import "core:sys/windows"
import "core:path/filepath"
import "core:path/slashpath"

import lua "vendor:lua/5.4"

VERSION :: "v0.5.0"
SPINDLE :: "Spindle " + VERSION

SCRIPT_PATH :: "_data"
SCRIPT_NAME :: "spindle.lua"
SCRIPT_TEXT :: #load(SCRIPT_NAME)

current_dir: string
file_array:  [dynamic]string

main :: proc() {
	when ODIN_OS == .Windows {
		windows.SetConsoleOutputCP(windows.CP_UTF8)
	}

	args := os.args[1:]

	if len(args) == 0 {
		fmt.println(USAGE)
		return
	}

	current_dir = os.get_current_directory()

	switch args[0] {
	case "init":
		init_project(args[1:])
	case "build":
		execute_spindle(args[1:])
	case:
		fmt.println(USAGE)
	}
}

init_project :: proc(args: []string) {
	os.make_directory("_data")
	os.make_directory("_site")

	if len(args) > 0 && args[0] == "--local" {
		os.write_entire_file(SCRIPT_PATH + "/" + SCRIPT_NAME, SCRIPT_TEXT)
	}
}

get_script_location :: proc() -> (cstring, bool) {
	DATA_PATH :: SCRIPT_PATH + "/" + SCRIPT_NAME
	if os.exists(DATA_PATH) {
		return DATA_PATH, true
	}

	context.allocator = context.temp_allocator

	here := filepath.dir(os.args[0])
	path := []string{here, SCRIPT_NAME}
	exec := filepath.join(path)

	if os.exists(exec) {
		return strings.clone_to_cstring(exec), true
	}
	return "", false
}

execute_spindle :: proc(args: []string) {
	load_project_dir()

	ctx := lua.L_newstate()
	lua.L_openlibs(ctx)

	lua.createtable(ctx, 0, 16)
	lua.setglobal(ctx, "spindle")

	register_procedure(ctx, "make_slug",      lua_make_slug)
	register_procedure(ctx, "file_exists",    lua_file_exists)
	register_procedure(ctx, "find_file",      lua_find_file)
	register_procedure(ctx, "copy_file",      lua_copy_file)
	register_procedure(ctx, "quoted_split",   lua_quoted_split)
	register_procedure(ctx, "make_directory", lua_make_directory)

	lua_loaded: i32

	path, found_real := get_script_location()
	if found_real {
		lua_loaded = i32(lua.L_dofile(ctx, path))
	} else {
		lua_loaded = i32(lua.L_dostring(ctx, cstring(SCRIPT_TEXT)))
	}

	if lua_loaded != LUA_OK {
		fmt.eprintln(lua.L_tostring(ctx, -1, nil))
		return
	}

	lua.getglobal(ctx, "main")

	if len(args) > 0 {
		lua.pushstring(ctx, strings.clone_to_cstring(args[0], context.temp_allocator))
	} else {
		lua.pushstring(ctx, "index.x")
	}

	lua.call(ctx, 1, 0)
	lua.close(ctx)
}

register_procedure :: proc(ctx: ^lua.State, name: cstring, func: lua.CFunction) {
	lua.getglobal(ctx, "spindle")

	lua.pushstring(ctx, name)
	lua.pushcfunction(ctx, func)
	lua.settable(ctx, -3)

	lua.pop(ctx, 1)
}

// the vendor bindings for Lua have these as enums, which are
// a 'distinct' i32 and therefore uncomparable without explicit
// casting, which is (and this is true) dumb as shit when
// there's nothing that requires them to be explicitly typed.
LUA_OK        :: 0
LUA_YIELD     :: 1
LUA_ERRRUN    :: 2
LUA_ERRSYNTAX :: 3
LUA_ERRMEM    :: 4
LUA_ERRERR    :: 5
LUA_ERRFILE   :: 6

// @todo make file_exists return (success, is_dir) as a double-arg
lua_file_exists :: proc "c" (ctx: ^lua.State) -> i32 {
	context = runtime.default_context()
	defer free_all(context.temp_allocator)

	file_name := strings.clone_from_cstring(lua.L_checkstring(ctx, 1), context.temp_allocator)
	lua.pushboolean(ctx, b32(os.exists(file_name)))
	return 1
}

lua_current_dir :: proc "c" (ctx: ^lua.State) -> i32 {
	context = runtime.default_context()
	defer free_all(context.temp_allocator)

	lua.pushstring(ctx, strings.clone_to_cstring(current_dir, context.temp_allocator))
	return 1
}

lua_copy_file :: proc "c" (ctx: ^lua.State) -> i32 {
	context = runtime.default_context()
	defer free_all(context.temp_allocator)

	// @todo handle argument errors here
	source_path := strings.clone_from_cstring(lua.L_checkstring(ctx, 1), context.temp_allocator)
	output_path := strings.clone_from_cstring(lua.L_checkstring(ctx, 2), context.temp_allocator)

	blob, success := os.read_entire_file_from_filename(source_path, context.temp_allocator)
	if success {
		os.write_entire_file(output_path, blob)
	}

	lua.pushboolean(ctx, b32(success))
	return 1
}

quoted_split :: proc(input: string, allocator := context.allocator) -> []string {
	input := input

	if input == "" {
		return nil
	}

	args := make([dynamic]string, 0, 8, context.temp_allocator)
	is_quote := false

	for {
		if len(input) == 0 {
			break
		}

		if is_quote {
			for c, i in input {
				if c == '"' {
					is_quote = false
					append(&args, input[:i])
					input = input[i + 1:]
					break
				}
			}
			continue
		}

		input = strings.trim_left_space(input)

		if len(input) == 0 {
			break
		}

		if rune(input[0]) == '"' {
			is_quote = true
			input = input[1:]
			continue
		}

		word := extract_non_space_word(input)
		append(&args, word)
		input = input[len(word):]
	}

	return args[:]
}

lua_quoted_split :: proc "c" (ctx: ^lua.State) -> i32 {
	context = runtime.default_context()
	defer free_all(context.temp_allocator)

	input := strings.clone_from_cstring(lua.L_checkstring(ctx, 1), context.temp_allocator)
	args  := quoted_split(input, context.temp_allocator)

	if args == nil {
		return 0
	}

	lua.createtable(ctx, 0, i32(len(args)))

	for entry, i in args {
		lua.pushnumber(ctx, lua.Number(i + 1))
		lua.pushstring(ctx, strings.clone_to_cstring(entry, context.temp_allocator))
		lua.settable(ctx, -3)
	}

	return 1
}

extract_non_space_word :: proc(x: string) -> string {
	for c, i in x {
		if unicode.is_space(c) {
			return x[:i]
		}
	}
	return x
}

lua_make_slug :: proc "c" (ctx: ^lua.State) -> i32 {
	context = runtime.default_context()
	defer free_all(context.temp_allocator)

	input  := strings.clone_from_cstring(lua.L_checkstring(ctx, 1), context.temp_allocator)
	buffer := strings.builder_make_len_cap(0, len(input), context.temp_allocator)

	inside_element := false

	for c in input {
		if c == '<' {
			inside_element = true
			continue
		}
		if c == '>' {
			inside_element = false
			strings.write_rune(&buffer, '-')
			continue
		}
		if inside_element do continue

		if unicode.is_letter(c) || unicode.is_number(c) {
			strings.write_rune(&buffer, unicode.to_lower(c))
			continue
		}
		if unicode.is_space(c) || c == '-' {
			strings.write_rune(&buffer, '-')
			continue
		}
	}

	lua.pushstring(ctx, strings.clone_to_cstring(strings.to_string(buffer), context.temp_allocator))
	return 1
}

// @todo super naive but it'll do for now
find_file :: proc(name: string) -> string {
	short_list := make([dynamic]string, 0, 16, context.temp_allocator)

	name := strings.to_lower(name, context.temp_allocator)

	for entry in file_array {
		a := strings.to_lower(entry, context.temp_allocator)
		if strings.contains(a, name) {
			append(&short_list, entry)
		}
	}

	if len(short_list) == 0 {
		return ""
	}

	index    := 0
	shortest := 9999
	for entry, i in short_list {
		length := len(entry)
		if length < shortest {
			shortest = length
			index    = i
		}
		if strings.contains(entry, "index") {
			return entry // @todo noooooooooooo
		}
	}

	return short_list[index]
}

lua_find_file :: proc "c" (ctx: ^lua.State) -> i32 {
	context = runtime.default_context()
	defer free_all(context.temp_allocator)

	file_name := strings.clone_from_cstring(lua.L_checkstring(ctx, 1), context.temp_allocator)

	path := find_file(file_name)
	if len(path) == 0 {
		return 0
	}

	lua.pushstring(ctx, strings.clone_to_cstring(path, context.temp_allocator))
	return 1
}

lua_make_directory :: proc "c" (ctx: ^lua.State) -> i32 {
	context = runtime.default_context()
	file_name := strings.clone_from_cstring(lua.L_checkstring(ctx, 1), context.temp_allocator)

	if slashpath.ext(file_name) != "" {
		file_name = slashpath.dir(file_name, context.temp_allocator)
	}

	os.make_directory(file_name)
	return 0
}

load_project_dir :: proc() {
	walk_proc :: proc(info: os.File_Info, in_err: os.Errno, user_data: rawptr) -> (os.Errno, bool) {
		if info.name[0] == '.' {
			return 0, info.is_dir
		}
		if info.is_dir && info.name[0] == '_' {
			return 0, true
		}

		// these *are* permanent allocations
		path, _ := filepath.rel(current_dir, info.fullpath, context.allocator)
		str,  _ := filepath.to_slash(path, context.allocator)

		append(&file_array, str)
		return 0, false
	}

	file_array = make([dynamic]string, 0, 64)
	filepath.walk(".", walk_proc, nil)
}

USAGE :: SPINDLE + `

	spindle [command] <flags>

Commands
--------

	init     create a new project
	build    build/render the project

init
----

	Creates the project directory structure
	in the current folder.

	--local    makes a local copy of the
			   spindle.lua file

    You can safely run this on an existing
    project if you wish to 'upgrade' it to
    a local copy during development.

build
-----

	Output the current project's rendered
	files to the _site directory.
`
