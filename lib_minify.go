package main

import (
	"os"
	"bufio"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/minify/v2/css"
)

var minifier *minify.M

func init_minify() {
	minifier = minify.New()
	minifier.AddFunc("text/css", css.Minify)
	minifier.AddFunc("text/js", js.Minify)
}

func copy_mini(the_file *file) {
	source, err := os.Open(the_file.source)

	if err != nil {
		panic(err) // @error
	}

	defer source.Close()

	output, err := os.Create(the_file.output)

	if err != nil {
		panic(err) // @error
	}

	mode := ""

	switch the_file.file_type {
	case STATIC_JS:  mode = "text/js"
	case STATIC_CSS: mode = "text/css"
	}

	defer output.Close()

	writer := bufio.NewWriter(output)

	if err := minifier.Minify(mode, writer, source); err != nil {
		panic(err) // @error
	}

	writer.Flush()
}