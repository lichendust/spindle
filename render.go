package main

import "fmt"
import "strings"

func render_syntax_tree(spindle *spindle, page *page_object) string {
	scope_stack := make([]map[uint32]*ast_declare, 4)
	scope_stack = append(scope_stack, make(map[uint32]*ast_declare, 16))

	renderer := &renderer{
		anon_stack:  make([]*anon_entry, 0, 4),
		scope_stack: scope_stack,
	}

	return renderer.render_ast(spindle, page, page.content)
}

type renderer struct {
	unwind      bool
	anon_stack  []*anon_entry
	scope_stack []map[uint32]*ast_declare
}

type anon_entry struct {
	anon_count int
	position   position
	children   []ast_data
	raw_string string
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

func (r *renderer) push_anon(content, wrapper []ast_data, pos position, raw_string string) {
	stack_entry := &anon_entry{
		anon_count: recursive_anon_count(wrapper),
		children:   content,
		position:   pos,
		raw_string:   raw_string,
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

func (r *renderer) push_blank_scope(alloc int) {
	r.scope_stack = append(r.scope_stack, make(map[uint32]*ast_declare, alloc))
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
			entry := entry.(*ast_base)
			if t, ok := spindle.templates[new_hash(entry.field)]; ok {
				r.write_collective_to_scope(spindle, t.top_scope)
			} else {
				spindle.errors.new_pos(RENDER_FAILURE, entry.position, "failed to load template %q", entry.field)
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

			r.push_blank_scope(8)
			buffer.WriteString(r.render_ast(spindle, page, p.content))
			r.pop_scope()

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
				found_file, ok = find_file_descending(spindle.file_tree, find_text)
			}

			imported_page, page_success := load_page(spindle, found_file.path)

			if !page_success {
				spindle.errors.new_pos(RENDER_FAILURE, entry.position, "didn't find page %q in import", found_file.path)
				r.unwind = true
				return ""
			}

			r.push_blank_scope(16)
			r.write_collective_to_scope(spindle, imported_page.top_scope)

			// @todo undefined behaviour for %% in imports
			// we should probably disallow it but we can't know until
			// we've parsed it

			buffer.WriteString(r.render_ast(spindle, page, t.get_children()))

			r.pop_scope()

			spindle.finder_cache[find_text] = found_file

		case TEMPLATE:
			entry := entry.(*ast_builtin)

			t, ok := spindle.templates[entry.hash_name]

			if !ok {
				spindle.errors.new_pos(RENDER_FAILURE, entry.position, "failed to load template %q", get_hash(entry.hash_name))
				r.unwind = true
				return ""
			}

			// if template is the first thing in the block/page
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
				r.push_blank_scope(16)

				r.write_collective_to_scope(spindle, input[1:])
				r.push_anon(input[1:], t.content, t.position, t.raw_string)

				buffer.WriteString(r.render_ast(spindle, page, t.content))
				r.pop_scope()

				return buffer.String() // hard exit
			}

			// if we're not the first, we just pull in the
			// declarations from the template to be used from
			// here on out in this scope
			r.write_collective_to_scope(spindle, t.top_scope)

		case VAR, VAR_ENUM, VAR_ANON:
			entry := entry.(*ast_variable)

			text := ""

			if entry.ast_type.is(VAR_ANON, VAR_ENUM) {
				popped_anon.anon_count -= 1

				if popped_anon.anon_count <= 0 {
					r.pop_anon()
				}

				if entry.modifier == RAW_SUB {
					pos := popped_anon.position
					raw := popped_anon.raw_string[pos.start:pos.end]

					raw = reindent_text(raw)

					new := []ast_data{&ast_base{
						ast_type: RAW,
						field:    raw,
					}}

					text = r.render_ast(spindle, page, new)
				} else {
					text = r.render_ast(spindle, page, popped_anon.children)
				}

			} else {
				if found, ok := r.get_in_scope(entry.field); ok {
					text = r.render_ast(spindle, page, found.get_children())
				}
			}

			if entry.ast_type == VAR_ENUM && entry.subname > 0 {
				args := unix_args(text)
				n    := int(entry.subname)

				if n > len(args) {
					panic("overflow on var_enum")
				}

				text = args[n - 1]
			}

			if entry.modifier > mod_active {
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

			// check cache
			found_file, ok := spindle.finder_cache[find_text]

			// if not in cache, do a full search
			if !ok {
				found_file, ok = find_file_descending(spindle.file_tree, find_text)
			}

			if ok {
				path := found_file.path

				switch entry.finder_type {
				case _IMAGE:
					if !(found_file.file_type > is_image && found_file.file_type < end_image) {
						spindle.errors.new_pos(RENDER_FAILURE, entry.position, "image resource finder cannot process non-image file %q", find_text)
						r.unwind = true
					}

					if entry.image_settings != nil {
						path = rewrite_image_path(path, entry.image_settings)

						if entry.image_settings.format == 0 {
							entry.image_settings.format = found_file.file_type
						}

						spindle.generated_images = append(spindle.generated_images, &generated_image{
							false, found_file, entry.image_settings,
						})

					} else {
						// if it has modifiers, only the generated image is used
						// so we don't mark it here
						found_file.is_used = true
					}

				case _PAGE:
					if !(found_file.file_type > is_page && found_file.file_type < end_page) {
						spindle.errors.new_pos(RENDER_FAILURE, entry.position, "page resource finder cannot process non-page file %q", find_text)
						r.unwind = true
					}
					path = rewrite_ext(path, "") // @todo global config on pretty urls
					found_file.is_used = true

				case _STATIC:
					if !(found_file.file_type > is_static && found_file.file_type < end_static) {
						spindle.errors.new_pos(RENDER_FAILURE, entry.position, "static resource finder cannot process non-static file %q", find_text)
						r.unwind = true
					}
					found_file.is_used = true
				}

				if spindle.server_mode {
					entry.path_type = ROOTED
				} else if entry.path_type == NO_PATH_TYPE {
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

				spindle.finder_cache[find_text] = found_file

			} else {
				spindle.errors.new_pos(RENDER_WARNING, entry.position, "resource finder did not find file %q", find_text)
			}
			continue

		case BLOCK:
			entry := entry.(*ast_block)

			x := entry.get_children()
			r.push_blank_scope(8)

			if entry.decl_hash > 0 {
				wrapper_block, ok := r.get_in_scope(entry.decl_hash)
				if ok {
					r.push_anon(x, wrapper_block.get_children(), *entry.get_position(), page.raw_string)

					buffer.WriteString(r.render_ast(spindle, page, wrapper_block.get_children()))

					r.pop_scope()
					continue
				}
			}

			// else:
			buffer.WriteString(r.render_ast(spindle, page, x))
			r.pop_scope()

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
				n := immediate_decl_count(wrapper_block.get_children())

				if n > 0 {
					r.push_blank_scope(n + 1)
				}

				r.push_anon(x, wrapper_block.get_children(), *entry.get_position(), page.raw_string)

				buffer.WriteString(r.render_ast(spindle, page, wrapper_block.get_children()))

				if n > 0 {
					r.pop_scope()
				}
				continue
			}

			if entry.decl_hash != stop_hash {
				spindle.errors.new_pos(RENDER_WARNING, entry.position, "token %q does not have a template — output may be unexpected unless it is escaped", entry.orig_field)
			}

			// else:
			buffer.WriteString(r.render_ast(spindle, page, x))

		case CONTROL_IF:
			entry := entry.(*ast_if)

			if !r.evaluate_if(entry) {
				continue
			}

			the_block := entry.get_children()[0].(*ast_block)

			x := the_block.get_children()
			r.push_blank_scope(8)

			if the_block.decl_hash > 0 {
				wrapper_block, ok := r.get_in_scope(the_block.decl_hash)
				if ok {
					r.push_anon(x, wrapper_block.get_children(), *entry.get_position(), page.raw_string)

					buffer.WriteString(r.render_ast(spindle, page, wrapper_block.get_children()))

					r.pop_scope()
					continue
				}
			}

			// else:
			buffer.WriteString(r.render_ast(spindle, page, x))
			r.pop_scope()

		case CONTROL_FOR:
			entry := entry.(*ast_for)
			array := unix_args(r.render_ast(spindle, page, []ast_data{entry.iterator_source}))

			if len(array) == 0 {
				continue
			}

			the_block := entry.get_children()[0].(*ast_block)

			r.push_blank_scope(12)

			sub_buffer := strings.Builder{}
			sub_buffer.Grow(512)

			for _, t := range array {
				r.push_string_on_scope(it_hash, t)
				sub_buffer.WriteString(r.render_ast(spindle, page, the_block.get_children()))
			}

			if the_block.decl_hash > 0 {
				wrapper_block, ok := r.get_in_scope(the_block.decl_hash)
				if ok {
					r.push_anon([]ast_data{&ast_base{ast_type:NORMAL,field:sub_buffer.String()}}, wrapper_block.get_children(), *entry.get_position(), page.raw_string)
					buffer.WriteString(r.render_ast(spindle, page, wrapper_block.get_children()))
				}
			} else {
				buffer.WriteString(sub_buffer.String())
			}

			sub_buffer.Reset()
			r.pop_scope()

		case NORMAL:
			entry := entry.(*ast_base)

			x := entry.get_children()

			if len(x) > 0 {
				wrapper_block, ok := r.get_in_scope(default_hash)
				if ok {
					n := immediate_decl_count(wrapper_block.get_children())

					if n > 0 {
						r.push_blank_scope(n + 1)
					}

					r.push_anon(x, wrapper_block.get_children(), *entry.get_position(), page.raw_string)

					buffer.WriteString(r.render_ast(spindle, page, wrapper_block.get_children()))

					if n > 0 {
						r.pop_scope()
					}
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

func count_repeat_tokens(input []ast_data, hash uint32) int {
	for i, forward_check := range input {
		t := forward_check.type_check()

		if t != TOKEN {
			return i
		}

		x := forward_check.(*ast_token)

		if x.decl_hash != hash {
			return i
		}
	}

	return len(input)
}

func reindent_text(input string) string {
	input = strings.ReplaceAll(input, "\t", "    ")
	lines := strings.Split(input, "\n")

	shortest_indent := len(input)

	for _, line := range lines {
		count := 0

		for _, c := range line {
			if c != ' ' {
				break
			}
			count ++
		}

		if count < shortest_indent {
			shortest_indent = count
		}
	}

	if shortest_indent == 0 {
		return input
	}

	buffer := strings.Builder{}
	buffer.Grow(len(input))

	for _, line := range lines {
		buffer.WriteString(line[shortest_indent:])
		buffer.WriteRune('\n')
	}

	render := buffer.String()
	render = render[:len(render) - 1] // trailing newline

	return render
}