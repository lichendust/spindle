package main

import (
	"strings"

	"path"
	"path/filepath"
)

func rewrite_image_path(image_path, image_prefix string, webp_enabled bool) string {
	if strings.HasPrefix(image_path, "http") {
		return image_path
	}

	if !strings.HasPrefix(image_path, image_prefix) {
		image_path = path.Join(image_prefix, image_path)
	}

	ext := filepath.Ext(image_path)

	if webp_enabled {
		if x, ok := config.image_ext[ext[1:]]; ok {
			image_path = image_path[:len(image_path) - len(ext) + 1] + x

		}
	}

	return image_path
}