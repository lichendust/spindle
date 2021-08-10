package main

import (
	"os"
	"io"
	"image"
	"image/png"
	"image/jpeg"
	"path/filepath"
	"github.com/nfnt/resize"
)

func generic_resize(img image.Image, w, h int) image.Image {
	return resize.Resize(uint(w), uint(h), img, resize.MitchellNetravali)
}

func image_handler(the_file *file, w, h int) {
	source_file, err := os.Open(the_file.source)

	if err != nil {
		panic(err)
	}

	defer source_file.Close()

	var con image.Config

	switch the_file.file_type {
		case IMAGE_JPG:
			con, err = jpeg.DecodeConfig(source_file)

		case IMAGE_PNG:
			con, err = png.DecodeConfig(source_file)
	}

	if err != nil {
		panic(err)
	}

	ratio := con.Width / con.Height

	target_width  := 0
	target_height := 0

	if ratio >= 1 {
		if con.Width > w {
			target_width = w
		} else {
			target_width = con.Width
		}
	} else {
		if con.Height > h {
			target_height = h
		} else {
			target_height = con.Height
		}
	}

	source_file.Seek(0, io.SeekStart)

	ext := filepath.Ext(the_file.source)
	out_path := the_file.output[:len(the_file.output) - len(ext)]

	var img image.Image

	switch the_file.file_type {
		case IMAGE_JPG:
			img, err = jpeg.Decode(source_file)

		case IMAGE_PNG:
			img, err = png.Decode(source_file)
	}

	if err != nil {
		panic(err)
	}

	switch the_file.file_type {
	case IMAGE_JPG:
		write_jpeg(generic_resize(img, target_width, target_height), the_file.output)
		write_jpeg(generic_resize(img, target_width / 2, target_height / 2), out_path + "@medium" + ext)
		write_jpeg(generic_resize(img, target_width / 4, target_height / 4), out_path + "@small"  + ext)

	case IMAGE_PNG:
		write_png(generic_resize(img, target_width, target_height), the_file.output)
		write_png(generic_resize(img, target_width / 2, target_height / 2), out_path + "@medium" + ext)
		write_png(generic_resize(img, target_width / 4, target_height / 4), out_path + "@small"  + ext)
	}
}

func write_jpeg(img image.Image, out_path string) {
	out, err := os.Create(out_path)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	jpeg.Encode(out, img, &jpeg.Options{90})
}

func write_png(img image.Image, out_path string) {
	out, err := os.Create(out_path)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	png.Encode(out, img)
}