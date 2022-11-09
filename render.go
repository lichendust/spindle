package main

import "fmt"
import "strings"

func render_syntax_tree(spindle *spindle, page *page_object, import_condition uint32) string {
	scope_stack := make([]map[uint32]*ast_declare, 0, 4)
	scope_stack = append(scope_stack, make(map[uint32]*ast_declare, 32))

	r := &renderer{
		import_condition: import_condition,
		anon_stack:       make([]*anon_entry, 0, 4),
		scope_stack:      scope_stack,
	}

	r.push_string_on_scope(is_server_hash, "") // just has to exist
	r.push_string_on_scope(reload_script_hash, reload_script)

	return r.render_ast(spindle, page, page.content)
}

type renderer struct {
	unwind           bool
	import_condition uint32
	anon_stack       []*anon_entry
	scope_stack      []map[uint32]*ast_declare
}

type anon_entry struct {
	anon_count int
	position   position
	children   []ast_data
}

func (r *renderer) get_anon() *anon_entry {
	if len(r.anon_stack) > 0 {
		return r.anon_stack[len(r.anon_stack) - 1]
	}
	return nil
}

func (r *renderer) pop_anon() {
	if len(r.anon_stack) > 0 {
		r.anon_stack = r.anon_stack[:len(r.anon_stack) - 1]
	}
}

func (r *renderer) push_anon(content, wrapper []ast_data, pos position) {
	stack_entry := &anon_entry{
		anon_count: recursive_anon_count(wrapper),
		children:   content,
		position:   pos,
	}
	r.anon_stack = append(r.anon_stack, stack_entry)
}

func (r *renderer) get_in_scope(value uint32) (ast_data, bool) {
	for i := len(r.scope_stack) - 1; i >= 0; i-- {
		level := r.scope_stack[i]

		if x, ok := level[value]; ok {
			return x, true
		}
	}
	return nil, false
}

func (r *renderer) write_to_scope(field uint32, entry *ast_declare) {
	r.scope_stack[len(r.scope_stack) - 1][field] = entry
}

func (r *renderer) push_blank_scope(alloc int) bool {
	if alloc == 0 {
		return false
	}
	r.scope_stack = append(r.scope_stack, make(map[uint32]*ast_declare, alloc))
	return true
}

func (r *renderer) pop_scope() {
	r.scope_stack = r.scope_stack[:len(r.scope_stack) - 1]
}

func (r *renderer) delete_scope_entry(value uint32) {
	for i := len(r.scope_stack) - 1; i >= 0; i-- {
		level := r.scope_stack[i]
		if _, ok := level[value]; ok {
			delete(level, value)
			break
		}
	}
}

func (r *renderer) push_string_on_scope(ident uint32, text string) {
	decl := &ast_declare {
		ast_type: DECL,
		field:    ident,
	}
	decl.children = []ast_data{
		&ast_base{
			ast_type: NORMAL,
			field:    text,
		},
	}
	r.write_to_scope(decl.field, decl)
}

func (r *renderer) evaluate_if(entry *ast_if) bool {
	result  := false
	has_not := false

	for _, sub := range entry.condition_list {
		switch sub.type_check() {
		case OP_NOT:
			has_not = true
			continue

		case OP_OR:
			if result {
				return true
			}
			continue

		case VAR:
			_, ok := r.get_in_scope(sub.(*ast_variable).field)
			if has_not {
				ok = !ok
			}
			result = ok
		}

		if has_not {
			has_not = false
		}
	}

	return result
}

func (r *renderer) write_collective_to_scope(spindle *spindle, input []ast_data) {
	for _, entry := range input {
		_type := entry.type_check()

		if _type.is(DECL, DECL_TOKEN, DECL_BLOCK) {
			entry := entry.(*ast_declare)
			r.write_to_scope(entry.field, entry)
			continue
		}

		if _type == TEMPLATE {
			entry := entry.(*ast_builtin)
			if t, ok := spindle.templates[entry.hash_name]; ok {
				r.write_collective_to_scope(spindle, t.top_scope)
			} else if t, ok := r.get_in_scope(entry.hash_name); ok {
				x := t.get_children()

				if len(x) == 1 && x[0].type_check() == BLOCK {
					r.write_collective_to_scope(spindle, x[0].get_children())
				}

			} else {
				spindle.errors.new_pos(RENDER_FAILURE, entry.position, "failed to load template %q", get_hash(entry.hash_name))
				r.unwind = true
				break
			}
			continue
		}
	}
}

func (r *renderer) render_ast(spindle *spindle, page *page_object, input []ast_data) string {
	if r.unwind {
		return ""
	}

	buffer := strings.Builder{}
	buffer.Grow(256)

	popped_anon := r.get_anon()

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
			r.write_to_scope(entry.field, entry)
			continue
		}

		if tc > is_lexer {
			// @todo this shouldn't exist in launch, it's
			// just here to catch mistakes in development
			panic("lexer type made it all the way to render")
		}

		/*if tc > is_formatter {
			continue
		}*/

		switch tc {
		case WHITESPACE:
			buffer.WriteRune(' ')

		case SCOPE_UNSET:
			entry := entry.(*ast_builtin)
			r.delete_scope_entry(entry.hash_name)

		case PARTIAL:
			entry := entry.(*ast_builtin)

			p, ok := spindle.partials[entry.hash_name]

			if !ok {
				spindle.errors.new_pos(RENDER_FAILURE, entry.position, "failed to load partial %q", get_hash(entry.hash_name))
				r.unwind = true
				return ""
			}

			did_push := r.push_blank_scope(immediate_decl_count(p.content))
			buffer.WriteString(r.render_ast(spindle, page, p.content))

			if did_push { r.pop_scope() }

		case IMPORT:
			entry := entry.(*ast_builtin)

			t, ok := r.get_in_scope(entry.hash_name)

			if !ok {
				spindle.errors.new_pos(RENDER_FAILURE, entry.position, "no such template for import %q", get_hash(entry.hash_name))
				r.unwind = true
				return ""
			}

			find_text := r.render_ast(spindle, page, entry.children)

			// check cache
			found_file, ok := spindle.finder_cache[find_text]

			// if not in cache, do a full search
			if !ok {
				found_file, ok = find_file(spindle.file_tree, find_text)
			}

			if ok {
				if !spindle.build_drafts && found_file.is_draft {
					spindle.errors.new_pos(RENDER_WARNING, entry.position, "imported page %q is draft!", found_file.path)
				}
			} else {
				spindle.errors.new_pos(RENDER_WARNING, entry.position, "didn't find page %q in import", find_text)
			}

			imported_page, page_success := load_page(spindle, found_file.path) // @todo cache

			if r.import_condition > 0 {
				has_match := false

				for _, entry := range imported_page.top_scope {
					if entry.type_check() != TEMPLATE {
						entry := entry.(*ast_declare)
						if entry.field == r.import_condition {
							has_match = true
							break
						}
					}
				}

				if !has_match {
					continue
				}
			}

			if !page_success {
				spindle.errors.new_pos(RENDER_FAILURE, entry.position, "didn't find page %q in import", found_file.path)
				r.unwind = true
				return ""
			}

			r.push_blank_scope(immediate_decl_count(imported_page.top_scope) + 1)
			r.write_collective_to_scope(spindle, imported_page.top_scope)
			r.push_string_on_scope(new_hash("path"), find_text)

			// @todo undefined behaviour for %% in imports
			// we should probably disallow it but we can't know until
			// we've parsed it

			buffer.WriteString(r.render_ast(spindle, page, t.get_children()))

			r.pop_scope()

			spindle.finder_cache[find_text] = found_file

		case TEMPLATE:
			entry := entry.(*ast_builtin)

			if t, ok := spindle.templates[entry.hash_name]; ok {
				// if first in page / block
				if t.has_body && index == 1 {
					// if this happens we completely break flow,
					// swapping the entire input for the template
					// and return the rendered string immediately
					// to the caller level above

					// we're treating the template body as a
					// block-template declaration and abdicating
					// responsibility on this pass

					// we also reverse the order in which the
					// top-level scope is applied
					did_push := r.push_blank_scope(immediate_decl_count(t.content))

					r.write_collective_to_scope(spindle, input[1:])
					r.push_anon(input[1:], t.content, t.position)

					buffer.WriteString(r.render_ast(spindle, page, t.content))

					if did_push {
						r.pop_scope()
					}
					return buffer.String() // hard exit
				}

				// if we're not the first, we just pull in the
				// declarations from the template to be used from
				// here on out in this scope
				r.write_collective_to_scope(spindle, t.top_scope)

			} else if t, ok := r.get_in_scope(entry.hash_name); ok {
				x := t.get_children()

				if len(x) == 1 && x[0].type_check() == BLOCK {
					r.write_collective_to_scope(spindle, x[0].get_children())
				}

			} else {
				spindle.errors.new_pos(RENDER_FAILURE, entry.position, "failed to load template %q", get_hash(entry.hash_name))
				r.unwind = true
				break
			}

		case VAR, VAR_ENUM, VAR_ANON:
			entry := entry.(*ast_variable)

			text := ""

			if entry.ast_type.is(VAR_ANON, VAR_ENUM) {
				if popped_anon == nil {
					_println(entry.position)
					panic("popped anon was missing!")
				}

				popped_anon.anon_count -= 1

				if popped_anon.anon_count <= 0 {
					r.pop_anon()
				}

				text = r.render_ast(spindle, page, popped_anon.children)

			} else {
				if found, ok := r.get_in_scope(entry.field); ok {
					text = r.render_ast(spindle, page, found.get_children())
				}
			}

			if entry.ast_type == VAR_ENUM && entry.subname > 0 {
				args := unix_args(text)
				n    := int(entry.subname)

				if n > len(args) {
					spindle.errors.new_pos(
						RENDER_WARNING, popped_anon.position,
						"this line only supplies %d arguments\n    template variable at %s — line %d has requested %d arguments\n    output may be unexpected",
						len(args), entry.position.file_path, entry.position.line, n,
					)
					text = ""
				} else {
					text = args[n - 1]
				}
			}

			if entry.modifier > NONE {
				switch entry.modifier {
				case SLUG:
					text = make_slug(text)
				case UNIQUE_SLUG:
					text = make_slug(text)
					if n, ok := page.slug_tracker[text]; ok {
						page.slug_tracker[text] = n + 1
						text = fmt.Sprintf("%s-%d", text, n)
					} else {
						page.slug_tracker[text] = 1
					}
				case TITLE:
					text = make_title(text)
				case UPPER:
					text = strings.ToUpper(text)
				case LOWER:
					text = strings.ToLower(text)
				}
			}

			buffer.WriteString(text)

		case RES_FINDER:
			entry := entry.(*ast_finder)

			find_text := r.render_ast(spindle, page, entry.children)

			if is_ext_url(find_text) {
				buffer.WriteString(find_text)
				continue
			}

			// check cache
			found_file, ok := spindle.finder_cache[find_text]

			// if not in cache, do a full search
			if !ok {
				found_file, ok = find_file(spindle.file_tree, find_text)
			}

			if ok {
				if !spindle.build_drafts && found_file.is_draft {
					spindle.errors.new_pos(RENDER_WARNING, entry.position, "resource finder links draft %q: becomes link to missing resource when built", found_file.path)
				}

				the_url := ""

				if entry.path_type == NO_PATH_TYPE {
					entry.path_type = spindle.default_path_mode
				}

				switch entry.finder_type {
				case _IMAGE:
					if !(found_file.file_type > is_image && found_file.file_type < end_image) {
						spindle.errors.new_pos(RENDER_FAILURE, entry.position, "image resource finder cannot process non-image file %q", find_text)
						r.unwind = true
					}

					if entry.image_settings != nil {
						settings := *entry.image_settings
						if settings.file_type == 0 {
							settings.file_type = found_file.file_type
						}

						the_url = make_generated_image_url(spindle, found_file, &settings, entry.path_type, page.page_path)

						{
							hash := new_hash(the_url)
							if _, ok := spindle.gen_images[hash]; !ok {
								spindle.gen_images[hash] = &gen_image{
									false, found_file, &settings,
								}
							}
						}

					} else {
						the_url = make_general_url(spindle, found_file, entry.path_type, page.page_path)

						// if it has modifiers, only the generated image is used
						// so we don't mark it here
						found_file.is_used = true
					}

				case _PAGE:
					if !(found_file.file_type > is_page && found_file.file_type < end_page) {
						spindle.errors.new_pos(RENDER_FAILURE, entry.position, "page resource finder cannot process non-page file %q", find_text)
						r.unwind = true
					}

					the_url = make_page_url(spindle, found_file, entry.path_type, page.page_path)
					found_file.is_used = true

				case _STATIC:
					if !(found_file.file_type > is_static && found_file.file_type < end_static) {
						spindle.errors.new_pos(RENDER_FAILURE, entry.position, "static resource finder cannot process non-static file %q", find_text)
						r.unwind = true
					}

					the_url = make_general_url(spindle, found_file, entry.path_type, page.page_path)
					found_file.is_used = true
				}

				buffer.WriteString(the_url)

				spindle.finder_cache[find_text] = found_file

			} else {
				spindle.errors.new_pos(RENDER_WARNING, entry.position, "resource finder did not find file %q", find_text)
			}
			continue

		case BLOCK:
			entry := entry.(*ast_block)

			x := entry.get_children()

			if entry.decl_hash > 0 {
				wrapper_block, ok := r.get_in_scope(entry.decl_hash)
				if ok {
					children := wrapper_block.get_children()

					did_push := r.push_blank_scope(immediate_decl_count(children))
					r.push_anon(x, children, *entry.get_position())

					buffer.WriteString(r.render_ast(spindle, page, children))

					if did_push {
						r.pop_scope()
					}
					continue
				}
			}

			// else:
			did_push := r.push_blank_scope(immediate_decl_count(x))
			buffer.WriteString(r.render_ast(spindle, page, x))
			if did_push {
				r.pop_scope()
			}

		case TOKEN:
			entry := entry.(*ast_token)

			/*_, has_group := r.get_in_scope(entry.decl_hash + 1)

			if has_group {
				n := count_repeat_tokens(input[index:], entry.decl_hash)
				index += n
			}*/

			x := entry.get_children()

			wrapper_block, ok := r.get_in_scope(entry.decl_hash)

			if ok {
				children := wrapper_block.get_children()

				did_push := r.push_blank_scope(immediate_decl_count(children))
				r.push_anon(x, children, *entry.get_position())

				buffer.WriteString(apply_regex_array(spindle.inline, r.render_ast(spindle, page, children)))

				if did_push {
					r.pop_scope()
				}
				continue
			} else {
				if len(x) == 0 {
					wrapper_block, ok := r.get_in_scope(default_hash)
					if ok {
						children := wrapper_block.get_children()
						r.push_anon(x, children, *entry.get_position())
						buffer.WriteString(apply_regex_array(spindle.inline, r.render_ast(spindle, page, children)))
						continue
					}
				}
			}

			if entry.decl_hash != stop_hash {
				spindle.errors.new_pos(RENDER_WARNING, entry.position, "token %q does not have a template — output may be unexpected unless it is escaped", entry.orig_field)
			}

			// else:
			buffer.WriteString(apply_regex_array(spindle.inline, r.render_ast(spindle, page, x)))

		case CONTROL_IF:
			entry := entry.(*ast_if)

			if !r.evaluate_if(entry) {
				continue
			}

			the_block := entry.get_children()[0].(*ast_block)

			x := the_block.get_children()

			if the_block.decl_hash > 0 {
				wrapper_block, ok := r.get_in_scope(the_block.decl_hash)
				if ok {
					children := wrapper_block.get_children()

					did_push := r.push_blank_scope(immediate_decl_count(children))
					r.push_anon(x, children, *entry.get_position())

					buffer.WriteString(r.render_ast(spindle, page, children))

					if did_push {
						r.pop_scope()
					}
					continue
				}
			}

			// else:
			did_push := r.push_blank_scope(immediate_decl_count(x))
			buffer.WriteString(r.render_ast(spindle, page, x))
			if did_push { r.pop_scope() }

		case CONTROL_FOR:
			entry := entry.(*ast_for)
			array := unix_args(r.render_ast(spindle, page, []ast_data{entry.iterator_source}))

			if len(array) == 0 {
				continue
			}

			the_block := entry.get_children()[0].(*ast_block)

			did_push := r.push_blank_scope(immediate_decl_count(the_block.get_children()))

			sub_buffer := strings.Builder{}
			sub_buffer.Grow(512)

			for _, t := range array {
				r.push_string_on_scope(it_hash, t)
				sub_buffer.WriteString(r.render_ast(spindle, page, the_block.get_children()))
			}

			if the_block.decl_hash > 0 {
				wrapper_block, ok := r.get_in_scope(the_block.decl_hash)
				if ok {
					r.push_anon([]ast_data{&ast_base{ast_type:NORMAL,field:sub_buffer.String()}}, wrapper_block.get_children(), *entry.get_position())
					buffer.WriteString(r.render_ast(spindle, page, wrapper_block.get_children()))
				}
			} else {
				buffer.WriteString(sub_buffer.String())
			}

			sub_buffer.Reset()

			if did_push {
				r.pop_scope()
			}

		case NORMAL:
			entry := entry.(*ast_base)

			x := entry.get_children()

			if len(x) > 0 {
				wrapper_block, ok := r.get_in_scope(default_hash)
				if ok {
					children := wrapper_block.get_children()
					did_push := r.push_blank_scope(immediate_decl_count(children))

					r.push_anon(x, children, *entry.get_position())

					buffer.WriteString(apply_regex_array(spindle.inline, r.render_ast(spindle, page, children)))

					if did_push { r.pop_scope() }
					continue
				}

				// else:
				buffer.WriteString(r.render_ast(spindle, page, x))
			} else {
				buffer.WriteString(entry.field)
			}

		case RAW:
			entry := entry.(*ast_base)

			x := entry.get_children()

			if len(x) > 0 {
				buffer.WriteString(r.render_ast(spindle, page, x))
			} else {
				buffer.WriteString(entry.field)
			}
		}
	}

	return buffer.String()
}