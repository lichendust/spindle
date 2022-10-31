package main

import (
	"os"
	"fmt"
	"image"
	"path/filepath"

	"github.com/nfnt/resize"

	lib_png "image/png"
	lib_jpg "image/jpeg"
	lib_tif "golang.org/x/image/tiff"

	lib_web "github.com/kolesa-team/go-webp/webp"
	web_encoder "github.com/kolesa-team/go-webp/encoder"
	web_decoder "github.com/kolesa-team/go-webp/decoder"
)

const default_quality = 100

type image_settings struct {
	width   uint
	height  uint
	quality int
	format  file_type
}

func resize_image(spindle *spindle, incoming_file *disk_object, settings *image_settings) bool {
	var source_image image.Image

	{
		source_file, err := os.Open(incoming_file.path)
		if err != nil {
			return false
		}

		var img_err error

		switch incoming_file.file_type {
		case IMG_JPG:
			source_image, img_err = lib_jpg.Decode(source_file)
		case IMG_TIF:
			source_image, img_err = lib_tif.Decode(source_file)
		case IMG_PNG:
			source_image, img_err = lib_png.Decode(source_file)
		case IMG_WEB:
			source_image, img_err = lib_web.Decode(source_file, &web_decoder.Options{})
		}

		if img_err != nil {
			return false
		}
	}

	// thumbnail just forces aspect ratio instead of allowing free-form sizing — it just has a misleading name

	output_image := source_image

	if settings.width > 0 && settings.height > 0 {
		output_image = resize.Thumbnail(settings.width, settings.height, source_image, resize.MitchellNetravali)
	}

	output_path := rewrite_root(rewrite_image_path(incoming_file.path, settings), public_path)
	make_dir(filepath.Dir(output_path))

	switch settings.format {
	case IMG_JPG:
		output_file, err := os.Create(output_path)
		if err != nil {
			panic(err)
		}

		err = lib_jpg.Encode(output_file, output_image, &lib_jpg.Options { settings.quality })
		if err != nil {
			panic(err)
		}

	case IMG_PNG:
		output_file, err := os.Create(output_path)
		if err != nil {
			panic(err)
		}

		encoder := lib_png.Encoder { CompressionLevel: -1 }

		err = encoder.Encode(output_file, output_image)
		if err != nil {
			panic(err)
		}

	case IMG_WEB:
		output_file, err := os.Create(output_path)
		if err != nil {
			panic(err)
		}

		options, err := web_encoder.NewLossyEncoderOptions(web_encoder.PresetDefault, float32(settings.quality))
		if err != nil {
			panic(err)
		}

		err = lib_web.Encode(output_file, output_image, options)
		if err != nil {
			panic(err)
		}
	}

	return true
}

func rewrite_image_path(path string, settings *image_settings) string {
	ext := ""
	switch settings.format {
	case IMG_JPG: ext = ".jpg"
	case IMG_PNG: ext = ".png"
	case IMG_TIF: ext = ".tif"
	case IMG_WEB: ext = ".webp"
	default:      ext = filepath.Ext(path)
	}

	if settings.width > 0 && settings.height > 0 {
		return rewrite_ext(path, fmt.Sprintf("_%dx%d%s", settings.width, settings.height, ext))
	}

	return rewrite_ext(path, ext)
}