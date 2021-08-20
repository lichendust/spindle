package main

import (
	"os"
	"os/exec"
	"fmt"
	"image"
	"image/png"
	"image/jpeg"
	"path/filepath"
)

func image_handler(the_file *file, long_axis_max int) {
	source_file, err := os.Open(the_file.source)

	if err != nil {
		panic(err)
	}

	var con image.Config

	switch the_file.file_type {
		case IMAGE_JPG:
			con, err = jpeg.DecodeConfig(source_file)

		case IMAGE_PNG:
			con, err = png.DecodeConfig(source_file)
	}

	source_file.Close()

	if err != nil {
		panic(err)
	}

	if con.Width < long_axis_max && con.Height < long_axis_max {
		long_axis_max = 0
	}

	ext := filepath.Ext(the_file.source)
	out_path := the_file.output[:len(the_file.output) - len(ext)]

	magick_copy(the_file.source, the_file.output,            long_axis_max)
	magick_copy(the_file.source, out_path + "@medium" + ext, long_axis_max / 2)
	magick_copy(the_file.source, out_path + "@small"  + ext, long_axis_max / 4)
}

func magick_copy(source, output string, long_axis_max int) {
	// convert -resize "1920x1920>" -strip -interlace Plane -quality 85% $1 $1

	command := make([]string, 0, 9)

	if long_axis_max > 0 {
		command = append(command, "-resize", fmt.Sprintf("%dX%d>", long_axis_max, long_axis_max))
	}

	command = append(command, "-strip", "-interlace", "Plane", "-quality", "85%", source, output)

	cmd := exec.Command("convert", command...)

	result, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Println("imagemagick:", output, string(result))
	}
}