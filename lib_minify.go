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

func copy_mini(src, dst, mode string) {
	source, err := os.Open(src)

	if err != nil {
		panic(err) // @error
	}

	defer source.Close()

	output, err := os.Create(dst)

	if err != nil {
		panic(err) // @error
	}

	defer output.Close()

	writer := bufio.NewWriter(output)

	if err := minifier.Minify(mode, writer, source); err != nil {
		panic(err) // @error
	}

	writer.Flush()
}