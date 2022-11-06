package main

import (
	"os"
	"fmt"
	"path/filepath"
)

func command_build(spindle *spindle) {
	if data, ok := load_file_tree(spindle); ok {
		spindle.file_tree = data
	}

	spindle.templates    = load_all_templates(spindle)
	spindle.partials     = load_all_partials(spindle)

	spindle.pages        = make(map[string]*page_object, 64)
	spindle.finder_cache = make(map[string]*disk_object, 64)
	spindle.gen_images   = make(map[uint32]*gen_image, 32)
	spindle.gen_pages    = make(map[string]*gen_page, 32)

	make_dir(public_path)

	found_file, ok := find_file(spindle.file_tree, "index")

	if !ok {
		panic("failed to find index")
	}

	found_file.is_used = true

	for {
		done := build_pages(spindle, spindle.file_tree)

		if spindle.errors.has_failures {
			break
		}
		if done {
			break
		}
	}

	for _, image := range spindle.gen_images {
		if image.original.is_draft && !spindle.config.build_drafts {
			continue
		}

		output_path := make_generated_image_path(spindle, image)
		make_dir(filepath.Dir(output_path))

		ok := copy_generated_image(image, output_path)
		if !ok {
			spindle.errors.new(FAILURE, "%q could failed to be generated", output_path)
		}

		image.is_built = true
	}

	if spindle.errors.has_errors() {
		fmt.Fprintln(os.Stderr, spindle.errors.render_term_errors())
	}
}

func build_pages(spindle *spindle, file *disk_object) bool {
	is_done := true

	main_loop: for _, file := range file.children {
		if file.file_type == DIRECTORY {
			done := build_pages(spindle, file)
			if !done {
				is_done = false
			}
			continue
		}

		if file.is_built {
			continue
		}
		if file.is_draft && !spindle.config.build_drafts {
			continue
		}
		if spindle.config.build_only_used && !file.is_used {
			continue
		}

		is_done = false
		file.is_built = true

		output_path := make_general_file_path(spindle, file)
		make_dir(filepath.Dir(output_path))

		switch file.file_type {
		case MARKUP:
			page, ok := load_page(spindle, file.path)
			if !ok {
				panic("failed to load page " + file.path)
			}

			assembled := render_syntax_tree(spindle, page, 0)

			if !write_file(output_path, assembled) {
				spindle.errors.new(FAILURE, "%q could not be written to disk", output_path)
				break main_loop
			}

		case SCSS:
			copy_scss(file, output_path)

		case CSS, JAVASCRIPT:
			copy_minify(file, output_path)

		default:
			copy_file(file, output_path)
		}
	}

	return is_done
}

/*func format_index(input string) string {
	if !strings.Contains(input, "index") {
		return filepath.Base(input)
	}
	return input
}*/