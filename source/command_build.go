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

import "os/exec"
import "path/filepath"

func command_build() {
	if data, ok := load_file_tree(); ok {
		spindle.file_tree = data
	}

	spindle.templates = load_support_directory(TEMPLATE, TEMPLATE_PATH)
	spindle.partials  = load_support_directory(PARTIAL,  PARTIAL_PATH)

	// if the user has requested any webp
	// output in the templates
	if spindle.has_webp {
		_, err := exec.LookPath("cwebp")
		if err != nil {
			panic("cwebp not found in path") // @error
			return
		}
	}

	spindle.finder_cache = make(map[string]*File, 64)

	spindle.pages      = make(map[string]*Page, 64)
	spindle.gen_pages  = make(map[string]*Gen_Page, 32)
	spindle.gen_images = make(map[uint32]*Image, 32)

	make_dir(spindle.output_path)

	if found_file, ok := find_file(spindle.file_tree, "index"); ok {
		found_file.is_used = true
	} else {
		panic("need a root index!")
	}
	if found_file, ok := find_file(spindle.file_tree, "favicon.ico"); ok {
		found_file.is_used = true
	}

	for {
		done := build_pages(spindle.file_tree)

		if spindle.errors.has_failures {
			break
		}
		if done {
			break
		}
	}

	for _, gen := range spindle.gen_pages {
		output_path := make_general_file_path(gen.file)

		// @todo tag path can't distinguish between a file extension
		// and the TLD of an index page on a domain.  This is the only
		// place it actually matters, so we just trim and replace it
		{
			ext := filepath.Ext(output_path)
			output_path = output_path[:len(output_path) - len(ext)]
			output_path = tag_path(output_path, spindle.tag_path, gen.import_cond) + ext
		}

		make_dir(filepath.Dir(output_path))

		page, ok := load_page_from_file(gen.file)
		if !ok {
			panic("failed to load page " + gen.file.path)
		}

		page.file        = gen.file
		page.import_cond = gen.import_cond
		page.import_hash = gen.import_hash

		assembled := render_syntax_tree(page)

		if !write_file(output_path, assembled) {
			spindle.errors.new(FAILURE, "%q could not be written to disk", output_path)
			break
		}
	}

	if !spindle.skip_images {
		for _, image := range spindle.gen_images {
			if image.original.is_draft && !spindle.build_drafts {
				continue
			}

			output_path := make_generated_image_path(image)
			make_dir(filepath.Dir(output_path))

			ok := copy_generated_image(image, output_path)
			if !ok {
				spindle.errors.new(FAILURE, "%q could failed to be generated", output_path)
			}

			image.is_built = true
		}
	}

	if spindle.errors.has_errors() {
		eprintln(spindle.errors.render_errors(ERR_TERM))
		eprint("\n")
	}

	if spindle.sitemap {
		sitemap()
	}
}

func build_pages(file *File) bool {
	is_done := true

	main_loop: for _, file := range file.children {
		if file.file_type == DIRECTORY {
			done := build_pages(file)
			if !done {
				is_done = false
			}
			continue
		}

		if file.is_built {
			continue
		}
		if file.is_draft && !spindle.build_drafts {
			continue
		}
		if spindle.build_only_used && !file.is_used {
			continue
		}

		is_done = false
		file.is_built = true

		output_path := make_general_file_path(file)
		make_dir(filepath.Dir(output_path))

		switch file.file_type {
		case MARKUP:
			page, ok := load_page_from_file(file)
			if !ok {
				panic("failed to load page " + file.path)
			}

			page.file = file // @todo put in load_page

			assembled := render_syntax_tree(page)

			if !write_file(output_path, assembled) {
				spindle.errors.new(FAILURE, "%q could not be written to disk", output_path)
				break main_loop
			}

		case CSS:
			if track_css_links(file.path) {
				is_done = false
			}
			fallthrough // css gets minified below too

		case JAVASCRIPT:
			copy_minify(file, output_path)

		default:
			copy_file(file, output_path)
		}
	}

	return is_done
}