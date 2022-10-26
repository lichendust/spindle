package main

import (
	"os"
	"time"
	"io/fs"
	"unicode"
	"path/filepath"
)

const (
	extension   = ".x"

	source_path = "source"
	public_path = "public"
	config_path = "config"

	template_path = config_path + "/templates"
	partial_path  = config_path + "/partials"
	script_path   = config_path + "/scripts"
)

type file_type uint8
const (
	DIRECTORY file_type = iota
	ROOT

	is_image
	IMG_JPG
	IMG_PNG
	IMG_TIF
	IMG_WEB
	end_image

	is_page
	MARKUP
	MARKDOWN

	is_static
	HTML
	end_page

	STATIC
	JAVASCRIPT
	SCSS
	CSS
	end_static
)

type disk_object struct {
	file_type file_type
	hash_name uint32
	is_used   bool
	is_draft  bool
	parent    *disk_object
	children  []*disk_object
	path      string
}

func get_template_path(name string) string {
	return filepath.Join(template_path, name) + extension
}

func get_partial_path(name string) string {
	return filepath.Join(partial_path, name) + extension
}

func get_script_path(name string) string {
	return filepath.Join(script_path, name) + extension
}

func new_file_tree() []*disk_object {
	return make([]*disk_object, 0, 32)
}

func load_file_tree() (*disk_object, bool) {
	f := &disk_object{
		file_type: ROOT,
		is_used:   true,
		path:      source_path,
	}

	x, ok := recurse_directories(f)

	if !ok {
		return nil, false
	}

	f.children = x

	return f, true
}

func recurse_directories(parent *disk_object) ([]*disk_object, bool) {
	array := new_file_tree()

	err := filepath.WalkDir(parent.path, func(path string, file fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if path == parent.path {
			return nil
		}

		path = filepath.ToSlash(path)

		if file.IsDir() {
			the_file := &disk_object{
				file_type: DIRECTORY,
				hash_name: new_hash(filepath.Base(path)),
				is_used:   false,
				is_draft:  is_draft(path),
				path:      path,
				parent:    parent,
			}

			if x, ok := recurse_directories(the_file); ok {
				the_file.children = x
			}

			array = append(array, the_file)
			return filepath.SkipDir
		}

		the_file := &disk_object{
			file_type: STATIC,
			is_used:   false,
			is_draft:  is_draft(path),
			path:      path,
			parent:    parent,
		}

		ext := filepath.Ext(path)

		switch ext {
		case ".x":
			the_file.file_type = MARKUP
		case ".md":
			the_file.file_type = MARKDOWN
		case ".html":
			the_file.file_type = HTML
		case ".css":
			the_file.file_type = CSS
		case ".scss":
			the_file.file_type = SCSS
		case ".js":
			the_file.file_type = JAVASCRIPT
		case ".png":
			the_file.file_type = IMG_PNG
		case ".jpg", ".jpeg":
			the_file.file_type = IMG_JPG
		case ".tif", ".tiff":
			the_file.file_type = IMG_TIF
		case ".webp":
			the_file.file_type = IMG_WEB
		}

		the_file.hash_name = new_hash(filepath.Base(path))

		array = append(array, the_file)
		return nil
	})
	if err != nil {
		return nil, false
	}

	return array, true
}

func find_file(file_tree *disk_object, start_location *disk_object, path string) (*disk_object, bool) {
	if path[0] == '/' {
		if x, ok := find_file_descending(file_tree, path[1:]); ok {
			return x, true
		}
	}

	return find_file_ascending(start_location, path)
}

func find_file_ascending(start_location *disk_object, path string) (*disk_object, bool) {
	for _, entry := range start_location.children {
		if entry.file_type == DIRECTORY {
			continue
		}

		diff := len(entry.path) - len(path)
		if diff < 0 {
			continue
		}

		if diff == 0 && entry.path == path {
			return entry, true
		}

		leven := levenshtein_distance(entry.path, path)
		if leven <= diff {
			return entry, true
		}
	}

	for _, entry := range start_location.children {
		if entry.file_type == DIRECTORY {
			if x, ok := find_file_descending(entry, path); ok {
				return x, true
			}
		}
	}

	if start_location.parent == nil {
		return nil, false
	}

	return find_file_ascending(start_location.parent, path)
}

func find_file_descending(start_location *disk_object, path string) (*disk_object, bool) {
	for _, entry := range start_location.children {
		if entry.file_type == DIRECTORY {
			continue
		}

		diff := len(entry.path) - len(path)
		if diff < 0 {
			continue
		}

		leven := levenshtein_distance(entry.path, path)
		if leven <= diff {
			return entry, true
		}
	}

	for _, entry := range start_location.children {
		if entry.file_type == DIRECTORY {
			if x, ok := find_file_descending(entry, path); ok {
				return x, true
			}
		}
	}

	return nil, false
}

func new_conf_list() map[uint32]*disk_object {
	return make(map[uint32]*disk_object, 32)
}

func load_conf_list() (map[uint32]*disk_object, bool) {
	the_map := new_conf_list()

	err := filepath.WalkDir(config_path, func(path string, file fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if file.IsDir() {
			return filepath.SkipDir
		}

		_println(file.Name())

		return nil
	})
	if err != nil {
		return nil, false
	}

	return the_map, true
}

func file_has_changes(path string, last_run time.Time) bool {
	f, err := os.Open(path)
	if err != nil {
		panic(path)
	}

	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		panic(path)
	}

	if info.ModTime().After(last_run) {
		return true
	}

	return false
}

func folder_has_changes(root_path string, last_run time.Time) bool {
	first := false
	has_changes := false

	err := filepath.Walk(root_path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if !first {
			first = true
			return nil
		}

		if info.ModTime().After(last_run) {
			has_changes = true
		}

		return nil
	})
	if err != nil {
		return false
	}

	return has_changes
}

func load_file(source_file string) (string, bool) {
	content, err := os.ReadFile(source_file)
	if err != nil {
		return "", false
	}

	return string(content), true
}

func write_file(path, content string) bool {
	err := os.WriteFile(path, []byte(content), 0777)
	if err != nil {
		return false
	}

	return true
}

func make_dir(path string) bool {
	err := os.MkdirAll(path, os.ModeDir)
	if err != nil {
		return false
	}

	return true
}

// path tools
func is_draft(input string) bool {
	if input[0] == '_' {
		return true
	}

	x := len(input) - 1

	for i, c := range input {
		if (c == '/' || c == '\\') && i < x && input[i + 1] == '_' {
			return true
		}
	}

	return false
}

// removes the root of the path, replacing it with new_root
func rewrite_root(target, new_root string) string {
	if target[0] == '/' {
		target = target[1:]
	}

	for i, c := range target {
		if c == '/' {
			target = target[i + 1:]
			break
		}
	}

	return filepath.ToSlash(filepath.Join(new_root, target))
}

// replaces the extension of the file
func rewrite_ext(target, new_ext string) string {
	target = target[:len(target) - len(filepath.Ext(target))]
	return target + new_ext
}

// shorthand path rewriter for a large number of output files
func rewrite_public(target, new_ext string) string {
	return rewrite_ext(rewrite_root(target, public_path), new_ext)
}

func filepath_relative(a, b string) (string, bool) {
	a = a[:len(a) - len(filepath.Base(a))]

	path, err := filepath.Rel(a, b)

	if err != nil {
		return "", false
	}
	return filepath.ToSlash(path), true
}

// validates user input to make sure you're not typing something insane
func is_valid_path(input string) bool {
	for _, c := range input {
		if !(unicode.IsLetter(c) || unicode.IsNumber(c) || c == '.' || c == '/' || c == '\\') {
			return false
		}
	}

	return true
}
