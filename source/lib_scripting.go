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

import "github.com/dop251/goja"

var the_vm *goja.Runtime

func init() {
	the_vm = goja.New()

	the_vm.SetFieldNameMapper(goja.UncapFieldNameMapper())

	the_vm.Set("console", println)
	the_vm.Set("modifier", map[string]AST_Modifier {
		"slug":        SLUG,
		"unique_slug": UNIQUE_SLUG,
		"upper":       UPPER,
		"lower":       LOWER,
		"title":       TITLE,
	})
	the_vm.Set("text_truncate", truncate)
	the_vm.Set("current_date", nsdate)
}

func (r *Renderer) script_call(spindle *Spindle, page *Page, line int, exec_blob string, args ...string) (string, bool) {
	slug_tracker := make(map[string]uint, 8) // @todo shouldn't be per-call

	the_vm.Set("_line", line)
	the_vm.Set("args", args)

	the_vm.Set("text_modifier", func(text string, mod AST_Modifier) string {
		return apply_modifier(slug_tracker, text, mod)
	})

	the_vm.Set("get", func(name string) string {
		if entry, ok := r.get_in_scope(new_hash(name)); ok && entry.ast_type == DECL {
			return r.render_ast(spindle, page, entry.get_children())
		}
		return ""
	})

	the_vm.Set("get_token", func(depth int, match ...string) []Script_Token {
		h := make([]uint32, len(match))
		for i, n := range match {
			h[i] = new_hash(n)
		}
		return r.script_get_tokens_as_strings(spindle, page, page.content, depth, h...)
	})

	the_vm.Set("get_array", func(name string) []string {
		if entry, ok := r.get_in_scope(new_hash(name)); ok && entry.ast_type == DECL {
			return unix_args(r.render_ast(spindle, page, entry.get_children()))
		}
		return []string{}
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
			dp := spindle.path_mode
			pp := page.page_path

			if tc > is_page && tc < end_page {
				return make_page_url(spindle, &found_file.File_Info, dp, pp)
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
		spindle.errors.new(FAILURE, err.Error())
		return "", false
	}

	if the_value := v.Export(); the_value != nil {
		return the_value.(string), true
	}
	return "", true
}

type Script_Token struct {
	Token string
	Text  string
	Line  int
}

func (r *Renderer) script_get_tokens_as_strings(spindle *Spindle, page *Page, input []AST_Data, depth int, match ...uint32) []Script_Token {
	if depth == 0 {
		return []Script_Token{}
	}

	array := make([]Script_Token, 0, len(input))

	for _, entry := range input {
		if entry.type_check() == TOKEN {
			token := entry.(*AST_Token)

			for _, h := range match {
				if token.decl_hash == h {
					array = append(array, Script_Token{
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
			array = append(array, sub...)
		}
	}

	return array
}

func (r *Renderer) script_has_elements(spindle *Spindle, page *Page, input []AST_Data, match ...uint32) bool {
	for _, entry := range input {
		tc := entry.type_check()

		if tc == TOKEN {
			token := entry.(*AST_Token)
			for _, h := range match {
				if token.decl_hash == h {
					return true
				}
			}
			continue
		}

		if tc == BLOCK {
			block := entry.(*AST_Block)
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