/*
	Spindle
	A static site generator
	Copyright (C) 2022-2023 Harley Denham

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package main

import (
	"fmt"
	"net/url"
	"strings"
	"unicode"
	"unicode/utf8"
	"path/filepath"
)

func make_general_file_path(file *File) string {
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

func make_generated_image_path(the_image *Image) string {
	public_path := spindle.output_path
	file_path   := the_image.original.path
	s           := the_image.settings

	new_ext := ext_for_file_type(s.file_type)

	// tack a name on if the file has changed size *unless* its
	// the default settings for the project
	if s.max_size > 0 && s.max_size != spindle.image_max_size {
		new_ext = fmt.Sprintf("_%d%s", s.max_size, new_ext)
	}

	return rewrite_ext(rewrite_root(file_path, public_path), new_ext)
}

func make_general_url(file *File, Path_Type Path_Type, current_location string) string {
	if spindle.server_mode {
		return rewrite_by_path_type(ROOTED, spindle.domain, current_location, file.path)
	}

	output_path := rewrite_by_path_type(Path_Type, spindle.domain, current_location, file.path)
	output_path = rewrite_ext(output_path, ext_for_file_type(file.file_type))

	if spindle.build_drafts && file.is_draft {
		output_path = undraft_path(output_path)
	}

	return output_path
}

func make_page_url(file *File, Path_Type Path_Type, current_location string) string {
	if spindle.server_mode {
		return _make_page_url(ROOTED, file.is_draft, file.path, current_location)
	}
	return _make_page_url(Path_Type, file.is_draft, file.path, current_location)
}

func _make_page_url(Path_Type Path_Type, is_draft bool, path, current_location string) string {
	output_path := rewrite_by_path_type(Path_Type, spindle.domain, current_location, path)

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

func make_generated_image_url(file *File, s *Image_Settings, Path_Type Path_Type, current_location string) string {
	if spindle.server_mode {
		return rewrite_by_path_type(ROOTED, spindle.domain, current_location, file.path)
	}

	output_path := rewrite_by_path_type(Path_Type, spindle.domain, current_location, file.path)
	new_ext     := ext_for_file_type(s.file_type)

	if s.max_size > 0 && s.max_size != spindle.image_max_size {
		new_ext = fmt.Sprintf("_%d%s", s.max_size, new_ext)
	}

	return rewrite_ext(output_path, new_ext)
}

func rewrite_by_path_type(Path_Type Path_Type, domain, current_location, target_location string) string {
	buffer := strings.Builder{}
	buffer.Grow(64)

	switch Path_Type {
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

func get_protocol(input string) (string, string) {
	for i, c := range input {
		if c == ':' {
			r, _ := utf8.DecodeRuneInString(input[i + 1:])
			if r == '/' {
				return input[:i + 3], input[i + 3:]
			}
		}
	}
	return "", input
}

func tag_path(input, sep, tag string) string {
	protocol, path := get_protocol(input)

	path = strings.TrimSuffix(path, "index")

	if protocol != "" {
		x, err := url.JoinPath(protocol, path, sep, tag)
		if err != nil {
			eprintf("failed to assemble URL from %q\n", input)
		}
		path = x

	} else {
		path = filepath.ToSlash(filepath.Join(path, sep, tag))
	}

	return path
}