package main

import (
	"io/fs"
	"path/filepath"
)

type Markup struct {
	content   []AST_Data
	top_scope []AST_Data
	position  position
}

type Page struct {
	Markup
	file         *File
	page_path    string
	import_hash  uint32
	import_cond  string
}

type Gen_Page struct {
	file *File
	import_hash uint32
	import_cond string
}

type Support_Markup struct {
	Markup
	has_body bool
}

// Image is only used when an image is *modified*
// some images are just straight copies and we bypass
// this representation and treat them as regular statics
type Image struct {
	is_built bool
	original *File
	settings *Image_Settings
}

func load_page_from_file(spindle *Spindle, o *File) (*Page, bool) {
	x, ok := load_page(spindle, o.path)
	if !ok {
		return nil, false
	}
	x.file = o
	return x, true
}

func load_page(spindle *Spindle, full_path string) (*Page, bool) {
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

	anon_info := File_Info{
		is_draft(full_path),
		full_path,
	}

	syntax_tree := parse_stream(spindle, &anon_info, token_stream, false)
	// print_syntax_tree(syntax_tree, 0)

	p := new(Page)

	p.page_path = full_path
	p.content   = syntax_tree
	p.top_scope = arrange_top_scope(syntax_tree)
	p.position  = position{0,0,0,full_path}

	if !spindle.server_mode {
		spindle.pages[cache_path] = p
	}

	return p, true
}

func load_support(spindle *Spindle, full_path string, support_type AST_Type) (*Support_Markup, bool) {
	blob, ok := load_file(full_path)
	if !ok {
		return nil, false
	}

	token_stream := lex_blob(full_path, blob)
	// print_token_stream(token_stream)

	anon_info := File_Info{
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

func load_support_directory(spindle *Spindle, support_type AST_Type, root_path string) map[uint32]*Support_Markup {
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