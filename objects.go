package main

import (
	"io/fs"
	"path/filepath"
)

type page_object struct {
	// page_id      uint32
	page_path    string
	content      []ast_data
	top_scope    []ast_data
	slug_tracker map[string]uint
	raw_string   string
	position     position
}

type template_object struct {
	has_body      bool
	template_path string
	content       []ast_data
	top_scope     []ast_data
	raw_string    string
	position      position
}

type partial_object struct {
	partial_path string
	content      []ast_data
	raw_string   string
}

func arrange_top_scope(content []ast_data) []ast_data {
	array := make([]ast_data, 0, 8)

	for _, entry := range content {
		if entry.type_check().is(DECL, DECL_TOKEN, DECL_BLOCK, TEMPLATE) {
			array = append(array, entry)
		}
	}

	return array
}

func load_page(spindle *spindle, full_path string) (*page_object, bool) {
	blob, ok := load_file(full_path)

	if !ok {
		return nil, false
	}

	token_stream := lex_blob(full_path, blob)
	// print_token_stream(token_stream)

	syntax_tree := parse_stream(spindle.errors, token_stream, false)
	// print_syntax_tree(syntax_tree, 0)

	/*if spindle.errors.has_failures {
		return nil, false
	}*/

	p := &page_object{
		page_path:    full_path,
		content:      syntax_tree,
		top_scope:    arrange_top_scope(syntax_tree),
		raw_string:   blob,
		slug_tracker: make(map[string]uint, 4),
		position:     position{0,0,0,full_path},
	}
	return p, true
}

func load_template(spindle *spindle, full_path string) (*template_object, bool) {
	blob, ok := load_file(full_path)

	if !ok {
		return nil, false
	}

	token_stream := lex_blob(full_path, blob)
	syntax_tree  := parse_stream(spindle.errors, token_stream, true)
	// print_syntax_tree(syntax_tree, 0)

	t := &template_object{
		content:       syntax_tree,
		template_path: full_path,
		top_scope:     arrange_top_scope(syntax_tree),
		has_body:      recursive_anon_count(syntax_tree) > 0,
		raw_string:    blob,
		position:      position{0,0,0,full_path},
	}
	return t, true
}

func load_all_templates(spindle *spindle) map[uint32]*template_object {
	the_map := make(map[uint32]*template_object, 8)

	err := filepath.WalkDir(template_path, func(path string, file fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if path == template_path {
			return nil
		}
		if file.IsDir() {
			return filepath.SkipDir
		}

		path = filepath.ToSlash(path)

		if is_draft(path) {
			return nil
		}

		name := filepath.Base(path)
		ext  := filepath.Ext(name)

		name = name[:len(name) - len(ext)]

		if x, ok := load_template(spindle, path); ok {
			the_map[new_hash(name)] = x
		}

		return nil
	})
	if err != nil {
		return nil
	}

	return the_map
}

func load_partial(spindle *spindle, full_path string) (*partial_object, bool) {
	blob, ok := load_file(full_path)

	if !ok {
		return nil, false
	}

	token_stream := lex_blob(full_path, blob)
	syntax_tree  := parse_stream(spindle.errors, token_stream, true)
	// print_syntax_tree(syntax_tree, 0)

	p := &partial_object{
		partial_path: full_path,
		content:      syntax_tree,
		raw_string:   blob,
	}

	return p, true
}

func load_all_partials(spindle *spindle) map[uint32]*partial_object {
	the_map := make(map[uint32]*partial_object, 8)

	err := filepath.WalkDir(partial_path, func(path string, file fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if path == partial_path {
			return nil
		}
		if file.IsDir() {
			return filepath.SkipDir
		}

		path = filepath.ToSlash(path)

		if is_draft(path) {
			return nil
		}

		name := filepath.Base(path)
		ext  := filepath.Ext(name)

		name = name[:len(name) - len(ext)]

		if x, ok := load_partial(spindle, path); ok {
			the_map[new_hash(name)] = x
		}

		return nil
	})
	if err != nil {
		return nil
	}

	return the_map
}