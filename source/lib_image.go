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

import "image"
import "strconv"

import "os"
import "os/exec"

import "github.com/nfnt/resize"

import lib_png "image/png"
import lib_jpg "image/jpeg"
import lib_tif "golang.org/x/image/tiff"
import lib_web "golang.org/x/image/webp"

// @todo remove quality from individual
//  images and just have one global setting?
//  it just never seems to come up in
//  practice

const DEFAULT_QUALITY = 100

type Image_Settings struct {
	max_size  uint
	width     uint
	height    uint
	quality   int
	file_type File_Type
}

/*func (s *Image_Settings) make_hash() uint32 {
	return new_hash(fmt.Sprintf("%d%d%d%d", s.width, s.height, s.quality, s.File_Type))
}*/

func copy_generated_image(the_image *Image, output_path string) bool {
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
			source_image, img_err = lib_web.Decode(source_file)
		}

		if img_err != nil {
			return false
		}
	}

	// break flow to escalate straight to cwebp
	// instead of wasting any time below
	if settings.file_type == IMG_WEB {
		b := source_image.Bounds()
		x := uint(b.Dx())
		y := uint(b.Dy())

		// correct aspect ratio
		settings.width, settings.height = scaling(settings.max_size, settings.max_size, x, y)

		ext_cwebp(incoming_file.path, output_path, settings)

		return true
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
			eprintf("failed to create image %q\n", output_path)
		}

		err = lib_jpg.Encode(output_file, output_image, &lib_jpg.Options { settings.quality })
		if err != nil {
			eprintf("failed to encode image %q\n", output_path)
		}

	case IMG_PNG:
		output_file, err := os.Create(output_path)
		if err != nil {
			eprintf("failed to create image %q\n", output_path)
		}

		encoder := lib_png.Encoder {
			CompressionLevel: lib_png.NoCompression,
		}

		err = encoder.Encode(output_file, output_image)
		if err != nil {
			eprintf("failed to encode image %q\n", output_path)
		}
	}

	return true
}

func ext_cwebp(input_path, output_path string, settings *Image_Settings) {
	args := make([]string, 0, 8)

	if settings.width > 0 && settings.height > 0 {
		args = append(args, "-resize", strconv.FormatUint(uint64(settings.width), 10), strconv.FormatUint(uint64(settings.height), 10))
	}

	args = append(args, "-q", strconv.FormatUint(uint64(settings.quality), 10), input_path, "-o", output_path)

	cmd := exec.Command("cwebp", args...)
	err := cmd.Run()
	if err != nil {
		eprintf("failed to call cwebp for %q\n", output_path)
	}
}

// borrowed from "github.com/nfnt/resize" (MIT)
func scaling(max_width, max_height, old_width, old_height uint) (uint, uint) {
	new_width, new_height := old_width, old_height

	if max_width >= old_width && max_height >= old_height {
		return old_width, old_height
	}

	// Preserve aspect ratio
	if old_width > max_width {
		new_height = uint(old_height * max_width / old_width)
		if new_height < 1 {
			new_height = 1
		}
		new_width = max_width
	}

	if new_height > max_height {
		new_width = uint(new_width * max_height / new_height)
		if new_width < 1 {
			new_width = 1
		}
		new_height = max_height
	}

	return new_width, new_height
}