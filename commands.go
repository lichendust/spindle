package main

import (
	"os"
	"fmt"
	"strings"
	"path/filepath"
)

const (
	VERSION uint8 = iota
	HELP
	INIT
	BUILD
	SERVE
)

func command_init(config *config) {
	if config.path != "" {
		make_dir(config.path)
		os.Chdir(config.path)
	}

	make_dir(template_path)
	make_dir(partial_path)
	make_dir(script_path)
	make_dir(source_path)

	write_file(filepath.Join(config_path, "main.x"),  main_template)
	write_file(filepath.Join(source_path, "index.x"), index_template)
}

func command_build(config *config) {
	spindle := &spindle{}
	spindle.config = config
	spindle.errors = new_error_handler()

	new_hash("default")

	config.domain = "https://qxoko.io/"
	config.default_path_type = ROOTED

	if data, ok := load_file_tree(); ok {
		spindle.file_tree = data
	}

	spindle.templates        = load_all_templates(spindle)
	spindle.partials         = load_all_partials(spindle)
	spindle.finder_cache     = make(map[string]*disk_object, 64)
	spindle.generated_images = make(map[uint32]*generated_image, 32)

	make_dir(public_path)

	found_file, ok := find_file(spindle.file_tree, "index")

	if !ok {
		panic("failed to find index")
	}

	found_file.is_used = true

	// _println("[used files]") 

	for {
		done := build_pages(spindle, spindle.file_tree)

		if spindle.errors.has_failures {
			break
		}
		if done {
			break
		}
	}

	// _println("\n[generated]")

	for _, image := range spindle.generated_images {
		if image.is_built {
			continue
		}
		if image.original.is_draft && !spindle.config.build_drafts {
			continue
		}
		if spindle.config.build_recent && !image.original.modtime.After(spindle.config.built_last) {
			continue
		}

		// _println("× ", format_index(rewrite_image_path(image.original.path, image.settings)))

		resize_image(spindle, image.original, image.settings) // @error
		image.is_built = true
	}

	if spindle.errors.has_errors() {
		fmt.Fprintln(os.Stderr, spindle.errors.render_term_errors())
	} else {
		save_time() // don't save built_last if there are any errors
	}

	// print_file_tree(spindle.file_tree.children, 0)
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

		switch file.file_type {
		default:
			if spindle.config.build_recent && !file.modtime.After(spindle.config.built_last) {
				continue
			}

			// _println(format_index(file.path))

			path := rewrite_root(file.path, public_path)
			make_dir(filepath.Dir(path))
			copy_file(file.path, path)

		case MARKUP:
			page, ok := load_page(spindle, file.path)
			if !ok {
				panic("failed to load page " + file.path)
			}

			assembled := render_syntax_tree(spindle, page)

			if spindle.config.build_recent && !file.modtime.After(spindle.config.built_last) {
				continue
			}

			// _println(format_index(file.path))

			output_path := rewrite_public(file.path, ".html")

			make_dir(filepath.Dir(output_path))

			assembled = format_html(assembled)

			if !write_file(output_path, assembled) {
				spindle.errors.new(FAILURE, "%q could not be written to disk", output_path)
				break main_loop
			}
		}
	}

	return is_done
}

func format_index(input string) string {
	if !strings.Contains(input, "index") {
		return filepath.Base(input)
	}
	return input
}