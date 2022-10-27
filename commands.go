package main

import (
	"os"
	"fmt"
	"path/filepath"
)

const (
	VERSION uint8 = iota
	HELP
	INIT
	BUILD
	SERVE
)

// @todo this should be optional
const only_used = true

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
	config.default_path_type = ABSOLUTE

	if data, ok := load_file_tree(); ok {
		spindle.file_tree = data
	}

	spindle.templates    = load_all_templates(spindle)
	spindle.finder_cache = make(map[string]*disk_object, 64)

	make_dir(public_path)

	/*if validate_html(assembled) {
		_println("GOOD!")
	} else {
		_println("BAD!")
	}*/

	found_file, ok := find_file_descending(spindle.file_tree, "/index")

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

	if spindle.errors.has_errors() {
		fmt.Fprintln(os.Stderr, spindle.errors.render_term_errors())
	}
}

func build_pages(spindle *spindle, file *disk_object) bool {
	has_changes := false

	main_loop: for _, file := range file.children {
		if file.file_type == DIRECTORY {
			child_changes := build_pages(spindle, file)
			if child_changes {
				has_changes = true
			}
			continue
		}

		if only_used && file.is_used && file.is_built {
			continue
		}

		has_changes   = true
		file.is_built = true

		switch file.file_type {
		case MARKUP:
			page, ok := load_page(spindle, file.path)
			if !ok {
				panic("failed to load page") // this should be impossible, so i'm leaving it so it'll catch me out if it ever happens
			}

			assembled   := render_syntax_tree(spindle, page)
			output_path := rewrite_ext(rewrite_root(file.path, public_path), ".html")

			// assembled = format_html(assembled)

			if !write_file(output_path, assembled) {
				spindle.errors.new(FAILURE, "%q could not be written to disk", output_path)
				break main_loop
			}
		}
	}

	return has_changes
}