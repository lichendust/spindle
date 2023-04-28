package main

import (
	"os"
	"io"
	"io/fs"
	"time"
	"path/filepath"
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

	// end_process_image

	IMG_GIF
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

func to_file_type(input string) file_type {
	switch filepath.Ext(input) {
	case extension:
		return MARKUP
	case ".md":
		return MARKDOWN
	case ".html":
		return HTML
	case ".css":
		return CSS
	case ".scss":
		return SCSS
	case ".js":
		return JAVASCRIPT
	case ".png":
		return IMG_PNG
	case ".jpg", ".jpeg":
		return IMG_JPG
	case ".tif", ".tiff":
		return IMG_TIF
	case ".webp":
		return IMG_WEB
	case ".gif":
		return IMG_GIF
	}
	return STATIC
}

func ext_for_file_type(file_type file_type) string {
	switch file_type {
	case MARKUP:
		return ".html"
	case MARKDOWN:
		return ".html"
	case HTML:
		return ".html"
	case CSS:
		return ".css"
	case SCSS:
		return ".css"
	case JAVASCRIPT:
		return ".js"
	case IMG_PNG:
		return ".png"
	case IMG_JPG:
		return ".jpg"
	case IMG_TIF:
		return ".tif"
	case IMG_WEB:
		return ".webp"
	}
	return ""
}

type anon_file_info struct {
	is_draft bool
	path     string
}

type disk_object struct {
	anon_file_info
	file_type file_type
	hash_name uint32
	hash_url  uint32
	is_used   bool
	is_built  bool
	parent    *disk_object
	children  []*disk_object
}

func new_file_tree() []*disk_object {
	return make([]*disk_object, 0, 32)
}

func load_file_tree(spindle *spindle) (*disk_object, bool) {
	f := &disk_object{
		file_type: ROOT,
		is_used:   true,
	}

	f.path = source_path

	children, ok := recurse_directories(spindle, f)
	if !ok {
		return nil, false
	}

	f.children = children

	return f, true
}

func hash_base_name(file *disk_object) uint32 {
	base := filepath.Base(file.path)

	if x := file.file_type; x > is_page && x < end_page {
		base = base[:len(base) - len(filepath.Ext(base))]
	}

	return new_hash(base)
}

func recurse_directories(spindle *spindle, parent *disk_object) ([]*disk_object, bool) {
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
				is_used:   false,
			}

			the_file.path     = path
			the_file.is_draft = is_draft(path)

			the_file.hash_name = hash_base_name(the_file)
			the_file.parent    = parent

			if x, ok := recurse_directories(spindle, the_file); ok {
				the_file.children = x
			}

			array = append(array, the_file)
			return filepath.SkipDir
		}

		the_file := &disk_object{
			file_type: STATIC,
			is_used:   false,
		}

		the_file.path     = path
		the_file.is_draft = is_draft(path)

		the_file.file_type = to_file_type(path)
		the_file.hash_name = hash_base_name(the_file)

		the_file.parent = parent

		if x := the_file.file_type; x > is_page && x < end_page {
			the_file.hash_url = new_hash(make_page_url(spindle, &the_file.anon_file_info, ROOTED, ""))
		} else {
			the_file.hash_url = new_hash(make_general_url(spindle, the_file, ROOTED, ""))
		}

		array = append(array, the_file)
		return nil
	})
	if err != nil {
		return nil, false
	}

	return array, true
}

func find_file_hash(start_location *disk_object, target uint32) (*disk_object, bool) {
	for _, entry := range start_location.children {
		if entry.file_type == DIRECTORY {
			if x, ok := find_file_hash(entry, target); ok {
				return x, true
			}
			continue
		}
		if entry.hash_url == target {
			return entry, true
		}
	}

	return nil, false
}

func find_depthless_file(entry *disk_object, target string) (*disk_object, bool) {
	check := entry.path

	if x := entry.file_type; x > is_page && x < end_page {
		check = check[:len(check) - len(filepath.Ext(check))]
	}

	diff := len(check) - len(target)
	if diff <= 0 {
		return nil, false
	}

	leven := levenshtein_distance(check, target)
	if leven <= diff {
		b_target := filepath.Base(target)
		b_check  := filepath.Base(check)

		if len(b_target) != len(b_check) || b_target[0] != b_check[0] {
			return nil, false
		}

		if entry.file_type == DIRECTORY {
			// if directory only has one friend
			// send that back
			if len(entry.children) == 1 {
				return entry.children[0], true
			}

			// else look for "index"
			for _, child := range entry.children {
				if child.hash_name == index_hash {
					return child, true
				}
			}
			return nil, false
		}

		return entry, true
	}

	return nil, false
}

func find_file(start_location *disk_object, target string) (*disk_object, bool) {
	for _, entry := range start_location.children {
		if x, ok := find_depthless_file(entry, target); ok {
			return x, true
		}
	}

	for _, entry := range start_location.children {
		if entry.file_type == DIRECTORY {
			if x, ok := find_file(entry, target); ok {
				return x, true
			}
		}
	}

	return nil, false
}

/*func file_has_changes(path string, last_run time.Time) bool {
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
}*/

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
			// random error that Walk would never return to exit early
			// why they don't just provide one like SkipDir is beyond me
			return io.EOF
		}

		return nil
	})
	if err != io.EOF {
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
	return err == nil
}

func make_dir(path string) bool {
	err := os.MkdirAll(path, os.ModeDir | 0777)
	return err == nil
}

func copy_file(file *disk_object, output_path string) {
	source, err := os.Open(file.path)
	if err != nil {
		panic(err) // @error
	}
	defer source.Close()

	destination, err := os.OpenFile(output_path, os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		panic(err)
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	if err != nil {
		panic(err)
	}
}

/*const last_build = "config/.last_build"

func read_time() time.Time {
	content, err := os.ReadFile(last_build)

	if err != nil {
		return time.Unix(0, 0)
	}

	i, err := strconv.ParseInt(string(content), 10, 64)

	if err != nil {
		return time.Unix(0, 0)
	}

	return time.Unix(i, 0)
}

func save_time() {
	the_time := []byte(strconv.FormatInt(time.Now().Unix(), 10))

	if err := os.WriteFile(last_build, the_time, 0777); err != nil {
		panic(err)
	}
}*/