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
	"os"
	"bufio"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/minify/v2/css"
)

func copy_minify(the_file *File, output_path string) bool {
	minifier := minify.New()
	minifier.AddFunc("text/js", js.Minify)
	minifier.AddFunc("text/css", css.Minify)

	source, err := os.Open(the_file.path)
	if err != nil {
		return false
	}

	defer source.Close()

	mode := ""

	switch the_file.file_type {
	case JAVASCRIPT: mode = "text/js"
	case CSS:        mode = "text/css"
	}

	output, err := os.Create(output_path)

	if err != nil {
		return false
	}

	defer output.Close()

	writer := bufio.NewWriter(output)

	err = minifier.Minify(mode, writer, source)
	if err != nil {
		return false
	}

	writer.Flush()
	return true
}