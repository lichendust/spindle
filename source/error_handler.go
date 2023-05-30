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

import (
	"fmt"
	"strings"
)

type Error_Type uint8
const (
	FAILURE Error_Type = iota
	RENDER_FAILURE
	PARSER_FAILURE
)

const (
	ERR_HTML uint8 = iota
	ERR_TERM
)

func (e Error_Type) String() string {
	switch e {
	case FAILURE:
		return "Error"
	case RENDER_FAILURE:
		return "Render"
	case PARSER_FAILURE:
		return "Parser"
	}
	return ""
}

type Spindle_Error struct {
	kind    Error_Type
	pos     position
	has_pos bool
	message string
}

func html_string(e *Spindle_Error) string {
	if e.has_pos {
		const t_error_html = `<section><p><b>%s</b></p><p class="space"><tt>Line %d — %s</tt></p><p>%s</p></section>`
		return fmt.Sprintf(t_error_html, e.kind, e.pos.line, e.pos.file_path, e.message)
	}

	const t_error_html = `<section><p><b>%s!</b></p><p>%s</p></section>`
	return fmt.Sprintf(t_error_html, e.kind, e.message)
}

func term_string(e *Spindle_Error) string {
	if e.has_pos {
		const t_error_term = "[%s] Line %d — %s\n    %s"
		return fmt.Sprintf(t_error_term, e.kind, e.pos.line, e.pos.file_path, e.message)
	}

	const t_error_term = "[%s]\n    %s"
	return fmt.Sprintf(t_error_term, e.kind, e.message)
}

func new_error_handler() *Error_Handler {
	e := new(Error_Handler)
	e.reset()
	return e
}

type Error_Handler struct {
	has_failures bool
	all_errors   []*Spindle_Error
}

func (e *Error_Handler) reset() {
	e.has_failures = false
	e.all_errors   = make([]*Spindle_Error, 0, 8)
}

func (e *Error_Handler) new_pos(kind Error_Type, pos position, message string, subst ...any) {
	e.has_failures = true
	e.all_errors = append(e.all_errors, &Spindle_Error{
		kind,
		pos,
		true,
		fmt.Sprintf(message, subst...),
	})
}

func (e *Error_Handler) new(kind Error_Type, message string, subst ...any) {
	e.has_failures = true
	e.all_errors = append(e.all_errors, &Spindle_Error{
		kind,
		position{},
		false,
		fmt.Sprintf(message, subst...),
	})
}

func (e *Error_Handler) has_errors() bool {
	return len(e.all_errors) > 0
}

func (e *Error_Handler) render_errors(error_function uint8) string {
	dedup := make(map[uint32]bool, len(e.all_errors))

	buffer := strings.Builder{}
	buffer.Grow(len(e.all_errors) * 128)

	for _, the_error := range e.all_errors {
		var x string

		switch error_function {
		case ERR_HTML: x = html_string(the_error)
		case ERR_TERM: x = term_string(the_error)
		}

		n := new_hash(x)

		if dedup[n] {
			continue
		}

		dedup[n] = true

		write_to(&buffer, x, "\n\n")
	}

	if error_function == ERR_HTML {
		return fmt.Sprintf(ERROR_PAGE, buffer.String())
	}

	return strings.TrimSpace(buffer.String())
}

func error_page_not_found() string {
	return fmt.Sprintf(ERROR_PAGE, "<section><p><b>Page not found...</b></p></section>")
}

const ERROR_PAGE = `<!DOCTYPE html>
<html>
	<head>
		<meta charset="utf-8">
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<title>Spindle</title>
		<link rel="stylesheet" type="text/css" href="/_spindle/manual/style.css"/>
		` + RELOAD_SCRIPT + `</head>
<body>
	<h1>` + SPINDLE + `</h1>
	<main>
		%s
	</main>
	<aside>
		<p><b>Resources</b></p>
		<ul>
			<li><a href="/_spindle/manual">Manual</a></li>
			<li><a href="https://github.com/qxoko/spindle">GitHub</a></li>
		</ul>
	</aside>
	<br clear="all">
</body></html>`