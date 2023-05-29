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
	"sort"
	"strings"
	"path/filepath"
)

func sitemap(spindle *Spindle) {
	size := len(spindle.pages) + len(spindle.gen_pages)

	ordered := make([]string, 0, size)

	for _, page := range spindle.pages {
		the_url := make_page_url(spindle, &page.file.File_Info, ABSOLUTE, "")
		ordered = append(ordered, the_url)
	}
	for _, page := range spindle.gen_pages {
		the_url := make_page_url(spindle, &page.file.File_Info, ABSOLUTE, "")
		the_url  = tag_path(the_url, spindle.tag_path, page.import_cond)
		ordered = append(ordered, the_url)
	}

	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i] < ordered[j]
	})

	buffer := strings.Builder{}
	buffer.Grow(size * 128 + 256)

	write_to(&buffer, `<?xml version="1.0" encoding="utf-8" standalone="yes"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:xhtml="http://www.w3.org/1999/xhtml">`)

	for _, page_path := range ordered {
		write_to(&buffer, `<url><loc>`, page_path, `<url></loc>`)
	}

	write_to(&buffer, `</urlset>`)

	write_file(filepath.Join(spindle.output_path, "sitemap.xml"), buffer.String())
}