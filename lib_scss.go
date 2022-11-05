package main

import (
	"io"
	"os"
	"bufio"
	"bytes"

	"github.com/wellington/go-libsass"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
)

func serve_scss(the_file *disk_object) ([]byte, bool) {
	source, err := os.Open(the_file.path)
	if err != nil {
		return nil, false
	}

	defer source.Close()

	buffer := bytes.Buffer{}
	buffer.Grow(512)

	comp, err := libsass.New(&buffer, source)
	if err != nil {
		return nil, false
	}

	err = comp.Run()
	if err != nil {
		return nil, false
	}

	return buffer.Bytes(), true
}

func copy_scss(the_file *disk_object, output_path string) bool {
	source, err := os.Open(the_file.path)
	if err != nil {
		return false
	}

	defer source.Close()

	output, err := os.Create(output_path)
	if err != nil {
		return false
	}

	defer output.Close()

	r, w := io.Pipe()

	{
		comp, err := libsass.New(w, source)
		if err != nil {
			return false
		}

		go func() {
			err := comp.Run()
			if err != nil {
				panic(err)
			}
			w.Close()
		}()
	}

	{
		minifier := minify.New()
		minifier.AddFunc("text/css", css.Minify)

		file_writer := bufio.NewWriter(output)

		err := minifier.Minify("text/css", file_writer, r)
		if err != nil {
			return false
		}

		file_writer.Flush()
	}

	return true
}