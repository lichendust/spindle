package main

import (
	"fmt"
	"time"
	"sort"
	"strings"
)

type renderer struct {
	unwind bool

	anon_stack   []*anon_entry
	scope_stack  []map[uint32]*ast_declare
	slug_tracker map[string]uint
}

func render_syntax_tree(spindle *spindle, page *Page) string {
	scope_stack := make([]map[uint32]*ast_declare, 0, 4)
	scope_stack = append(scope_stack, make(map[uint32]*ast_declare, 32))

	r := &renderer{
		anon_stack:   make([]*anon_entry, 0, 4),
		scope_stack:  scope_stack,
		slug_tracker: make(map[string]uint, 16),
	}

	if spindle.server_mode {
		r.push_string_to_scope(IS_SERVER_HASH, "") // just has to exist
		r.push_string_to_scope(RELOAD_SCRIPT_HASH, reload_script)
	}
	if page.import_hash > 0 {
		r.push_string_to_scope(TAGINATOR_ACTIVE_HASH, "") // just has to exist
		r.push_string_to_scope(TAGINATOR_TAG_HASH, page.import_cond)
		r.push_string_to_scope(TAGINATOR_PARENT_HASH, make_page_url(spindle, &page.file.file_info, spindle.path_mode, ""))
	}

	r.push_string_to_scope(CANONICAL_HASH, make_page_url(spindle, &page.file.file_info, ABSOLUTE, ""))
	r.push_string_to_scope(URL_HASH, make_page_url(spindle, &page.file.file_info, spindle.path_mode, ""))

	// spindle.current_year
	r.push_string_to_scope(519639417, fmt.Sprintf("%d", time.Now().Year()))

	return r.render_ast(spindle, page, page.content)
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

func (r *renderer) get_in_scope(value uint32) (*ast_declare, bool) {
	for i := len(r.scope_stack) - 1; i >= 0; i -= 1 {
		level := r.scope_stack[i]

		if x, ok := level[value]; ok {
			if x.ast_type == DECL_REJECT {
				break
			}
			return x, true
		}
	}

	return nil, false
}

func (r *renderer) write_to_scope(field uint32, entry *ast_declare) {
	if entry.is_soft {
		x, ok := r.get_in_scope(field)
		if ok && !x.is_soft {
			return
		}
	}

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
	x := &ast_declare{ast_type:DECL_REJECT}
	r.write_to_scope(value,     x)
	r.write_to_scope(value + 1, x)
}

func (r *renderer) push_string_to_scope(ident uint32, text string) {
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

func (r *renderer) write_collective_to_scope(spindle *spindle, page *Page, input []ast_data) {
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
				r.write_collective_to_scope(spindle, page, t.top_scope)
			} else if t, ok := r.get_in_scope(entry.hash_name); ok {
				x := t.get_children()

				if len(x) == 1 && x[0].type_check() == BLOCK {
					r.write_collective_to_scope(spindle, page, x[0].get_children())
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

	if entry.is_else {
		result = !result
	}

	return result
}

func (r *renderer) skip_import_condition(spindle *spindle, page *Page, data_slice []ast_data) bool {
	for _, entry := range data_slice {
		if entry.type_check() == DECL {
			entry := entry.(*ast_declare)

			if entry.field == page.import_hash {
				fields := unix_args(strings.ToLower(r.render_ast(spindle, page, entry.children)))

				for _, tag := range fields {
					if tag == page.import_cond {
						return false
					}
				}

				return true
			}
		}
	}

	return false
}


func __recurse(r *renderer, spindle *spindle, page *Page, input []ast_data, target_hash uint32) map[string]bool {
	index := 0

	capture := make(map[string]bool, 4)

	for {
		if index > len(input) - 1 {
			break
		}

		entry := input[index]
		index += 1

		switch entry.type_check() {
		case DECL:
			entry := entry.(*ast_declare)

			if entry.field == target_hash {
				fields := unix_args(r.render_ast(spindle, page, entry.children))

				for _, tag := range fields {
					tag = strings.ToLower(tag)

					capture[tag] = true

					file_path := tag_path(make_general_url(spindle, page.file, NO_PATH_TYPE, ""), spindle.tag_path, tag)
					seek_path := rewrite_ext(file_path, "")

					if _, ok := spindle.gen_pages[seek_path]; ok {
						continue
					}

					copy := &Page{}

					copy.content   = page.content
					copy.top_scope = page.top_scope
					copy.position  = page.position

					copy.file        = page.file
					copy.page_path   = page.page_path
					copy.import_cond = tag
					copy.import_hash = target_hash

					spindle.gen_pages[seek_path] = copy
				}
			}

		case IMPORT:
			entry := entry.(*ast_builtin)

			// @todo CRASH CENTRAL BAYBEEE
			find_text := r.render_ast(spindle, page, entry.children)
			found_file, _ := render_find_file(spindle, page, find_text)
			imported_page, _ := load_page(spindle, found_file.path)

			res := __recurse(r, spindle, page, imported_page.top_scope, target_hash)
			for x := range res {
				capture[x] = true
			}

		case BLOCK:
			res := __recurse(r, spindle, page, entry.get_children(), target_hash)
			for x := range res {
				capture[x] = true
			}
		}
	}

	return capture
}

func (r *renderer) do_import_seek(spindle *spindle, page *Page, target_hash uint32) {
	capture := __recurse(r, spindle, page, page.content, target_hash)

	array := make([]string, 0, len(capture))

	for x := range capture {
		array = append(array, x) // @todo make sure strings get requoted if needed
	}

	sort.Strings(array)

	r.push_string_to_scope(TAGINATOR_ALL_HASH, strings.Join(array, " "))
}

/*func (r *renderer) do_import_seek(spindle *spindle, page *Page, data_slice []ast_data) {
	for _, entry := range data_slice {
		if entry.type_check() == DECL {
			entry := entry.(*ast_declare)

			if entry.field == page.import_hash {
				fields := unix_args(r.render_ast(spindle, page, entry.children))

				for _, tag := range fields {
					tag = strings.ToLower(tag)

					file_path := tag_path(make_general_url(spindle, page.file, NO_PATH_TYPE, ""), spindle.tag_path, tag)
					seek_path := rewrite_ext(file_path, "")

					if _, ok := spindle.gen_pages[seek_path]; ok {
						continue
					}

					// @todo
					copy := &Page{}

					copy.content   = page.content
					copy.top_scope = page.top_scope
					copy.position  = page.position

					copy.file        = page.file
					copy.page_path   = page.page_path
					copy.import_cond = tag

					spindle.gen_pages[seek_path] = copy
				}

				break
			}
		}
	}
}*/

func (r *renderer) render_ast(spindle *spindle, page *Page, input []ast_data) string {
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
		index += 1

		tc := entry.type_check()

		if tc.is(DECL, DECL_BLOCK, DECL_TOKEN) {
			entry := entry.(*ast_declare)
			r.write_to_scope(entry.field, entry)

			// if we find a taginator in a scope
			if tc == DECL && entry.field == TAGINATOR_HASH {
				r.do_import_seek(spindle, page, new_hash(r.render_ast(spindle, page, entry.children)))
			}
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

		case SCRIPT:
			entry := entry.(*ast_script)

			x := get_hash(entry.hash_name) // @todo aaaaaaaaaaaaaaaaaaargh gross

			blob, ok := load_file("config/scripts/" + x + ".js")
			if !ok {
				spindle.errors.new_pos(RENDER_FAILURE, entry.position, "failed to load script %q", x) // get_hash(entry.hash_name))
				r.unwind = true
				return ""
			}

			args := unix_args(r.render_ast(spindle, page, entry.children))

			if res, ok := r.script_call(spindle, page, entry.position.line, blob, args...); ok {
				buffer.WriteString(res)
			} else {
				spindle.errors.new_pos(RENDER_FAILURE, entry.position, "failed to execute script %q", x) // get_hash(entry.hash_name))
			}

		case SCOPE_UNSET:
			entry := entry.(*ast_builtin)
			r.delete_scope_entry(entry.hash_name)
			r.delete_scope_entry(entry.hash_name + 1)

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

			// @todo we need a type check pass all these inline errors are going to give me a seizure

			t, ok := r.get_in_scope(entry.hash_name)
			if !ok {
				spindle.errors.new_pos(RENDER_FAILURE, entry.position, "no such template for import %q", get_hash(entry.hash_name))
				r.unwind = true
				return ""
			}

			find_text := r.render_ast(spindle, page, entry.children)

			found_file, ok := render_find_file(spindle, page, find_text)
			if ok {
				if !spindle.server_mode && !spindle.build_drafts && found_file.is_draft {
					spindle.errors.new_pos(RENDER_WARNING, entry.position, "imported page %q is draft!", found_file.path)
				}
			} else {
				spindle.errors.new_pos(RENDER_WARNING, entry.position, "didn't find page %q in import", find_text)
				continue
			}

			imported_page, page_success := load_page(spindle, found_file.path)
			if !page_success {
				spindle.errors.new_pos(RENDER_FAILURE, entry.position, "didn't find page %q in import", found_file.path)
				r.unwind = true
				return ""
			}

			if page.import_hash > 0 {
				if r.skip_import_condition(spindle, page, imported_page.top_scope) {
					continue
				}
			}

			r.push_blank_scope(immediate_decl_count(imported_page.top_scope) + 1)
			r.write_collective_to_scope(spindle, page, imported_page.top_scope)

			r.push_string_to_scope(new_hash("path"), find_text) // @todo why is this design choice buried here

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
				// this means before any content has been discovered
				// only declarations can come before this
				if t.has_body && buffer.Len() == 0 {
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

					// with the change to using buffer.len instead of index == 1
					// we now slice off by the index because everything above is
					// already on scope — it's... _funky_... but it works.
					r.write_collective_to_scope(spindle, page, input[index:])
					r.push_anon(input[index:], t.content, t.position)

					buffer.WriteString(r.render_ast(spindle, page, t.content))

					if did_push {
						r.pop_scope()
					}
					return buffer.String() // hard exit
				}

				// if we're not the first, we just pull in the
				// declarations from the template to be used from
				// here on out in this scope
				r.write_collective_to_scope(spindle, page, t.top_scope)

			} else if t, ok := r.get_in_scope(entry.hash_name); ok {
				x := t.get_children()

				if len(x) == 1 && x[0].type_check() == BLOCK {
					r.write_collective_to_scope(spindle, page, x[0].get_children())
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
					// @error
					panic("popped anon was missing!")
				}

				popped_anon.anon_count -= 1

				if popped_anon.anon_count <= 0 {
					r.pop_anon()
				}

				text = r.render_ast(spindle, page, popped_anon.children)

			} else if found, ok := r.get_in_scope(entry.field); ok {
				text = r.render_ast(spindle, page, found.get_children())
			}

			if entry.ast_type == VAR_ENUM && entry.subname > 0 {
				args := unix_args(text)
				n    := int(entry.subname)

				if n > len(args) {
					spindle.errors.new_pos(
						RENDER_WARNING, popped_anon.position,
						"input token only supplies %d arguments\n    %s — line %d\n    needs %d arguments.",
						len(args), entry.position.file_path, entry.position.line, n,
					)
					text = ""
				} else {
					text = args[n - 1]
				}
			}

			if entry.modifier > NONE {
				text = apply_modifier(r, text, entry.modifier)
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
			found_file, ok := render_find_file(spindle, page, find_text)
			if ok {
				if !spindle.server_mode && !spindle.build_drafts && found_file.is_draft {
					spindle.errors.new_pos(RENDER_WARNING, entry.position, "%q is a draft", found_file.path)
				}

				the_url := ""

				if entry.path_type == NO_PATH_TYPE {
					entry.path_type = spindle.path_mode
				}

				switch entry.finder_type {
				case _IMAGE:
					if !(found_file.file_type > is_image && found_file.file_type < end_image) {
						spindle.errors.new_pos(RENDER_FAILURE, entry.position, "resource finder cannot handle non-image %q", find_text)
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
								spindle.gen_images[hash] = &Image{
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

					the_url = make_page_url(spindle, &found_file.file_info, entry.path_type, page.page_path)
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

			if page.import_hash > 0 {
				if r.skip_import_condition(spindle, page, x) {
					continue
				}
			}

			if entry.decl_hash > 0 {
				wrapper_block, ok := r.get_in_scope(entry.decl_hash)

				if ok && wrapper_block.ast_type == DECL_BLOCK {
					children := wrapper_block.get_children()
					did_push := r.push_blank_scope(immediate_decl_count(children) * 2)

					r.push_anon(x, children, *entry.get_position())
					r.write_collective_to_scope(spindle, page, x)

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
			// @todo optimisation needed to stop looking up
			//  everything we need 800 times

			entry := entry.(*ast_token)

			x := entry.get_children()

			did_push := false
			wrapper_block, ok := r.get_in_scope(entry.decl_hash)

			if ok && wrapper_block.ast_type == DECL_TOKEN {
				did_push = r.push_blank_scope(immediate_decl_count(wrapper_block.get_children()))
			} else {
				if len(x) == 0 {
					wrapper_block, ok := r.get_in_scope(DEFAULT_HASH)
					if ok {
						children := wrapper_block.get_children()
						did_push := r.push_blank_scope(immediate_decl_count(children))

						r.push_anon(x, children, *entry.get_position())

						buffer.WriteString(apply_regex_array(spindle.inline, r.render_ast(spindle, page, children)))

						if did_push {
							r.pop_scope()
						}
						continue
					}

				} else {
					if entry.decl_hash != STOP_HASH {
						spindle.errors.new_pos(RENDER_WARNING, entry.position, "token %q does not have a template — output may be unexpected unless it is escaped", entry.orig_field)
					}

					buffer.WriteString(apply_regex_array(spindle.inline, r.render_ast(spindle, page, x)))
					continue
				}
			}

			// @todo scopes won't expand from the wrapper
			group_block, has_group := r.get_in_scope(entry.decl_hash + 1)

			// this can only be DECL_TOKEN from the parser, so we don't check
			if has_group {
				sub_buffer := strings.Builder{}
				sub_buffer.Grow(512)

				index -= 1 // step back one to get the original again

				for _, sub := range input[index:] {
					if sub.type_check() == TOKEN {
						sub := sub.(*ast_token)
						if sub.decl_hash == entry.decl_hash {
							children := wrapper_block.get_children()
							r.push_anon(sub.get_children(), children, *entry.get_position())
							sub_buffer.WriteString(apply_regex_array(spindle.inline, r.render_ast(spindle, page, children)))
							index++
							continue
						}
					}
					break
				}

				rendered_block := sub_buffer.String()
				sub_buffer.Reset()
				new_inner := []ast_data{&ast_base{ ast_type: NORMAL, field: rendered_block }}

				r.push_anon(new_inner, group_block.get_children(), *entry.get_position())

				// expand here
				buffer.WriteString(r.render_ast(spindle, page, group_block.get_children()))

				if did_push {
					r.pop_scope()
				}
				continue
			}

			children := wrapper_block.get_children()
			r.push_anon(x, children, *entry.get_position())
			buffer.WriteString(apply_regex_array(spindle.inline, r.render_ast(spindle, page, children)))

			if did_push {
				r.pop_scope()
			}

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

			for i, t := range array {
				r.push_string_to_scope(IT_HASH, t)

				if i == len(array) - 1 {
					r.push_string_to_scope(1787721130, "") // "end" @todo
				}

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
				wrapper_block, ok := r.get_in_scope(DEFAULT_HASH)
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

func apply_modifier(renderer *renderer, text string, modifier ast_modifier) string {
	switch modifier {
	case SLUG:
		text = make_slug(text)
	case UNIQUE_SLUG:
		text = make_slug(text)
		if n, ok := renderer.slug_tracker[text]; ok {
			renderer.slug_tracker[text] = n + 1
			text = fmt.Sprintf("%s-%d", text, n)
		} else {
			renderer.slug_tracker[text] = 1
		}
	case TITLE:
		text = make_title(text)
	case UPPER:
		text = strings.ToUpper(text)
	case LOWER:
		text = strings.ToLower(text)
	}
	return text
}

func render_find_file(spindle *spindle, page *Page, search_term string) (*File, bool) {
	found_file, ok := spindle.finder_cache[search_term]
	if !ok {
		return find_file(spindle.file_tree, search_term)
	}
	return found_file, true
}