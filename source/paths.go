package main

import (
	"fmt"
	"strings"
	"unicode"
	"path/filepath"
)

func make_general_file_path(spindle *spindle, file *disk_object) string {
	output_path := ""

	new_ext := ext_for_file_type(file.file_type)
	if new_ext == "" {
		output_path = rewrite_root(file.path, spindle.output_path)
	} else {
		output_path = rewrite_ext(rewrite_root(file.path, spindle.output_path), new_ext)
	}

	if file.is_draft {
		output_path = undraft_path(output_path)
	}

	return output_path
}

func make_generated_image_path(spindle *spindle, the_image *gen_image) string {
	public_path := spindle.output_path
	file_path   := the_image.original.path
	s           := the_image.settings

	new_ext := ext_for_file_type(s.file_type)
	hash    := s.make_hash()

	// @todo reject additional name if only ext is changed?

	return rewrite_ext(rewrite_root(file_path, public_path), fmt.Sprintf("_%d%s", hash, new_ext))
}

func make_general_url(spindle *spindle, file *disk_object, path_type path_type, current_location string) string {
	if spindle.server_mode {
		return rewrite_by_path_type(ROOTED, spindle.domain, current_location, file.path)
	}

	output_path := rewrite_by_path_type(path_type, spindle.domain, current_location, file.path)
	output_path = rewrite_ext(output_path, ext_for_file_type(file.file_type))

	if spindle.build_drafts && file.is_draft {
		output_path = undraft_path(output_path)
	}

	return output_path
}

func make_page_url(spindle *spindle, file *anon_file_info, path_type path_type, current_location string) string {
	if spindle.server_mode {
		return _make_page_url(spindle, ROOTED, file.is_draft, file.path, current_location)
	}

	return _make_page_url(spindle, path_type, file.is_draft, file.path, current_location)
}

func _make_page_url(spindle *spindle, path_type path_type, is_draft bool, path, current_location string) string {
	output_path := rewrite_by_path_type(path_type, spindle.domain, current_location, path)

	if spindle.build_drafts && is_draft {
		output_path = undraft_path(output_path)
	}

	output_path = rewrite_ext(output_path, "")

	if filepath.Base(output_path) == "index" {
		if len(output_path) == 6 {
			output_path = "/"
		} else if len(output_path) < 6 {

		} else {
			output_path = output_path[:len(output_path) - 6]

			if output_path == ".." {
				output_path = "../"
			}
		}
	}

	return output_path
}

func make_generated_image_url(spindle *spindle, file *disk_object, s *image_settings, path_type path_type, current_location string) string {
	if spindle.server_mode {
		return rewrite_by_path_type(ROOTED, spindle.domain, current_location, file.path)
	}

	output_path := rewrite_by_path_type(path_type, spindle.domain, current_location, file.path)
	new_ext     := ext_for_file_type(s.file_type)
	hash        := s.make_hash()

	return rewrite_ext(output_path, fmt.Sprintf("_%d%s", hash, new_ext))
}

func rewrite_by_path_type(path_type path_type, domain, current_location, target_location string) string {
	buffer := strings.Builder{}
	buffer.Grow(64)

	switch path_type {
	case ROOTED:
		buffer.WriteRune('/')
		buffer.WriteString(rewrite_root(target_location, ""))

	case RELATIVE:
		if path, ok := filepath_relative(current_location, target_location); ok {
			buffer.WriteString(path)
		}

	case ABSOLUTE:
		buffer.WriteString(domain)
		buffer.WriteString(rewrite_root(target_location, ""))
	}

	return buffer.String()
}

func undraft_path(input string) string {
	chunks := strings.Split(input, "/")

	for i, chunk := range chunks {
		for ci, c := range chunk {
			if c != '_' {
				chunks[i] = chunk[ci:]
				break
			}
		}
	}

	return strings.Join(chunks, "/")
}

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

func rewrite_ext(target, new_ext string) string {
	target = target[:len(target) - len(filepath.Ext(target))]
	return target + new_ext
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

func is_draft(input string) bool {
	if input[0] == '_' {
		return true
	}

	x := len(input) - 1

	for i, c := range input {
		if c == '/' && i < x && input[i + 1] == '_' {
			return true
		}
	}

	return false
}