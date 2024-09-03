/*
	MIT License
	Spindle: a static site generator
	Copyright (C) 2019-2024 Harley Denham
*/

package main

import "base:runtime"

import "core:os"
import "core:fmt"
import "core:strings"
import "core:unicode"
import "core:sys/windows"
import "core:path/filepath"
import "core:path/slashpath"

import lua "lua/5.4"

VERSION :: "v0.5.0"
SPINDLE :: "Spindle " + VERSION

OUTPUT_PATH :: "_site"
CONFIG_PATH :: "_data"
CONFIG_NAME :: "config.lua"
SCRIPT_NAME :: "spindle.lua"
CONFIG_TEXT :: #load(CONFIG_NAME)
SCRIPT_TEXT :: #load(SCRIPT_NAME)

current_dir: string
file_array:  [dynamic]string

main :: proc() {
	when ODIN_OS == .Windows {
		windows.SetConsoleOutputCP(.UTF8)
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
		execute_spindle(args[1:], false)
	case "prod":
		execute_spindle(args[1:], true)
	case:
		fmt.println(USAGE)
	}
}

init_project :: proc(args: []string) {
	os.make_directory(CONFIG_PATH)
	os.make_directory(OUTPUT_PATH)

	if len(args) > 0 && args[0] == "--local" {
		os.write_entire_file(CONFIG_PATH + "/" + SCRIPT_NAME, SCRIPT_TEXT)
	}

	CONF :: CONFIG_PATH + "/" + CONFIG_NAME
	if !os.exists(CONF) {
		os.write_entire_file(CONF, CONFIG_TEXT)
	}
}

get_script_location :: proc() -> (cstring, bool) {
	DATA_PATH :: CONFIG_PATH + "/" + SCRIPT_NAME
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

execute_spindle :: proc(args: []string, is_prod: bool) {
	load_project_dir()

	ctx := lua.L_newstate()
	lua.L_openlibs(ctx)

	lua.createtable(ctx, 0, 16)
	lua.setglobal(ctx, "spindle")

	register_procedure(ctx, "to_slug",           lua_to_slug)
	register_procedure(ctx, "file_exists",       lua_file_exists)
	register_procedure(ctx, "find_file",         lua_find_file)
	register_procedure(ctx, "find_file_pattern", lua_find_file_pattern)
	register_procedure(ctx, "copy_file",         lua_copy_file)
	register_procedure(ctx, "split_fields",      lua_field_split)
	register_procedure(ctx, "split_quoted",      lua_quoted_split)
	register_procedure(ctx, "make_directory",    lua_make_directory)
	register_procedure(ctx, "write_file",        lua_write_file)
	register_procedure(ctx, "to_title",          lua_title_case)

	register_procedure(ctx, "set_working_directory", lua_set_working_directory)
	register_procedure(ctx, "get_working_directory", lua_get_working_directory)
	register_procedure(ctx, "size_of_file",          lua_size_of_file)

	// @note what is this? who put this here? I know it was me but why??
	register_procedure(ctx, "_balance_parens",   lua_balance_parentheses)

	{
		lua.getglobal(ctx, "spindle")
		lua.pushstring(ctx, "production")
		lua.pushboolean(ctx, b32(is_prod))
		lua.settable(ctx, -3)
		lua.pop(ctx, 1)
	}

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

	file_name := read_nth_string(ctx, 1)
	lua.pushboolean(ctx, b32(os.exists(file_name)))
	return 1
}

lua_copy_file :: proc "c" (ctx: ^lua.State) -> i32 {
	context = runtime.default_context()
	defer free_all(context.temp_allocator)

	// @todo handle argument errors here
	source_path := read_nth_string(ctx, 1)
	output_path := read_nth_string(ctx, 2)

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

	input := read_nth_string(ctx, 1)
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

lua_field_split :: proc "c" (ctx: ^lua.State) -> i32 {
	context = runtime.default_context()
	defer free_all(context.temp_allocator)

	// @todo just check for whitespace instead of always calling fields
	// then we can run a fields_iterator and save the fields allocation

	input := read_nth_string(ctx, 1)
	args  := strings.fields(input, context.temp_allocator)
	count := len(args)

	if count == 0 {
		return 0
	}

	lua.createtable(ctx, 0, i32(count))

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

lua_to_slug :: proc "c" (ctx: ^lua.State) -> i32 {
	context = runtime.default_context()
	defer free_all(context.temp_allocator)

	input  := read_nth_string(ctx, 1)
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

/*lua_thousand_sep :: proc "c" (ctx: ^lua.State) -> i32 {
	@todo

	context = runtime.default_context()
	defer free_all(context.temp_allocator)

	arg1 := read_nth_string(ctx, 1)
	arg2 := strings.clone_from_cstring(lua.L_checkstring(ctx, 2), context.temp_allocator)
}*/

title_case :: proc(input: string) -> string {

	// @todo this is not an exhaustive list
	short_words :: proc(t: string) -> bool {
		switch t {
		case "a":   return true
		case "an":  return true
		case "and": return true
		case "for": return true
		case "in":  return true
		case "is":  return true
		case "nor": return true
		case "of":  return true
		case "on":  return true
		case "or":  return true
		case "the": return true
		case "to":  return true
		}
		return false
	}

	input  := strings.to_lower(input, context.temp_allocator)
	output := strings.builder_make(0, len(input), context.temp_allocator)

	word_index := -1
	for word in strings.fields_iterator(&input) {
		word_index += 1
		if word_index > 0 && short_words(word) {
			strings.write_string(&output, word)
			strings.write_rune(&output, ' ')
			continue
		}

		set_next := false
		for c, index in word {
			x: rune
			if unicode.is_letter(c) && (index == 0 || set_next) {
				x = unicode.to_upper(c)
			} else {
				x = unicode.to_lower(c)
			}
			set_next = (c == '-' || c == '—')
			strings.write_rune(&output, x)
		}

		// @todo this currently re-spaces the string
		strings.write_rune(&output, ' ')
	}

	return strings.to_string(output)
}

lua_title_case :: proc "c" (ctx: ^lua.State) -> i32 {
	context = runtime.default_context()
	defer free_all(context.temp_allocator)

	input := read_nth_string(ctx, 1)
	lua.pushstring(ctx, strings.clone_to_cstring(title_case(input), context.temp_allocator))
	return 1
}

load_project_dir :: proc() {
	walk_proc :: proc(info: os.File_Info, in_err: os.Errno, user_data: rawptr) -> (os.Errno, bool) {
		if info.name[0] == '.' {
			return 0, info.is_dir
		}
		if info.is_dir {
			if info.name[0] == '_' {
				return 0, true
			}
			return 0, false
		}

		// this *is* a permanent allocation
		path, _ := filepath.rel(current_dir, info.fullpath, context.allocator)
		str, did_alloc := filepath.to_slash(path, context.allocator)
		if did_alloc {
			delete(path)
		}

		append(&file_array, str)
		return 0, false
	}

	file_array = make([dynamic]string, 0, 64)
	filepath.walk(".", walk_proc, nil)
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

		if strings.contains(entry, "index") {
			length -= 5
		}

		if length < shortest {
			shortest = length
			index    = i
		}
	}

	return short_list[index]
}

lua_find_file :: proc "c" (ctx: ^lua.State) -> i32 {
	context = runtime.default_context()
	defer free_all(context.temp_allocator)

	file_name := read_nth_string(ctx, 1)

	path := find_file(file_name)
	if len(path) == 0 {
		return 0
	}

	lua.pushstring(ctx, strings.clone_to_cstring(path, context.temp_allocator))
	return 1
}

find_file_pattern :: proc(pattern: string) -> []string {
	short_list := make([dynamic]string, 0, len(file_array), context.temp_allocator)
	pattern := strings.to_lower(pattern, context.temp_allocator)

	for entry in file_array {
		a := strings.to_lower(entry, context.temp_allocator)
		if strings.contains(a, pattern) {
			append(&short_list, entry)
		}
	}

	return short_list[:]
}

lua_find_file_pattern :: proc "c" (ctx: ^lua.State) -> i32 {
	context = runtime.default_context()
	defer free_all(context.temp_allocator)

	file_name := read_nth_string(ctx, 1)

	search := find_file_pattern(file_name)
	if len(search) == 0 {
		return 0
	}

	lua.createtable(ctx, 0, i32(len(search)))

	for entry, i in search {
		lua.pushnumber(ctx, lua.Number(i + 1))
		lua.pushstring(ctx, strings.clone_to_cstring(entry, context.temp_allocator))
		lua.settable(ctx, -3)
	}

	return 1
}

// @todo this does not function like mkdir -p -- nested directories aren't created
make_directory :: #force_inline proc(file_name: string) {
	if slashpath.ext(file_name) != "" {
		os.make_directory(slashpath.dir(file_name, context.temp_allocator))
		return
	}
	os.make_directory(file_name)
}

lua_make_directory :: proc "c" (ctx: ^lua.State) -> i32 {
	context = runtime.default_context()
	defer free_all(context.temp_allocator)

	file_name := read_nth_string(ctx, 1)
	make_directory(file_name)
	return 0
}

lua_write_file :: proc "c" (ctx: ^lua.State) -> i32 {
	context = runtime.default_context()
	defer free_all(context.temp_allocator)

	file_name := read_nth_string(ctx, 1)
	file_blob := read_nth_string(ctx, 2)

	make_directory(file_name)
	os.write_entire_file(file_name, transmute([]u8) file_blob)
	return 0
}

lua_balance_parentheses :: proc "c" (ctx: ^lua.State) -> i32 {
	context = runtime.default_context()
	defer free_all(context.temp_allocator)

	input := read_nth_string(ctx, 1)
	count := 0

	for c in input {
		if c == '{' do count += 1
		if c == '}' do count -= 1
	}

	lua.pushnumber(ctx, lua.Number(count))
	return 1
}

lua_set_working_directory :: proc "c" (ctx: ^lua.State) -> i32 {
	context = runtime.default_context()

	file  := read_nth_string(ctx, 1)
	errno := os.set_current_directory(file)

	lua.pushboolean(ctx, errno == 0)
	return 1
}

lua_get_working_directory :: proc "c" (ctx: ^lua.State) -> i32 {
	// @note turns out Odin's `get_current_directory` procedure
	// accepts an allocator on Windows, but not on Darwin or
	// Linux, so we need to just set the allocator to temp for
	// the scope.
	context = runtime.default_context()
	context.allocator = context.temp_allocator

	str := os.get_current_directory()
	lua.pushstring(ctx, strings.clone_to_cstring(str))
	return 1
}

lua_size_of_file :: proc "c" (ctx: ^lua.State) -> i32 {
	context = runtime.default_context()

	file := read_nth_string(ctx, 1)
	size := os.file_size_from_path(file)

	lua.pushinteger(ctx, cast(lua.Integer) size)
	return 1
}

read_nth_string :: #force_inline proc(ctx: ^lua.State, n: int) -> string {
	return strings.clone_from_cstring(lua.L_checkstring(ctx, cast(i32) n), context.temp_allocator)
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
