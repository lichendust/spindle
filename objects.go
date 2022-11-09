package main

import (
	"io/fs"
	"path/filepath"
)

type markup struct {
	content   []ast_data
	top_scope []ast_data
	position  position
}

type page_object struct {
	markup
	// page_id      uint32
	slug_tracker map[string]uint
	page_path    string
}

type template_object struct {
	markup
	has_body      bool
	template_path string
}

type partial_object struct {
	markup
	partial_path string
}

type gen_image struct {
	is_built bool
	original *disk_object
	settings *image_settings
}

type gen_page struct {
	is_built bool
	original *disk_object
}

func load_page(spindle *spindle, full_path string) (*page_object, bool) {
	cache_path := full_path[:len(full_path) - len(filepath.Ext(full_path))]

	if !spindle.server_mode {
		if p, ok := spindle.pages[cache_path]; ok {
			return p, true
		}
	}

	blob, ok := load_file(full_path)

	if !ok {
		return nil, false
	}

	token_stream := lex_blob(full_path, blob)
	// print_token_stream(token_stream)

	syntax_tree := parse_stream(spindle.errors, token_stream, false)
	// print_syntax_tree(syntax_tree, 0)

	p := &page_object{
		page_path:    full_path,
		slug_tracker: make(map[string]uint, 4),
	}

	p.content   = syntax_tree
	p.top_scope = arrange_top_scope(syntax_tree)
	p.position  = position{0,0,0,full_path}

	if !spindle.server_mode {
		spindle.pages[cache_path] = p
	}

	return p, true
}

func load_template(spindle *spindle, full_path string) (*template_object, bool) {
	blob, ok := load_file(full_path)

	if !ok {
		return nil, false
	}

	token_stream := lex_blob(full_path, blob)
	// print_token_stream(token_stream)

	syntax_tree := parse_stream(spindle.errors, token_stream, true)
	// print_syntax_tree(syntax_tree, 0)

	t := &template_object{
		has_body:      recursive_anon_count(syntax_tree) > 0,
		template_path: full_path,
	}

	t.content   = syntax_tree
	t.top_scope = arrange_top_scope(syntax_tree)
	t.position  = position{0,0,0,full_path}

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
	// print_token_stream(token_stream)

	syntax_tree  := parse_stream(spindle.errors, token_stream, true)
	// print_syntax_tree(syntax_tree, 0)

	p := &partial_object{
		partial_path: full_path,
	}

	p.content = syntax_tree

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