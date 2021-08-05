package main

import (
	"sort"
	"strings"
	"path/filepath"
)

func build_project(args []string) {
	files, folders := get_files("source", true)

	public_dir := "public"

	if len(args) > 0 {
		public_dir = args[0]
	}

	make_directory(public_dir)

	init_minify() // for copy_mini

	for _, folder := range folders {
		make_directory(filepath.Join(public_dir, folder.path[7:]))
	}

	for _, file := range files {
		out_path := filepath.Join(public_dir, file.path[7:])

		switch filepath.Ext(file.path) {
		case ".x":
			page_obj, _ := load_page(file.path, true)

			out_text := markup_render(page_obj)
			out_path = sprint("%s.html", out_path[:len(out_path) - len(filepath.Ext(out_path))])

			make_file(out_path, out_text)

		case ".js":
			copy_mini(file.path, out_path, "text/js")

		case ".css":
			copy_mini(file.path, out_path, "text/css")

		default:
			copy_file(file.path, out_path)
		}
	}

	sitemap(files, public_dir)

	console_handler.flush()
}

func load_page(path string, no_drafts bool) (*markup, bool) {
	raw_text, ok := load_file(path)

	if !ok {
		return nil, false
	}

	page_obj := markup_parser(raw_text)

	page_obj.no_drafts = no_drafts

	assign_plate(page_obj)
	process_vars(page_obj)

	page_obj.vars["page_path"] = make_url_from_path(path[6:])

	return page_obj, true
}

func make_url_from_path(input string) string {
	input = input[:len(input) - len(filepath.Ext(input))]

	if strings.HasSuffix(input, "/index") {
		input = input[:len(input) - 6]
	}

	return input
}

func sitemap(files []*file, public_dir string) {
	ordered := make([]string, len(files) + 24)

	for _, file := range files {
		ext := filepath.Ext(file.path)

		switch ext {
		case ".x", ".html":
			path := make_url_from_path(file.path[6:])
			ordered = append(ordered, sprint(sitemap_entry, join_url(config.vars["domain"], path)))
		default:
			continue
		}
	}

	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i] < ordered[j]
	})

	buffer := strings.Builder {}
	buffer.Grow(len(sitemap_template) + len(ordered) * len(sitemap_entry))

	buffer.WriteString(sitemap_template)

	for _, page := range ordered {
		buffer.WriteString(page)
	}

	buffer.WriteString(`</urlset>`)

	make_file(filepath.Join(public_dir, "sitemap.xml"), buffer.String())
}