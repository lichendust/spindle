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

type Page struct {
	markup
	file         *File
	slug_tracker map[string]uint
	page_path    string
	import_cond  string
}

type Support_Markup struct {
	markup
	has_body bool
}

type Gen_Image struct {
	is_built bool
	original *File
	settings *image_settings
}

func load_page_from_file(spindle *spindle, o *File) (*Page, bool) {
	x, ok := load_page(spindle, o.path)
	if !ok {
		return nil, false
	}
	x.file = o
	return x, true
}

func load_page(spindle *spindle, full_path string) (*Page, bool) {
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

	anon_info := file_info{
		is_draft(full_path),
		full_path,
	}

	syntax_tree := parse_stream(spindle, &anon_info, token_stream, false)
	// print_syntax_tree(syntax_tree, 0)

	p := &Page{
		page_path:    full_path,
		slug_tracker: make(map[string]uint, 16),
	}

	p.content   = syntax_tree
	p.top_scope = arrange_top_scope(syntax_tree)
	p.position  = position{0,0,0,full_path}

	if !spindle.server_mode {
		spindle.pages[cache_path] = p
	}

	return p, true
}

func load_support(spindle *spindle, full_path string, support_type ast_type) (*Support_Markup, bool) {
	blob, ok := load_file(full_path)
	if !ok {
		return nil, false
	}

	token_stream := lex_blob(full_path, blob)
	// print_token_stream(token_stream)

	anon_info := file_info{
		is_draft(full_path),
		full_path,
	}

	syntax_tree := parse_stream(spindle, &anon_info, token_stream, true)
	// print_syntax_tree(syntax_tree, 0)

	t := Support_Markup{}

	t.content  = syntax_tree
	t.position = position{0,0,0,full_path}

	// we don't bother with this for partials
	if support_type == TEMPLATE {
		t.has_body  = recursive_anon_count(syntax_tree) > 0
		t.top_scope = arrange_top_scope(syntax_tree)
	}

	return &t, true
}

func load_support_directory(spindle *spindle, support_type ast_type, root_path string) map[uint32]*Support_Markup {
	the_map := make(map[uint32]*Support_Markup, 8)

	err := filepath.WalkDir(root_path, func(path string, file fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if path == root_path {
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

		if x, ok := load_support(spindle, path, support_type); ok {
			the_map[new_hash(name)] = x
		}

		return nil
	})
	if err != nil {
		return nil
	}

	return the_map
}