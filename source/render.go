package main

import (
	"fmt"
	"time"
	"sort"
	"strings"
)

type Renderer struct {
	unwind bool

	scope_offset []int
	scope_stack  []*AST_Declare
	anon_stack   []*anon_entry

	slug_tracker map[string]uint
}

func render_syntax_tree(spindle *Spindle, page *Page) string {
	r := new(Renderer)

	r.scope_offset = []int{0}
	r.scope_stack  = make([]*AST_Declare, 0, 64)
	r.anon_stack   = make([]*anon_entry, 0, 4)
	r.slug_tracker = make(map[string]uint, 16)

	if spindle.server_mode {
		r.push_string_to_scope(_IS_SERVER, "") // just has to exist
		r.push_string_to_scope(_RELOAD_SCRIPT, reload_script)
	}

	canonical_path := make_page_url(spindle, &page.file.File_Info, ABSOLUTE, "")
	url_path       := make_page_url(spindle, &page.file.File_Info, spindle.path_mode, "")

	if page.import_hash > 0 {
		r.push_string_to_scope(_TAGINATOR_ACTIVE, "") // just has to exist
		r.push_string_to_scope(_TAGINATOR_TAG, page.import_cond)
		r.push_string_to_scope(_PAGE_CANONICAL, tag_path(canonical_path, spindle.tag_path, page.import_cond))
		r.push_string_to_scope(_PAGE_URL, tag_path(url_path, spindle.tag_path, page.import_cond))
	} else {
		r.push_string_to_scope(_PAGE_CANONICAL, canonical_path)
		r.push_string_to_scope(_PAGE_URL, url_path)
	}

	// taginator parent url falls back to own url so it can be used regardless
	r.push_string_to_scope(_TAGINATOR_PARENT, make_page_url(spindle, &page.file.File_Info, spindle.path_mode, ""))

	// spindle.current_year
	r.push_string_to_scope(519639417, fmt.Sprintf("%d", time.Now().Year()))

	return r.render_ast(spindle, page, page.content)
}

type anon_entry struct {
	anon_count int
	position   position
	children   []AST_Data
}

func (r *Renderer) get_anon() *anon_entry {
	if len(r.anon_stack) > 0 {
		return r.anon_stack[len(r.anon_stack) - 1]
	}
	return nil
}

func (r *Renderer) pop_anon() {
	if len(r.anon_stack) > 0 {
		r.anon_stack = r.anon_stack[:len(r.anon_stack) - 1]
	}
}

func (r *Renderer) push_anon(content, wrapper []AST_Data, pos position) {
	stack_entry := &anon_entry{
		anon_count: recursive_anon_count(wrapper),
		children:   content,
		position:   pos,
	}
	r.anon_stack = append(r.anon_stack, stack_entry)
}

/*
	declaration scope stack
*/
func (r *Renderer) start_scope(alloc int) bool {
	if alloc == 0 {
		return false
	}
	r.scope_offset = append(r.scope_offset, 0)
	return true
}

func (r *Renderer) get_in_scope(value uint32) (*AST_Declare, bool) {
	for i := len(r.scope_stack) - 1; i >= 0; i -= 1 {
		x := r.scope_stack[i]

		if x.field == value {
			if x.ast_type == DECL_REJECT {
				break
			}
			return x, true
		}
	}

	return nil, false
}

func (r *Renderer) push_to_scope(entry *AST_Declare) {
	if entry.is_soft {
		x, ok := r.get_in_scope(entry.field)
		if ok && !x.is_soft {
			return
		}
	}

	r.scope_stack = append(r.scope_stack, entry)
	r.scope_offset[len(r.scope_offset) - 1] += 1
}

func (r *Renderer) push_string_to_scope(ident uint32, text string) {
	decl := new(AST_Declare)

	decl.ast_type = DECL
	decl.field    = ident

	decl.children = []AST_Data{
		&AST_Base{
			ast_type: NORMAL,
			field:    text,
		},
	}

	r.push_to_scope(decl)
}

func (r *Renderer) push_collective_to_scope(spindle *Spindle, page *Page, input []AST_Data) {
	for _, entry := range input {
		_type := entry.type_check()

		if _type.is(DECL, DECL_TOKEN, DECL_BLOCK) {
			entry := entry.(*AST_Declare)
			r.push_to_scope(entry)
			continue
		}

		if _type == TEMPLATE {
			entry := entry.(*AST_Builtin)

			if t, ok := spindle.templates[entry.hash_name]; ok {
				r.push_collective_to_scope(spindle, page, t.top_scope)
			} else if t, ok := r.get_in_scope(entry.hash_name); ok {
				x := t.get_children()

				if len(x) == 1 && x[0].type_check() == BLOCK {
					r.push_collective_to_scope(spindle, page, x[0].get_children())
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

func (r *Renderer) pop_scope() {
	x := r.scope_offset[len(r.scope_offset) - 1]

	r.scope_offset = r.scope_offset[:len(r.scope_offset) - 1]
	r.scope_stack  = r.scope_stack[:len(r.scope_stack) - x]
}

func (r *Renderer) delete_scope_entry(value uint32) {
	x := AST_Declare{
		field:    value,
		ast_type: DECL_REJECT,
	}
	y := AST_Declare{
		field:    value + 1,
		ast_type: DECL_REJECT,
	}

	r.push_to_scope(&x)
	r.push_to_scope(&y)
}

/*
	if statements
*/
func (r *Renderer) evaluate_if(entry *AST_If) bool {
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

func (r *Renderer) skip_import_condition(spindle *Spindle, page *Page, data_slice []AST_Data) bool {
	for _, entry := range data_slice {
		if entry.type_check() == DECL {
			entry := entry.(*AST_Declare)

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

func __recurse(r *Renderer, spindle *Spindle, page *Page, input []AST_Data, target_hash uint32) map[string]bool {
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
			entry := entry.(*AST_Declare)

			if entry.field == target_hash {
				fields := unix_args(r.render_ast(spindle, page, entry.children))

				for _, tag := range fields {
					tag = strings.ToLower(tag)

					capture[tag] = true

					file_path := tag_path(make_page_url(spindle, &page.file.File_Info, NO_PATH_TYPE, ""), spindle.tag_path, tag)

					if _, ok := spindle.gen_pages[file_path]; ok {
						continue
					}

					/*
						@todo this is the wrong structure and causes cache fails
						we need to reference (not copy) the original page
					*/

					copy := new(Gen_Page)

					copy.file        = page.file
					copy.import_cond = tag
					copy.import_hash = target_hash

					spindle.gen_pages[file_path] = copy
				}
			}

		case IMPORT:
			entry := entry.(*AST_Builtin)

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

func (r *Renderer) do_import_seek(spindle *Spindle, page *Page, target_hash uint32) {
	capture := __recurse(r, spindle, page, page.content, target_hash)

	array := make([]string, 0, len(capture))

	for x := range capture {
		array = append(array, x) // @todo make sure strings get requoted if needed
	}

	sort.Strings(array)

	r.push_string_to_scope(_TAGINATOR_ALL, strings.Join(array, " "))
}

/*func (r *Renderer) do_import_seek(spindle *Spindle, page *Page, data_slice []AST_Data) {
	for _, entry := range data_slice {
		if entry.type_check() == DECL {
			entry := entry.(*AST_Declare)

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

func (r *Renderer) render_ast(spindle *Spindle, page *Page, input []AST_Data) string {
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
			entry := entry.(*AST_Declare)
			r.push_to_scope(entry)

			// if we find a taginator in a scope
			if tc == DECL && entry.field == _TAGINATOR {
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
			entry := entry.(*AST_Script)

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
			entry := entry.(*AST_Builtin)
			r.delete_scope_entry(entry.hash_name)
			r.delete_scope_entry(entry.hash_name + 1)

		case PARTIAL:
			entry := entry.(*AST_Builtin)

			p, ok := spindle.partials[entry.hash_name]

			if !ok {
				spindle.errors.new_pos(RENDER_FAILURE, entry.position, "failed to load partial %q", get_hash(entry.hash_name))
				r.unwind = true
				return ""
			}

			needs_pop := r.start_scope(immediate_decl_count(p.content))
			buffer.WriteString(r.render_ast(spindle, page, p.content))

			if needs_pop { r.pop_scope() }

		case IMPORT:
			entry := entry.(*AST_Builtin)

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

			r.start_scope(immediate_decl_count(imported_page.top_scope) + 1)
			r.push_collective_to_scope(spindle, page, imported_page.top_scope)

			// @todo we're copying the now-cached finder text
			// rather than the _actual_ path, which is not
			// semantically correct, but does achieve the same result
			// for the primary use-case
			r.push_string_to_scope(_IMPORT_PATH, find_text)

			// @todo undefined behaviour for %% in imports
			// we should probably disallow it but we can't know until
			// we've parsed it

			buffer.WriteString(r.render_ast(spindle, page, t.get_children()))

			r.pop_scope()

			spindle.finder_cache[find_text] = found_file

		case TEMPLATE:
			entry := entry.(*AST_Builtin)

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
					needs_pop := r.start_scope(immediate_decl_count(t.content))

					// with the change to using buffer.len instead of index == 1
					// we now slice off by the index because everything above is
					// already on scope — it's... _funky_... but it works.
					r.push_collective_to_scope(spindle, page, input[index:])
					r.push_anon(input[index:], t.content, t.position)

					buffer.WriteString(r.render_ast(spindle, page, t.content))

					if needs_pop {
						r.pop_scope()
					}
					return buffer.String() // hard exit
				}

				// if we're not the first, we just pull in the
				// declarations from the template to be used from
				// here on out in this scope
				r.push_collective_to_scope(spindle, page, t.top_scope)

			} else if t, ok := r.get_in_scope(entry.hash_name); ok {
				x := t.get_children()

				if len(x) == 1 && x[0].type_check() == BLOCK {
					r.push_collective_to_scope(spindle, page, x[0].get_children())
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
				text = apply_modifier(r.slug_tracker, text, entry.modifier)
			}

			buffer.WriteString(text)

		case RES_FINDER:
			entry := entry.(*AST_Finder)

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

				switch entry.Finder_Type {
				case _IMAGE:
					if !(found_file.file_type > is_image && found_file.file_type < end_image) {
						spindle.errors.new_pos(RENDER_FAILURE, entry.position, "resource finder cannot handle non-image %q", find_text)
						r.unwind = true
					}

					if entry.Image_Settings != nil {
						settings := *entry.Image_Settings
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

					the_url = make_page_url(spindle, &found_file.File_Info, entry.path_type, page.page_path)
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
			entry := entry.(*AST_Block)

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
					needs_pop := r.start_scope(immediate_decl_count(children) * 2)

					r.push_anon(x, children, *entry.get_position())
					r.push_collective_to_scope(spindle, page, x)

					buffer.WriteString(r.render_ast(spindle, page, children))

					if needs_pop {
						r.pop_scope()
					}
					continue
				}
			}

			// else:
			needs_pop := r.start_scope(immediate_decl_count(x))
			buffer.WriteString(r.render_ast(spindle, page, x))
			if needs_pop {
				r.pop_scope()
			}

		case TOKEN:
			// @todo optimisation needed to stop looking up
			//  everything we need 800 times

			entry := entry.(*AST_Token)

			x := entry.get_children()

			needs_pop := false
			wrapper_block, ok := r.get_in_scope(entry.decl_hash)

			if ok && wrapper_block.ast_type == DECL_TOKEN {
				needs_pop = r.start_scope(immediate_decl_count(wrapper_block.get_children()))
			} else {
				if len(x) == 0 {
					wrapper_block, ok := r.get_in_scope(_DEFAULT)
					if ok {
						children := wrapper_block.get_children()
						needs_pop := r.start_scope(immediate_decl_count(children))

						r.push_anon(x, children, *entry.get_position())

						buffer.WriteString(apply_regex_array(spindle.inline, r.render_ast(spindle, page, children)))

						if needs_pop {
							r.pop_scope()
						}
						continue
					}

				} else {
					if entry.decl_hash != _STOP {
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
						sub := sub.(*AST_Token)
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
				new_inner := []AST_Data{&AST_Base{ ast_type: NORMAL, field: rendered_block }}

				r.push_anon(new_inner, group_block.get_children(), *entry.get_position())

				// expand here
				buffer.WriteString(r.render_ast(spindle, page, group_block.get_children()))

				if needs_pop {
					r.pop_scope()
				}
				continue
			}

			children := wrapper_block.get_children()
			r.push_anon(x, children, *entry.get_position())
			buffer.WriteString(apply_regex_array(spindle.inline, r.render_ast(spindle, page, children)))

			if needs_pop {
				r.pop_scope()
			}

		case CONTROL_IF:
			entry := entry.(*AST_If)

			if !r.evaluate_if(entry) {
				continue
			}

			the_block := entry.get_children()[0].(*AST_Block)

			x := the_block.get_children()

			if the_block.decl_hash > 0 {
				wrapper_block, ok := r.get_in_scope(the_block.decl_hash)
				if ok {
					children := wrapper_block.get_children()

					needs_pop := r.start_scope(immediate_decl_count(children))
					r.push_anon(x, children, *entry.get_position())

					buffer.WriteString(r.render_ast(spindle, page, children))

					if needs_pop {
						r.pop_scope()
					}
					continue
				}
			}

			// else:
			needs_pop := r.start_scope(immediate_decl_count(x))
			buffer.WriteString(r.render_ast(spindle, page, x))
			if needs_pop { r.pop_scope() }

		case CONTROL_FOR:
			entry := entry.(*AST_For)
			array := unix_args(r.render_ast(spindle, page, []AST_Data{entry.iterator_source}))

			if len(array) == 0 {
				continue
			}

			the_block := entry.get_children()[0].(*AST_Block)

			needs_pop := r.start_scope(immediate_decl_count(the_block.get_children()))

			sub_buffer := strings.Builder{}
			sub_buffer.Grow(512)

			for i, t := range array {
				r.push_string_to_scope(_IT, t)

				if i == len(array) - 1 {
					r.push_string_to_scope(1787721130, "") // "end" @todo
				}

				sub_buffer.WriteString(r.render_ast(spindle, page, the_block.get_children()))
			}

			if the_block.decl_hash > 0 {
				wrapper_block, ok := r.get_in_scope(the_block.decl_hash)
				if ok {
					r.push_anon([]AST_Data{&AST_Base{ast_type:NORMAL,field:sub_buffer.String()}}, wrapper_block.get_children(), *entry.get_position())
					buffer.WriteString(r.render_ast(spindle, page, wrapper_block.get_children()))
				}
			} else {
				buffer.WriteString(sub_buffer.String())
			}

			sub_buffer.Reset()

			if needs_pop {
				r.pop_scope()
			}

		case NORMAL:
			entry := entry.(*AST_Base)

			x := entry.get_children()

			if len(x) > 0 {
				wrapper_block, ok := r.get_in_scope(_DEFAULT)
				if ok {
					children := wrapper_block.get_children()
					needs_pop := r.start_scope(immediate_decl_count(children))

					r.push_anon(x, children, *entry.get_position())

					buffer.WriteString(apply_regex_array(spindle.inline, r.render_ast(spindle, page, children)))

					if needs_pop { r.pop_scope() }
					continue
				}

				// else:
				buffer.WriteString(r.render_ast(spindle, page, x))
			} else {
				buffer.WriteString(entry.field)
			}

		case RAW:
			entry := entry.(*AST_Base)

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

func apply_modifier(slug_tracker map[string]uint, text string, modifier AST_Modifier) string {
	switch modifier {
	case SLUG:
		text = make_slug(text)
	case UNIQUE_SLUG:
		text = make_slug(text)
		if n, ok := slug_tracker[text]; ok {
			slug_tracker[text] = n + 1
			text = fmt.Sprintf("%s-%d", text, n)
		} else {
			slug_tracker[text] = 1
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

func render_find_file(spindle *Spindle, page *Page, search_term string) (*File, bool) {
	found_file, ok := spindle.finder_cache[search_term]
	if !ok {
		return find_file(spindle.file_tree, search_term)
	}
	return found_file, true
}