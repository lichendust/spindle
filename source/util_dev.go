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

//go:build debug

package main

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"encoding/xml"
	// "path/filepath"
)

func _println(name ...any) {
	fmt.Println(name...)
}

func print_token_stream(array []*Lexer_Token) {
	for _, entry := range array {
		fmt.Println(entry)
	}
}

func (x Position) String() string {
	return fmt.Sprintf("%-3d [%d:%d]", x.line, x.start, x.end)
}

func (t *Lexer_Token) String() string {
	return fmt.Sprintf("%-3d [%d:%d] %s %q", t.position.line, t.position.start, t.position.end, t.ast_type, t.field)
}

func print_syntax_tree(array []AST_Data, level int) {
	indent := strings.Repeat("    ", level)

	for _, entry := range array {
		_type := entry.type_check()
		pos := entry.get_position()

		fmt.Print(indent)
		fmt.Print(_type)
		fmt.Print(" ", pos.start, pos.end)

		switch _type {
		case NORMAL:
			cast := entry.(*AST_Base)
			fmt.Print(" ", cast.field)

		case DECL, DECL_TOKEN, DECL_BLOCK:
			cast := entry.(*AST_Declare)
			fmt.Print(" ", get_hash(cast.field))

		case VAR, VAR_ENUM, VAR_ANON:
			cast := entry.(*AST_Variable)
			fmt.Print(" ", get_hash(cast.field))

		case RES_FINDER:
			fmt.Print(" ", len(entry.get_children()))
		}

		fmt.Print("\n")

		if x := entry.get_children(); len(x) > 0 {
			print_syntax_tree(x, level + 1)
		}
	}
}

func print_file_tree(array []*File, level int) {
	indent := strings.Repeat("    ", level)

	for _, d := range array {
		fmt.Print(indent)
		fmt.Println(d)

		if len(d.children) > 0 {
			print_file_tree(d.children, level + 1)
		}
	}
}

func (d *File) String() string {
	// return fmt.Sprint(d.file_type, " ", filepath.Base(d.path))
	return fmt.Sprint(d.file_type, " ", d.path, " ", d.is_used, " ", d.is_built)
}

// this is a *bad* implementation of HTML
// validation, but we're just using it
// in development to catch dumb errors
// if/when they happen
func validate_html(input string) bool {
	render  := strings.NewReader(input)
	decoder := xml.NewDecoder(render)

	decoder.Strict    = false
	decoder.AutoClose = xml.HTMLAutoClose
	decoder.Entity    = xml.HTMLEntity

	for {
		_, err := decoder.Token()
		switch err {
		case nil:
		case io.EOF:
			return true  // all good
		default:
			return false // found error
		}
	}

	return false
}

var has_printed_scope = make(map[string]bool, 10)

func print_scope_stack(location string, scope_stack []map[uint32]*AST_Declare) {
	if has_printed_scope[location] {
		return
	}

	fmt.Println("\n\n", location)

	for i, level := range scope_stack {
		array := make([]string, 0, len(level))

		for id, _ := range level {
			array = append(array, get_hash(id))
		}

		fmt.Printf("[level %d]\n", i)

		sort.Strings(array)

		for _, entry := range array {
			fmt.Println(entry)
		}
	}

	has_printed_scope[location] = true
}