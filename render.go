package main

import "strings"

// base % variable
const base_hash uint32 = 537692064

func copy_map(input map[uint32]*ast_declare) map[uint32]*ast_declare {
	copy := make(map[uint32]*ast_declare, len(input))

	for k, v := range input {
		copy[k] = v
	}

	return copy
}

type _scope map[uint32]*ast_declare

type stack_container struct {
	anon_count int
	children   []ast_data
}

func push(array []*stack_container, content, wrapper []ast_data) []*stack_container {
	stack_entry := &stack_container{}

	stack_entry.anon_count = recursive_anon_count(wrapper)
	stack_entry.children   = content

	array = append(array, stack_entry)
	return array
}

func recursive_anon_count(children []ast_data) int {
	count := 0
	for _, entry := range children {
		if entry.type_check().is(DECL, DECL_TOKEN, DECL_BLOCK) {
			continue
		}
		if entry.type_check().is(VAR_ANON, VAR_ENUM) {
			count++
			continue
		}
		if x := entry.get_children(); len(x) > 0 {
			count += recursive_anon_count(x)
		}
	}
	return count
}

func immediate_decl_count(children []ast_data) int {
	count := 0
	for _, entry := range children {
		if entry.type_check().is(DECL, DECL_TOKEN, DECL_BLOCK) {
			count++
		}
	}
	return count
}

func render_syntax_tree(spindle *spindle, page *page_object) string {
	return render_ast(spindle, page, page.content, make(map[uint32]*ast_declare, 64), make([]*stack_container, 0, 8))
}

func render_ast(spindle *spindle, page *page_object, input []ast_data, incoming_scope _scope, anon_stack []*stack_container) string {
	scope := copy_map(incoming_scope)

	var popped_anon *stack_container
	can_commit_pop := false

	if len(anon_stack) > 0 {
		popped_anon = anon_stack[len(anon_stack) - 1]
		can_commit_pop = (popped_anon.anon_count <= 0)
	}

	buffer := strings.Builder{}
	buffer.Grow(256)

	index := 0

	for {
		if index > len(input) - 1 {
			break
		}

		entry := input[index]
		index++

		tc := entry.type_check()

		if tc.is(DECL, DECL_BLOCK, DECL_TOKEN) {
			entry := entry.(*ast_declare)
			scope[entry.field] = entry
			continue
		}

		if tc > is_lexer {
			// @todo this shouldn't exist in launch, it's
			// just here to catch mistakes in development
			panic("lexer type made it all the way to render")
		}

		if tc > is_formatter {
			// buffer.WriteString(format_convert[tc]) // @todo this was naive
			continue
		}

		switch entry.type_check() {
		case VAR, VAR_ENUM, VAR_ANON:
			entry := entry.(*ast_variable)

			text := ""

			// @todo this is difficult to follow, might need to clean up

			if entry.ast_type.is(VAR_ANON, VAR_ENUM) {
				if can_commit_pop {
					can_commit_pop = false
					anon_stack = anon_stack[:len(anon_stack) - 1]
				}
				popped_anon.anon_count -= 1
				text = render_ast(spindle, page, popped_anon.children, scope, anon_stack)

			} else if x, ok := scope[entry.field]; ok {
				text = render_ast(spindle, page, x.get_children(), scope, anon_stack)
			}

			if entry.ast_type == VAR_ENUM && entry.subname > 0 {
				args := unix_args(text)
				n    := int(entry.subname)

				if n > len(args) {
					panic("overflow on var_enum")
				}

				text = args[n - 1]
			}

			if entry.modifier != NONE {
				switch entry.modifier {
				// we'll pass the page down here which will
				// track slug usage and append -1 etc.
				case SLUG:  text = make_slug(text)
				case TITLE: text = make_title(text)
				case UPPER: text = strings.ToUpper(text)
				case LOWER: text = strings.ToLower(text)
				}
			}

			buffer.WriteString(text)

		case RES_FINDER:
			entry := entry.(*ast_finder)

			/*
				@todo cache these lookups for subsequent runs because they are _slow_.
				in fact, make the parser knowledgeable about whether or not _any_ node can be cached.
				then we can cache text on the fly, based on whether vars exist.
			*/

			find_text := render_ast(spindle, page, entry.children[:1], scope, anon_stack)

			// check cache
			found_file, ok := spindle.finder_cache[find_text]

			// if not in cache, do a full search
			if !ok {
				found_file, ok = find_file_descending(spindle.file_tree, find_text)
			}

			if ok {
				path := found_file.path

				switch entry.finder_type {
				case IMAGE:
					if !(found_file.file_type > is_image && found_file.file_type < end_image) {
						panic("res_find not an image") // @error
					}
				case PAGE:
					if !(found_file.file_type > is_page && found_file.file_type < end_page) {
						panic("res_find not a page") // @error
					}

					path = rewrite_ext(path, "") // @todo global config on pretty urls
				}

				if entry.path_type == NO_PATH_TYPE {
					entry.path_type = spindle.config.default_path_type
				}

				switch entry.path_type {
				case ROOTED:
					buffer.WriteRune('/')
					buffer.WriteString(rewrite_root(path, ""))

				case RELATIVE:
					if path, ok := filepath_relative(page.page_path, path); ok {
						buffer.WriteString(path)
					}

				case ABSOLUTE:
					buffer.WriteString(spindle.config.domain)
					buffer.WriteString(rewrite_root(path, ""))
				}

				found_file.is_used = true
				spindle.finder_cache[find_text] = found_file

			} else {
				// @error not found file
				panic("res_find didn't find file") // @error
			}
			continue

		case BLOCK:
			entry := entry.(*ast_block)

			x := entry.get_children()

			if entry.decl_hash > 0 {
				wrapper_block, ok := scope[entry.decl_hash]
				if ok {
					var new_scope _scope

					if immediate_decl_count(wrapper_block.get_children()) > 0 {
						new_scope = copy_map(scope)
					} else {
						new_scope = scope
					}

					anon_stack = push(anon_stack, x, wrapper_block.get_children())

					buffer.WriteString(render_ast(spindle, page, wrapper_block.get_children(), new_scope, anon_stack))
					continue
				}
			}

			// else:
			buffer.WriteString(render_ast(spindle, page, x, scope, anon_stack))

		case TOKEN:
			entry := entry.(*ast_token)

			x := entry.get_children()

			wrapper_block, ok := scope[entry.decl_hash]
			if ok {
				var new_scope _scope

				if immediate_decl_count(wrapper_block.get_children()) > 0 {
					new_scope = copy_map(scope)
				} else {
					new_scope = scope
				}

				anon_stack = push(anon_stack, x, wrapper_block.get_children())

				buffer.WriteString(render_ast(spindle, page, wrapper_block.get_children(), new_scope, anon_stack))
				continue
			}

			// else:
			buffer.WriteString(render_ast(spindle, page, x, scope, anon_stack))

		case NORMAL:
			entry := entry.(*ast_normal)

			x := entry.get_children()

			var text string

			if len(x) > 0 {
				text = render_ast(spindle, page, x, scope, anon_stack)
			} else {
				text = entry.field
			}

			buffer.WriteString(text)
		}
	}

	return buffer.String()
}

/*func push_string_on_scope(scope map[uint32]*ast_declare, ident, text string) {
	decl := &ast_declare {
		ast_type: DECL,
		field:    new_hash(ident),
	}
	decl.children = []ast_data{
		&ast_normal{
			ast_type: NORMAL,
			field:    text,
		},
	}
	scope[decl.field] = decl
}*/