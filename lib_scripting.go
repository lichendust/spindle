package main

import "github.com/dop251/goja"

var the_vm *goja.Runtime

func init() {
	the_vm = goja.New()

	the_vm.SetFieldNameMapper(goja.UncapFieldNameMapper())

	the_vm.Set("console", _println)
	the_vm.Set("modifier", map[string]ast_modifier {
		"slug":        SLUG,
		"unique_slug": UNIQUE_SLUG,
		"upper":       UPPER,
		"lower":       LOWER,
		"title":       TITLE,
	})
	the_vm.Set("text_truncate", truncate)
}

func (r *renderer) script_call(spindle *spindle, page *page_object, line int, exec_blob string, args ...string) (string, bool) {
	slug_tracker := make(map[string]uint, 16) // @todo should be unique per-page, not per script-call

	the_vm.Set("_line", line)
	the_vm.Set("args", args)

	the_vm.Set("text_modifier", func(text string, mod ast_modifier) string {
		return apply_modifier(slug_tracker, text, mod)
	})

	the_vm.Set("get", func(name string) (string, bool) {
		if entry, ok := r.get_in_scope(new_hash(name)); ok && entry.ast_type == DECL {
			return r.render_ast(spindle, page, entry.get_children()), true
		}
		return "", false
	})
	the_vm.Set("get_token", func(depth int, match ...string) []script_token {
		h := make([]uint32, len(match))
		for i, n := range match {
			h[i] = new_hash(n)
		}
		return r.script_get_tokens_as_strings(spindle, page, page.content, depth, h...)
	})
	the_vm.Set("has_elements", func(match ...string) bool {
		h := make([]uint32, len(match))
		for i, n := range match {
			h[i] = new_hash(n)
		}
		return r.script_has_elements(spindle, page, page.content, h...)
	})
	the_vm.Set("find_file", func(find_text string) string {
		found_file, ok := spindle.finder_cache[find_text]

		if !ok {
			found_file, ok = find_file(spindle.file_tree, find_text)

			if ok {
				spindle.finder_cache[find_text] = found_file
			}
		}

		if ok {
			found_file.is_used = true

			tc := found_file.file_type
			dp := spindle.default_path_mode
			pp := page.page_path

			if tc > is_page && tc < end_page {
				return make_page_url(spindle, &found_file.anon_file_info, dp, pp)
			}
			if tc > is_image && tc < end_image {
				return make_general_url(spindle, found_file, dp, pp)
			}
			if tc > is_static && tc < end_static {
				return make_general_url(spindle, found_file, dp, pp)
			}
		}

		return ""
	})

	v, err := the_vm.RunString(`function _(){` + exec_blob + `};_()`)
	if err != nil {
		_println(err)
		return "", false
	}

	if the_value := v.Export(); the_value != nil {
		return the_value.(string), true
	}
	return "", true
}

type script_token struct {
	Token string
	Text  string
	Line  int
}

func (r *renderer) script_get_tokens_as_strings(spindle *spindle, page *page_object, input []ast_data, depth int, match ...uint32) []script_token {
	if depth == 0 {
		return []script_token{}
	}

	array := make([]script_token, 0, len(input))

	for _, entry := range input {
		if entry.type_check() == TOKEN {
			token := entry.(*ast_token)

			for _, h := range match {
				if token.decl_hash == h {
					array = append(array, script_token{
						Token: token.orig_field,
						Text:  r.render_ast(spindle, page, token.get_children()),
						Line:  token.position.line,
					})
				}
			}

			continue
		}
		if depth <= 0 {
			continue
		}
		if x := entry.get_children(); len(x) > 0 {
			sub := r.script_get_tokens_as_strings(spindle, page, x, depth - 1, match...)
			for _, s := range sub {
				array = append(array, s)
			}
		}
	}

	return array
}

func (r *renderer) script_has_elements(spindle *spindle, page *page_object, input []ast_data, match ...uint32) bool {
	for _, entry := range input {
		tc := entry.type_check()

		if tc == TOKEN {
			token := entry.(*ast_token)
			for _, h := range match {
				if token.decl_hash == h {
					return true
				}
			}
			continue
		}

		if tc == BLOCK {
			block := entry.(*ast_block)
			for _, h := range match {
				if block.decl_hash == h {
					return true
				}
			}
		}

		if x := entry.get_children(); len(x) > 0 {
			if r.script_has_elements(spindle, page, x, match...) {
				return true
			}
		}
	}

	return false
}