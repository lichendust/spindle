package main

//go:generate stringer -type=ast_type,ast_modifier,file_type -output=parser_string.go

import "os"

const title = "Spindle 0.4.0"

type spindle struct {
	config       *config
	file_tree    *disk_object
	templates    map[uint32]*page_object
	finder_cache map[string]*disk_object
}

func main() {
	spindle := &spindle{}

	spindle.config = &config{
		domain: "https://qxoko.io/",
		default_path_type: ABSOLUTE,
	}

	if data, ok := load_file_tree(); ok {
		spindle.file_tree = data
	}

	spindle.finder_cache = make(map[string]*disk_object, 64)

	blob, ok := load_file("X:/dev/spindle/v2/parser_build.x")

	if !ok {
		panic("aa")
	}

	token_stream := lex_blob(blob)
	// print_token_stream(token_stream)

	syntax_tree := parse_stream(token_stream)
	print_syntax_tree(syntax_tree, 0)

	assembled := render_syntax_tree(spindle, &page_object{
		page_path: "source/notes/index.x",
		content:   syntax_tree,
	})

	if validate_html(assembled) {
		_println("GOOD!")
	} else {
		_println("BAD!")
	}

	_println(assembled)
}

func test_loading() {
	spindle := &spindle{}

	if data, ok := load_file_tree(); ok {
		spindle.file_tree = data
	}

	print_file_tree(spindle.file_tree.children, 0)

	_println("")

	if a, ok := find_file_descending(spindle.file_tree, os.Args[1:][0]); ok {
		_println(ok, a)

		b, ok := find_file(spindle.file_tree, a, os.Args[1:][1])
		_println(ok, b)

		_println(filepath_relative(a.path, b.path))
	}
}