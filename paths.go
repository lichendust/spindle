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
		output_path = rewrite_root(file.path, spindle.config.output_path)
	} else {
		output_path = rewrite_ext(rewrite_root(file.path, spindle.config.output_path), new_ext)
	}

	if file.is_draft {
		output_path = undraft_path(output_path)
	}

	return output_path
}

func make_generated_image_path(spindle *spindle, the_image *generated_image) string {
	public_path := spindle.config.output_path
	file_path   := the_image.original.path
	s           := the_image.settings

	new_ext := ext_for_file_type(s.file_type)
	hash    := s.make_hash()

	// @todo reject additional name if only ext is changed?

	return rewrite_ext(rewrite_root(file_path, public_path), fmt.Sprintf("_%d%s", hash, new_ext))
}

func make_general_url(spindle *spindle, file *disk_object, path_type path_type, current_location string) string {
	if spindle.server_mode {
		return rewrite_by_path_type(ROOTED, spindle.config.domain, current_location, file.path)
	}

	output_path := rewrite_by_path_type(path_type, spindle.config.domain, current_location, file.path)
	output_path = rewrite_ext(output_path, ext_for_file_type(file.file_type))

	if file.is_draft {
		output_path = undraft_path(output_path)
	}
	return output_path
}

func make_page_url(spindle *spindle, file *disk_object, path_type path_type, current_location string) string {
	if spindle.server_mode {
		return rewrite_by_path_type(ROOTED, spindle.config.domain, current_location, file.path)
	}

	output_path := rewrite_by_path_type(path_type, spindle.config.domain, current_location, file.path)
	if file.is_draft {
		output_path = undraft_path(output_path)
	}
	output_path = rewrite_ext(output_path, "")
	return output_path
}

func make_generated_image_url(spindle *spindle, file *disk_object, s *image_settings, path_type path_type, current_location string) string {
	if spindle.server_mode {
		return rewrite_by_path_type(ROOTED, spindle.config.domain, current_location, file.path)
	}

	output_path := rewrite_by_path_type(path_type, spindle.config.domain, current_location, file.path)
	new_ext := ext_for_file_type(s.file_type)
	hash    := s.make_hash()

	return rewrite_ext(output_path, fmt.Sprintf("_%d%s", hash, new_ext))
}

/*func make_public_file_path(spindle *spindle, file *disk_object) string {
	output_path := ""

	new_ext := ext_for_file_type(file.file_type)
	if new_ext == "" {
		output_path = rewrite_root(file.path, public_path)
	} else {
		output_path = rewrite_ext(rewrite_root(file.path, public_path), new_ext)
	}

	return output_path
}

func make_public_url(spindle *spindle, file *disk_object, path_type path_type, current_location string) string {
	path := make_public_file_path(spindle, file)

	if spindle.server_mode {
		path_type = ROOTED
	} else {
		if is_draft(path) {
			path = undraft_path(path)
		}
		if path_type == NO_PATH_TYPE {
			path_type = spindle.config.default_path_mode
		}
	}

	return rewrite_by_path_type(path_type, spindle.config.domain, current_location, path)
}*/

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