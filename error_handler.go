package main

import "fmt"
import "bytes"

const (
	WARNING uint8 = iota
	FAILURE
)

type spindle_error struct {
	kind    uint8
	line    uint
	file    string
	message string
}

func (e *spindle_error) html_string() string {
	return fmt.Sprintf(t_error_html, e.line, e.message)
}

type error_handler struct {
	all_errors []spindle_error
}

func (e *error_handler) new(kind uint8, file string, line uint, message string, subst ...any) {
	e.all_errors = append(e.all_errors, spindle_error {
		kind,
		line,
		file,
		fmt.Sprintf(message, subst...),
	})
}

func (e *error_handler) has_errors() bool {
	return len(e.all_errors) > 0
}

func (e *error_handler) render_failures() string {
	buffer := bytes.Buffer{}
	buffer.Grow(len(t_error_html) * len(e.all_errors) * 2)

	for _, the_error := range e.all_errors {
		buffer.WriteString(the_error.html_string()) // @todo right now we don't separate warnings/failures
	}

	const page_title = `some title`

	return fmt.Sprintf(t_error_page, title, page_title, buffer.String())
}

/*func (e *error_handler) render_warnings() string {
	buffer := bytes.Buffer{}
	buffer.Grow(len(t_error_html) * len(e.all_errors) * 2)

	for _, the_error := range e.all_errors {
		if the_error.kind == WARNING {
			buffer.WriteString(the_error.html_string())
		}
	}

	const page_title = `some title`

	return fmt.Sprintf(t_error_overlay, title, page_title, buffer.String())
}*/

const t_error_html = `<section>
	<p><b>Line %d</b></p>
	<p>%s</p>
</section>`

const t_error_overlay = `` // this will be inserted into the page as a modal

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
			<p><b>%s</b></p>
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