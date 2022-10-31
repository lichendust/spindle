//go:build debug

package main

import (
	"fmt"
	"io"
	"strings"
	"encoding/xml"
	// "path/filepath"

	"github.com/yosssi/gohtml"
)

func _println(name ...any) {
	fmt.Println(name...)
}

func print_token_stream(array []*lexer_token) {
	for _, entry := range array {
		fmt.Println(entry)
	}
}

func (t *lexer_token) String() string {
	return fmt.Sprintf("%-3d [%d:%d] %s %q", t.position.line, t.position.start, t.position.end, t.ast_type, t.field)
}

func print_syntax_tree(array []ast_data, level int) {
	indent := strings.Repeat("    ", level)

	for _, entry := range array {
		_type := entry.type_check()
		pos := entry.get_position()

		fmt.Print(indent)
		fmt.Print(_type)
		fmt.Print(" ", pos.start, pos.end)

		switch _type {
		case NORMAL:
			cast := entry.(*ast_base)
			fmt.Print(" ", cast.field)

		case VAR, VAR_ENUM, VAR_ANON:
			cast := entry.(*ast_variable)
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

func print_file_tree(array []*disk_object, level int) {
	indent := strings.Repeat("    ", level)

	for _, d := range array {
		fmt.Print(indent)
		fmt.Println(d)

		if len(d.children) > 0 {
			print_file_tree(d.children, level + 1)
		}
	}
}

func (d *disk_object) String() string {
	// return fmt.Sprint(d.file_type, " ", filepath.Base(d.path))
	return fmt.Sprint(d.file_type, " ", d.path, " ", d.is_used)
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

func format_html(input string) string {
	return gohtml.Format(input)
}