package main

import (
	"io/fs"
	"path/filepath"
)

type page_object struct {
	page_id      uint32
	page_path    string
	content      []ast_data
	top_scope    []*ast_declare
	slug_tracker map[string]uint
}

type template_object struct {
	content   []ast_data
	top_scope []*ast_declare
}

func load_page(full_path string) (*page_object, bool) {
	blob, ok := load_file(full_path)

	if !ok {
		return nil, false
	}

	token_stream := lex_blob(blob)
	// print_token_stream(token_stream)

	syntax_tree := parse_stream(token_stream)
	// print_syntax_tree(syntax_tree, 0)

	array := make([]*ast_declare, 0, 8)

	for _, entry := range syntax_tree {
		if entry.type_check().is(DECL, DECL_TOKEN, DECL_TOKEN) {
			entry := entry.(*ast_declare)
			array = append(array, entry)
		}
	}

	return &page_object{
		page_path:    full_path,
		content:      syntax_tree,
		slug_tracker: make(map[string]uint, 4),
	}, true
}

func load_template(full_path string) (*template_object, bool) {
	blob, ok := load_file(full_path)

	if !ok {
		return nil, false
	}

	token_stream := lex_blob(blob)
	syntax_tree  := parse_stream(token_stream)

	array := make([]*ast_declare, 0, 8)

	for _, entry := range syntax_tree {
		if entry.type_check().is(DECL, DECL_TOKEN, DECL_BLOCK) {
			array = append(array, entry.(*ast_declare))
		}
	}

	return &template_object{
		content:   syntax_tree,
		top_scope: array,
	}, true
}

func load_all_templates() map[uint32]*template_object {
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

		name := filepath.Base(path)
		ext  := filepath.Ext(name)

		name = name[:len(name) - len(ext)]

		if x, ok := load_template(path); ok {
			the_map[new_hash(name)] = x
		}

		return nil
	})

	if err != nil {
		return nil
	}

	return the_map
}