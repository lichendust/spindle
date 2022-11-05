package main

import (
	"os"
	"bufio"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/minify/v2/css"
)

func copy_minify(the_file *disk_object, output_path string) bool {
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