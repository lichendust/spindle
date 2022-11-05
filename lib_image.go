package main

import (
	"os"
	"fmt"
	"image"

	"github.com/nfnt/resize"

	lib_png "image/png"
	lib_jpg "image/jpeg"
	lib_tif "golang.org/x/image/tiff"

	lib_web "github.com/kolesa-team/go-webp/webp"
	web_encoder "github.com/kolesa-team/go-webp/encoder"
	web_decoder "github.com/kolesa-team/go-webp/decoder"
)

type image_settings struct {
	width     uint
	height    uint
	quality   int
	file_type file_type
}

func (s *image_settings) make_hash() uint32 {
	return new_hash(fmt.Sprintf("%d%d%d%d", s.width, s.height, s.quality, s.file_type))
}

func copy_generated_image(the_image *generated_image, output_path string) bool {
	var source_image image.Image

	incoming_file := the_image.original
	settings      := the_image.settings

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

	output_image := source_image

	if settings.width > 0 || settings.height > 0 {
		// thumbnail just forces aspect ratio instead of allowing free-form sizing — it just has a misleading name
		output_image = resize.Thumbnail(settings.width, settings.height, source_image, resize.MitchellNetravali)
	}

	switch settings.file_type {
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