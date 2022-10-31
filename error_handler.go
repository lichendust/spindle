package main

import (
	"fmt"
	"strings"
	"path/filepath"
)

type _error interface {
	html_string() string
	term_string() string
}

type error_type uint8
const (
	WARNING error_type = iota
	RENDER_WARNING
	PARSER_WARNING

	is_failure
	FAILURE
	RENDER_FAILURE
	PARSER_FAILURE
)

func (e error_type) String() string {
	switch e {
	case WARNING:
		return "warning"
	case RENDER_WARNING:
		return "render warning"
	case PARSER_WARNING:
		return "parser warning"
	case FAILURE:
		return "failure"
	case RENDER_FAILURE:
		return "render failure"
	case PARSER_FAILURE:
		return "parser failure"
	}
	return "unknown"
}

type spindle_pos_error struct {
	kind    error_type
	pos     position
	message string
}

func (e *spindle_pos_error) html_string() string {
	const t_error_html = `<section><p><b>%s — line %d</b></p><p class="space"><tt>%s</tt></p><p>%s</p></section>`

	return fmt.Sprintf(t_error_html, e.kind, e.pos.line, e.pos.file_path, e.message)
}

func (e *spindle_pos_error) term_string() string {
	const t_error_term = "%s! %s — line %d\n    %s"

	return fmt.Sprintf(t_error_term, e.kind, e.pos.file_path, e.pos.line, e.message)
}



type spindle_error struct {
	kind    error_type
	message string
}

func (e *spindle_error) html_string() string {
	const t_error_html = `<section><p><b>%s!</b></p><p>%s</p></section>`

	return fmt.Sprintf(t_error_html, e.kind, e.message)
}

func (e *spindle_error) term_string() string {
	const t_error_term = "%s!\n    %s"

	return fmt.Sprintf(t_error_term, e.kind, e.message)
}


func new_error_handler() *error_handler {
	e := error_handler{}
	e.reset()
	return &e
}

type error_handler struct {
	has_failures bool
	all_errors   []_error
}

func (e *error_handler) reset() {
	e.has_failures = false
	e.all_errors   = make([]_error, 0, 8)
}

func (e *error_handler) new_pos(kind error_type, pos position, message string, subst ...any) {
	if kind > is_failure {
		e.has_failures = true
	}

	file := ""

	if !strings.Contains(file, "index") {
		file = filepath.Base(pos.file_path)
	} else {
		file = pos.file_path
	}

	e.all_errors = append(e.all_errors, &spindle_pos_error {
		kind,
		pos,
		fmt.Sprintf(message, subst...),
	})
}

func (e *error_handler) new(kind error_type, message string, subst ...any) {
	if kind > is_failure {
		e.has_failures = true
	}

	e.all_errors = append(e.all_errors, &spindle_error {
		kind,
		fmt.Sprintf(message, subst...),
	})
}

func (e *error_handler) has_errors() bool {
	return len(e.all_errors) > 0
}

/*
	@todo right now we just render all error
	types together — warnings will become a
	modal, while failures will be served as
	an error page
*/
func (e *error_handler) render_html_errors() string {
	buffer := strings.Builder{}
	buffer.Grow(len(e.all_errors) * 128)

	for _, the_error := range e.all_errors {
		buffer.WriteString(the_error.html_string())
	}

	return fmt.Sprintf(t_error_page, title, buffer.String())
}

func (e *error_handler) render_term_errors() string {
	// @todo sort these by severity

	buffer := strings.Builder{}
	buffer.Grow(len(e.all_errors) * 128)

	for _, the_error := range e.all_errors {
		buffer.WriteString(the_error.term_string())
		buffer.WriteString("\n\n")
	}

	return strings.TrimSpace(buffer.String())
}

const t_error_page = `<!DOCTYPE html>
<html>
	<head>
		<meta charset="utf-8">
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<title>Spindle</title>
		<style type="text/css">
			body {
				font-family: Atkinson Hyperlegible, Helvetica, Arial, sans-serif;
				margin: 5ex;
				font-size: 1.2rem;
			}
			tt {
				font-family: DM Mono, SF Mono, Roboto Mono, Source Code Pro, Fira Code, monospace;
			}
			.space {
				margin-bottom: 1ex;
			}
			code {
				background: #eee;
				padding: .2ex .5ex;
			}
			ul { padding-left: 2ex }
			p  { padding: 0; margin: 0 }
			a  { color: black }
			a:hover {
				color: white;
				background: black;
			}
			main {
				float: left;
				width: 60ex;
				margin-right: 2vw;
				margin-bottom: 4vh;
			}
			aside {
				float: left;
				max-width: 24ex;
			}
			section:not(:first-child) {
				margin-top: 1rem;
			}
		</style>
	</head>
	<body>
		<h1>%s</h1>
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
	</body>
</html>`