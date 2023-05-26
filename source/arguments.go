package main

import (
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

type config struct {
	command uint8
	domain  string

	path_mode path_type

	build_drafts    bool
	build_only_used bool
	skip_images     bool
	tag_path        string

	image_quality  int
	image_max_size uint
	image_format   file_type

	output_path string

	inline []*regex_entry
}

type TOMLConfig struct {
	Domain            string          `toml:"domain"`
	Default_Path_Mode string          `toml:"path_mode"`
	Build_Path        string          `toml:"build_path"`
	Tag_Path          string          `toml:"tag_path"`
	Inline            []*Regex_Config `toml:"inline"`

	Image_Quality  int    `toml:"image_quality"`
	Image_Max_Size uint   `toml:"image_size"`
	Image_Format   string `toml:"image_format"`
}

func load_config() (config, bool) {
	blob, ok := load_file(CONFIG_FILE_PATH)
	if !ok {
		return config{}, false // @error
	}

	var conf TOMLConfig
	_, err := toml.Decode(blob, &conf)
	if err != nil {
		return config{}, false // @error
	}

	output := config{}

	if x, ok := process_regex_array(conf.Inline); ok {
		output.inline = x
	} else {
		return config{}, false // @error
	}

	switch strings.ToLower(conf.Default_Path_Mode) {
	case "relative":
		output.path_mode = RELATIVE
	case "absolute":
		output.path_mode = ABSOLUTE
	case "root", "rooted":
		output.path_mode = ROOTED
	default:
		panic("not a valid path mode in config")
	}

	if conf.Build_Path == "" {
		output.output_path = PUBLIC_PATH
	} else {
		output.output_path = conf.Build_Path
	}
	if conf.Tag_Path == "" {
		output.tag_path = "tag"
	} else {
		output.tag_path = conf.Tag_Path
	}

	output.image_quality  = conf.Image_Quality
	output.image_max_size = conf.Image_Max_Size

	switch strings.ToLower(conf.Image_Format) {
	case "webp":
		output.image_format = IMG_WEB
	case "jpg", "jpeg":
		output.image_format = IMG_JPG
	case "tif", "tiff":
		output.image_format = IMG_TIF
	case "png":
		output.image_format = IMG_PNG
	}

	output.domain = conf.Domain

	return output, true
}

func get_arguments() (config, bool) {
	args := os.Args[1:]

	conf, ok := load_config()
	if !ok {
		conf = config{}
	}

	conf.build_only_used = true

	counter    := 0
	has_errors := false

	for {
		args = args[counter:]

		if len(args) == 0 {
			break
		}

		counter = 0

		if len(args) > 0 {
			switch args[0] {
			case "build":
				counter += 1
				conf.command = COMMAND_BUILD
				continue

			case "serve":
				counter += 1
				conf.command = COMMAND_SERVE
				continue

			case "init":
				conf.command = COMMAND_INIT

				if len(args) < 2 {
					return conf, true
				}

				path := args[1]

				if !is_valid_path(path) {
					// @error
					return conf, false
				}

				conf.output_path = path
				return conf, true

			case "help":
				conf.command = COMMAND_HELP
				return conf, true

			case "version":
				conf.command = COMMAND_VERSION
				return conf, true
			}
		}

		a, b := pull_argument(args[counter:])

		counter += 1

		switch a {
		case "":
		case "version":
			conf.command = COMMAND_VERSION
			return conf, true

		case "help", "h":
			conf.command = COMMAND_HELP
			return conf, true

		case "all", "a":
			conf.build_only_used = false

		case "skip-images":
			conf.skip_images = true

		default:
			eprintf("args: %q flag is unknown\n", a)
			has_errors = true

			if b != "" {
				counter += 1
			}
		}
	}

	return conf, !has_errors
}

func pull_argument(args []string) (string, string) {
	if len(args) == 0 {
		return "", ""
	}

	if len(args[0]) >= 1 {
		n := count_rune(args[0], '-')
		a := args[0]

		if n > 0 {
			a = a[n:]
		} else {
			return "", ""
		}

		if len(args[1:]) >= 1 {
			b := args[1]

			if len(b) > 0 && b[0] != '-' {
				return a, b
			}
		}

		return a, ""
	}

	return "", ""
}