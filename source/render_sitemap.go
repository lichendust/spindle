package main

import (
	"sort"
	"strings"
	"path/filepath"
)

func sitemap(spindle *spindle) {
	size := len(spindle.pages) + len(spindle.gen_pages)

	ordered := make([]string, 0, size)

	for _, page := range spindle.pages {
		ordered = append(ordered, make_page_url(spindle, &page.file.file_info, ABSOLUTE, ""))
	}
	for _, page := range spindle.gen_pages {
		ordered = append(ordered, tag_path(make_page_url(spindle, &page.file.file_info, ABSOLUTE, ""), spindle.tag_path, page.import_cond))
	}

	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i] < ordered[j]
	})

	buffer := strings.Builder{}
	buffer.Grow(size * 128 + 256)

	buffer.WriteString(`<?xml version="1.0" encoding="utf-8" standalone="yes"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:xhtml="http://www.w3.org/1999/xhtml">`)

	for _, page_path := range ordered {
		buffer.WriteString(`<url><loc>`)
		buffer.WriteString(page_path)
		buffer.WriteString(`</url></loc>`)
	}

	buffer.WriteString(`</urlset>`)

	write_file(filepath.Join(spindle.output_path, "sitemap.xml"), buffer.String())
}