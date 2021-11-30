package main

import (
	"strings"

	"path"
	"path/filepath"
)



// we could do webp for GIFs too, but magick
// doesn't have animation target flags and
// i'm loathed to add cwebp as a dep.
var rewrite_image_table = map[string]string {}

func rewrite_image_path(image_path, image_prefix string, webp_enabled bool) string {
	if strings.HasPrefix(image_path, "http") {
		return image_path
	}

	if !strings.HasPrefix(image_path, image_prefix) {
		image_path = path.Join(image_prefix, image_path)
	}

	ext := filepath.Ext(image_path)

	if webp_enabled {
		if x, ok := rewrite_image_table[ext[1:]]; ok {
			image_path = image_path[:len(image_path) - len(ext) + 1] + x

		}
	}

	return image_path
}