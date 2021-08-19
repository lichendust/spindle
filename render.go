package main

import (
	"fmt"
	"sort"
	"strings"
	"path/filepath"
)

func markup_render(markup *markup) string {
	buffer := strings.Builder {}
	buffer.Grow(100)

	buffer.WriteString(`<!DOCTYPE html>`)

	head_render(markup, &buffer)
	body_render(markup, &buffer)

	buffer.WriteString(`</html>`)

	return buffer.String()
}

func head_render(markup *markup, buffer *strings.Builder) {
	buffer.Grow(2048)

	buffer.WriteString(`<head>`)
	buffer.WriteString(`<meta charset='utf-8'>`)

	if x, ok := markup.vars["favicon"]; ok {
		buffer.WriteString(make_favicon(x))
	}

	{
		title := markup.vars["title"]

		if x, ok := markup.vars["title_prefix"]; ok {
			title = sprint("<title>%s %s</title>", x, title)
		} else {
			title = sprint(`<title>%s</title>`, title)
		}

		buffer.WriteString(title)
	}

	domain := markup.vars["domain"]

	{
		canon_path := join_url(domain, markup.vars["url_pretty"])

		buffer.WriteString(sprint(meta_canonical, canon_path))
		buffer.WriteString(sprint(meta_source, "og:url", canon_path))
	}

	clean_list := make([]string, 0, len(markup.vars))

	for key, _ := range markup.vars {
		if strings.HasPrefix(key, "meta.") {
			clean_list = append(clean_list, key)
		}
	}

	sort.Strings(clean_list)

	for _, key := range clean_list {
		value := markup.vars[key]

		key = key[5:]

		switch key {
		case "viewport":
			buffer.WriteString(sprint(meta_viewport, value))
			continue

		case "image":
			value = join_url(domain, value)

		case "description":
			buffer.WriteString(sprint(meta_description, value))

		case "twitter_creator":
			buffer.WriteString(sprint(meta_source, "twitter:creator", value))
			continue

		case "twitter_site":
			buffer.WriteString(sprint(meta_source, "twitter:site", value))
			continue

		case "twitter_card_type":
			if ok := valid_twitter_card[value]; ok {
				buffer.WriteString(sprint(meta_source, "twitter:card", value))
			} else {
				console_handler.print("not a twitter card type: %q", value) // @error
			}
			continue
		}

		buffer.WriteString(sprint(meta_source, "og:" + key, value))
	}

	if x, ok := markup.vars["style"]; ok {
		for _, n := range strings.Fields(x) {
			buffer.WriteString(sprint(style_template, n))
		}
	}

	if x, ok := markup.vars["script"]; ok {
		for _, n := range strings.Fields(x) {
			buffer.WriteString(sprint(script_template, n))
		}
	}

	buffer.WriteString(`</head>`)
}

func body_render(markup *markup, buffer *strings.Builder) {
	markup.pos = -1

	buffer.Grow(256 * len(markup.data))

	buffer.WriteString(`<body>`)

	above_content := execute_looped_chunks(markup.vars, "above_content")
	below_content := execute_looped_chunks(markup.vars, "below_content")

	above_wrap := execute_looped_chunks(markup.vars, "above_wrap")
	below_wrap := execute_looped_chunks(markup.vars, "below_wrap")

	text := data_render(markup, markup.vars)

	text = above_content + text + below_content

	if x, ok := markup.vars["wrap_content"]; ok {
		text = sprint(complex_key_mapper(x, markup.vars), text)
	}

	buffer.WriteString(above_wrap)
	buffer.WriteString(text)
	buffer.WriteString(below_wrap)

	buffer.WriteString(`</body>`)
}

func data_render(markup *markup, vars map[string]string) string {
	buffer := strings.Builder {}
	buffer.Grow(256 * len(markup.data) - markup.pos)

	for {
		markup.pos++

		if markup.pos >= len(markup.data) {
			return buffer.String()
		}

		obj := markup.data[markup.pos]

		switch obj.object_type {
		case WHITESPACE:
			continue

		case BLOCK_END:
			return buffer.String()

		case BLOCK:
			name := obj.text[0]

			new_vars := process_vars(markup, merge_maps(obj.vars, vars))
			new_text := data_render(markup, new_vars)
			temp, ok := vars[name]

			if ok {
				new_text = sprint(temp, new_text)
			} else if ok := html_defaults[name]; ok {
				new_text = sprint("<%s>%s</%s>", name, new_text, name)
			} else {
				new_text = sprint("<div class='%s'>%s</div>", name, new_text)
			}

			buffer.WriteString(complex_key_mapper(new_text, new_vars))

		case BLOCK_IF:
			text, eval := vars[obj.text[0]]

			eval = eval && text != "0"

			// is inverse
			if (obj.offset == 1) {
				eval = !eval
			}

			if eval {
				new_vars := process_vars(markup, merge_maps(obj.vars, vars))
				new_text := data_render(markup, new_vars)
				buffer.WriteString(complex_key_mapper(new_text, new_vars))
			} else {
				skip_block(markup)
			}

			i := markup.pos + 1

			if i >= len(markup.data) {
				continue
			}

			obj := markup.data[i]

			if obj.object_type == BLOCK_ELSE {
				markup.pos++ // we can safely advance to this "else" token

				if eval {
					skip_block(markup)
				} else {
					new_vars := merge_maps(obj.vars, vars)
					new_text := data_render(markup, new_vars)
					buffer.WriteString(complex_key_mapper(new_text, new_vars))
				}
			}

		case BLOCK_ELSE:
			console_handler.print("orphaned else block") // @error
			skip_block(markup)

		case BLOCK_CODE:
			text := obj.text[0]

			// @todo highlighting engine
			/*if len(obj.text) > 1 {
				text = highlight(text, obj.text[1])
			}*/

			buffer.WriteString(sprint(vars["codeblock"], text))

		case LIST_O, LIST_U:
			the_type := obj.object_type

			local_buffer := strings.Builder {}

			for _, obj := range markup.data[markup.pos:] {
				if obj.object_type != the_type {
					break
				}

				markup.pos++
				local_buffer.WriteString(sprint(vars["li"], complex_key_mapper(obj.text[0], vars)))
			}

			// because the main loop was
			// already indexing the first
			// list item, we've technically
			// counted it twice in this
			// internal loop so we just
			// subtract one to compensate
			markup.pos -= 1

			final_text := local_buffer.String()

			switch the_type {
			case LIST_O: final_text = sprint(vars["ol"], final_text)
			case LIST_U: final_text = sprint(vars["ul"], final_text)
			}

			buffer.WriteString(final_text)

		case FUNCTION:
			fname := obj.text[0]

			raw, ok := load_file_cache(sprint("config/chunks/%s.js", fname))

			if !ok {
				console_handler.print("function not found:", fname) // @error
				continue
			}

			text, ok := call_script(vars, raw, obj.text[1:])

			if !ok {
				console_handler.print(fname, text) // @error
				continue
			}

			buffer.WriteString(complex_key_mapper(text, vars))

		case FUNCTION_INLINE:
			text, ok := call_script(vars, obj.text[0], []string{})

			if !ok {
				console_handler.print(text) // @error
				continue
			}

			buffer.WriteString(complex_key_mapper(text, vars))

		case CHUNK:
			buffer.WriteString(execute_chunk(vars, obj.text[0]))

		case MEDIA:
			buffer.WriteString(do_media(obj.text))

		case IMPORT:
			path := filepath.ToSlash(sprint("source/%s.x", obj.text[0]))

			if markup.no_drafts && is_draft(path) {
				console_handler.print("import: %q is draft; skipped", obj.text[0]) // @warning
				continue
			}

			page, ok := load_page(path, markup.no_drafts)

			if !ok {
				console_handler.print("import: path %q does not exist", obj.text[0]) // @error
				continue
			}

			key := "import"

			if len(obj.text) > 1 {
				key = obj.text[1]
			}

			x, ok := markup.vars[key]

			if !ok {
				console_handler.print("import: no template %q", obj.text[0]) // @error
				continue
			}

			buffer.WriteString(complex_key_mapper(x, page.vars))

		case IMAGE:
			temp := vars[get_id(obj, "img")]

			image_path := safe_join_image_prefix(markup, obj.text[0])

			if !markup.no_drafts {
				image_path = strip_image_size(image_path)
			}

			obj.text[0] = image_path
			buffer.WriteString(sprint(temp, obj.text...))

		case DIVIDER:
			buffer.WriteString(vars["hr"])

		case RAW_TEXT:
			buffer.WriteString(complex_key_mapper(obj.text[0], vars))

		case HEADING:
			temp := vars[fmt.Sprintf("h%d", obj.offset)] // this is an exception to get_id
			id   := make_element_id(obj.text[0])
			text := complex_key_mapper(obj.text[0], vars)
			buffer.WriteString(sprint(temp, id, text))

		case PARAGRAPH:
			buffer.WriteString(sprint(vars["paragraph"], complex_key_mapper(obj.text[0], vars)))
		}
	}

	return buffer.String()
}

func skip_block(markup *markup) {
	depth := 1

	for {
		markup.pos++

		if markup.pos >= len(markup.data) {
			return
		}

		obj := markup.data[markup.pos]

		if obj.object_type >= BLOCK {
			depth++
 		} else if obj.object_type == BLOCK_END {
			depth--
		}

		if depth == 0 {
			return
		}
	}
}

func get_id(m *markup_object, name string) string {
	if m.offset <= 1 {
		return name
	}
	return fmt.Sprintf("%s%d", name, m.offset)
}

// a takes precedence over b
func merge_maps(a, b map[string]string) map[string]string {
	new := make(map[string]string, len(a) + len(b))

	for k, v := range b {
		new[k] = v
	}

	for k, v := range a {
		new[k] = v
	}

	return new
}

// specialist version that takes into account
// script and style stacking only needed in
// plates and config loading
func merge_vars(a, b map[string]string) map[string]string {
	new := merge_maps(a, b)

	if x, ok := a["script"]; ok {
		new["script"] = strings.Replace(x, "default", b["script"], 1)
	} else {
		new["script"] = b["script"]
	}

	if x, ok := a["style"]; ok {
		new["style"] = strings.Replace(x, "default", b["style"], 1)
	} else {
		new["style"] = b["style"]
	}

	return new
}

func execute_chunk(vars map[string]string, name string) string {
	path := filepath.Join("config/chunks", name)
	ext  := filepath.Ext(path)

	is_markup := (ext == ".x")

	if ext == "" {
		path += ".x"
		is_markup = true
	}

	raw, ok := load_file_cache(path)

	if !ok {
		fmt.Println("chunk", name, "not found")
		return ""
	}

	if !is_markup {
		return complex_key_mapper(raw, vars)
	}

	chunk_markup := markup_parser(raw)

	chunk_markup.vars = merge_maps(chunk_markup.vars, vars)
	chunk_markup.pos  = -1

	assign_plate(chunk_markup)
	chunk_markup.vars = process_vars(chunk_markup, chunk_markup.vars)

	return data_render(chunk_markup, chunk_markup.vars)
}

func execute_looped_chunks(vars map[string]string, v string) string {
	if x, ok := vars[v]; ok {
		chunks := strings.Fields(x)
		buffer := strings.Builder {}
		buffer.Grow(len(chunks) * 1024)

		for _, c := range chunks {
			buffer.WriteString(execute_chunk(vars, c))
		}

		return buffer.String()
	}

	return ""
}

func make_favicon(f string) string {
	ext := filepath.Ext(f)

	switch ext {
		case ".ico":
			return sprint(meta_favicon, "x-icon", f)
		case ".png":
			return sprint(meta_favicon, "png", f)
		case ".gif":
			return sprint(meta_favicon, "gif", f)
		default:
			fmt.Println("favicon: unknown file extension", ext)
	}

	return ""
}