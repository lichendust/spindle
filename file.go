package main

import (
	"os"
	"io"
	"fmt"
	"time"
	"errors"
	"io/ioutil"
	"path/filepath"
)

type file struct {
	path    string
	draft   bool
}

func is_draft(input string) bool {
	if input[0] == '_' {
		return true
	}

	x := len(input) - 1

	for i, c := range input {
		if c == '/' && i < x {
			if input[i + 1] == '_' {
				return true
			}
		}
	}
	return false
}

func base_name(input string) string {
	input = filepath.Base(input)
	return input[:len(input) - len(filepath.Ext(input))]
}

func load_file(source_file string) (string, bool) {
	f, err := os.Open(source_file)

	if err != nil {
		return "", false
	}

	defer f.Close()

	bytes, err := ioutil.ReadAll(f)

	if err != nil {
		return "", false
	}

	str := string(bytes)

	cache_rtext[source_file] = str

	return str, true
}

func load_file_cache(source_file string) (string, bool) {
	if x, ok := cache_rtext[source_file]; ok {
		return x, true
	}
	return load_file(source_file)
}

func make_directory(path string) {
	os.MkdirAll(path, os.ModeDir)
}

func make_file(path, content string) {
	if err := ioutil.WriteFile(path, []byte(content), 0644); err != nil {
		panic(err)
	}
}

func copy_file(source_path, target_path string) {
	source, err := os.Open(source_path)

	if err != nil {
		panic(err)
	}

	defer source.Close()

	destination, err := os.OpenFile(target_path, os.O_CREATE, 0755)

	if err != nil {
		panic(err)
	}

	defer destination.Close()

	_, err = io.Copy(destination, source)

	if err != nil {
		panic(err)
	}
}

func delete_file(path string) {
	err := os.RemoveAll(path)

	if err != nil {
		panic(err)
	}
}

func make_blank_project() {
	make_directory("config/chunks")
	make_directory("config/plates")
	make_directory("source")
	make_file("config/config.x", config_template)
	make_file("source/index.x",  index_template)
	fmt.Println(new_project_message)
}

func is_dir(path string) bool {
	f, err := os.Open(path)

	if err != nil {
		return false
	}

	defer f.Close()

	info, err := f.Stat()

	if err != nil {
		return false
	}

	return info.IsDir()
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

func directory_has_changes(root_path string, last_run time.Time) bool {
	first := false
	has_changes := false

	filepath.Walk(root_path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			console_handler.print("error accessing", path) // @error
			return nil
		}

		if !first {
			first = true
			return nil
		}

		if info.ModTime().After(last_run) {
			has_changes = true
			return errors.New("")
		}

		return nil
	})

	return has_changes
}

func get_files(root_path string, reject_drafts bool) ([]*file, []*file) {
	files   := make([]*file, 0, 32)
	folders := make([]*file, 0, 16)

	err := filepath.Walk(root_path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		path = filepath.ToSlash(path)

		name := info.Name()
		fchr := name[0]

		draft := is_draft(path)

		if info.IsDir() {
			if path == root_path {
				return nil
			}

			if fchr == '.' || draft && reject_drafts {
				return filepath.SkipDir
			}

			folders = append(folders, &file {path, draft})
			return nil
		}

		if fchr == '.' || draft && reject_drafts {
			return nil
		}

		files = append(files, &file {path, draft})
		return nil
	})

	if err != nil {
		panic(err)
	}

	return files, folders
}