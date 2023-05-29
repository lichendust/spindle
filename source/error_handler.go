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

type _error interface {
	html_string() string
	term_string() string
}

type Error_Type uint8
const (
	WARNING Error_Type = iota
	RENDER_WARNING
	PARSER_WARNING

	is_failure
	FAILURE
	RENDER_FAILURE
	PARSER_FAILURE
)

func (e Error_Type) String() string {
	switch e {
	case WARNING:
		return "Warning"
	case RENDER_WARNING:
		return "Render Warning"
	case PARSER_WARNING:
		return "Parser Warning"
	case FAILURE:
		return "Failure"
	case RENDER_FAILURE:
		return "Render Failure"
	case PARSER_FAILURE:
		return "Parser Failure"
	}
	return ""
}

type Spindle_Pos_Error struct {
	kind    Error_Type
	pos     position
	message string
}

func (e *Spindle_Pos_Error) html_string() string {
	const t_error_html = `<section><p><b>%s — line %d</b></p><p class="space"><tt>%s</tt></p><p>%s</p></section>`

	return fmt.Sprintf(t_error_html, e.kind, e.pos.line, e.pos.file_path, e.message)
}

func (e *Spindle_Pos_Error) term_string() string {
	const t_error_term = "%s! %s — line %d\n    %s"

	return fmt.Sprintf(t_error_term, e.kind, e.pos.file_path, e.pos.line, e.message)
}



type Spindle_Error struct {
	kind    Error_Type
	message string
}

func (e *Spindle_Error) html_string() string {
	const t_error_html = `<section><p><b>%s!</b></p><p>%s</p></section>`

	return fmt.Sprintf(t_error_html, e.kind, e.message)
}

func (e *Spindle_Error) term_string() string {
	const t_error_term = "%s!\n    %s"

	return fmt.Sprintf(t_error_term, e.kind, e.message)
}


func new_error_handler() *Error_Handler {
	e := Error_Handler{}
	e.reset()
	return &e
}

type Error_Handler struct {
	has_failures bool
	all_errors   []_error
}

func (e *Error_Handler) reset() {
	e.has_failures = false
	e.all_errors   = make([]_error, 0, 8)
}

func (e *Error_Handler) new_pos(kind Error_Type, pos position, message string, subst ...any) {
	if kind > is_failure {
		e.has_failures = true
	}

	e.all_errors = append(e.all_errors, &Spindle_Pos_Error{
		kind,
		pos,
		fmt.Sprintf(message, subst...),
	})
}

func (e *Error_Handler) new(kind Error_Type, message string, subst ...any) {
	if kind > is_failure {
		e.has_failures = true
	}

	e.all_errors = append(e.all_errors, &Spindle_Error {
		kind,
		fmt.Sprintf(message, subst...),
	})
}

func (e *Error_Handler) has_errors() bool {
	return len(e.all_errors) > 0
}

/*
	@todo right now we just render all error
	types together — warnings will become a
	modal, while failures will be served as
	an error page
*/
func (e *Error_Handler) render_html_page() string {
	buffer := strings.Builder{}
	buffer.Grow(len(e.all_errors) * 128)

	for _, the_error := range e.all_errors {
		buffer.WriteString(the_error.html_string())
	}

	return fmt.Sprintf(ERROR_PAGE, buffer.String())
}

/*func (e *Error_Handler) render_html_modal() string {
	buffer := strings.Builder{}
	buffer.Grow(len(e.all_errors) * 128)

	for _, the_error := range e.all_errors {
		buffer.WriteString(the_error.html_string())
	}

	return fmt.Sprintf(ERROR_MODAL, buffer.String())
}*/

func (e *Error_Handler) render_term_errors() string {
	// @todo sort these by severity

	buffer := strings.Builder{}
	buffer.Grow(len(e.all_errors) * 128)

	for _, the_error := range e.all_errors {
		buffer.WriteString(the_error.term_string())
		buffer.WriteString("\n\n")
	}

	return strings.TrimSpace(buffer.String())
}

const ERROR_PAGE_NOT_FOUND = `<html>` + ERROR_HEAD + `<body>
<h1>` + SPINDLE + `</h1>
<main>
	<section><p><b>Page not found...</b></p></section>
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

const ERROR_HEAD = `<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<title>Spindle</title>
	<link rel="stylesheet" type="text/css" href="/_spindle/manual/style.css"/>` + RELOAD_SCRIPT + `</head>`

const ERROR_PAGE = `<!DOCTYPE html>
<html>` + ERROR_HEAD + `<body>
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