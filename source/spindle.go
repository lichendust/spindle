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

import "os"
import "sync"
import "strings"

import "github.com/BurntSushi/toml"

var spindle *Spindle

type Spindle struct {
	server_mode bool

	errors    *Error_Handler
	file_tree *File

	Config

	has_webp bool

	cache_lock *sync.WaitGroup

	finder_cache map[string]*File

	pages        map[string]*Page
	templates    map[uint32]*Support_Markup
	partials     map[uint32]*Support_Markup

	gen_pages    map[string]*Gen_Page
	gen_images   map[uint32]*Image
}

type Config struct {
	command uint8
	domain  string

	path_mode Path_Type

	build_drafts    bool
	build_only_used bool
	skip_images     bool

	port_number string

	image_quality  int
	image_max_size uint
	image_format   File_Type

	sitemap bool

	output_path string

	inline []*Regex_Entry
}

func load_config() (Config) {
	type TOMLConfig struct {
		Domain            string          `toml:"domain"`
		Default_Path_Mode string          `toml:"path_mode"`
		Build_Path        string          `toml:"build_path"`
		Inline            []*Regex_Config `toml:"inline"`

		Image_Quality  int                `toml:"image_quality"`
		Image_Max_Size uint               `toml:"image_size"`
		Image_Format   string             `toml:"image_format"`

		Sitemap bool       `toml:"sitemap"`
		Port_Number string `toml:"port_number"`
	}

	blob, ok := load_file(CONFIG_FILE_PATH)
	if !ok {
		return Config{}
	}

	var conf TOMLConfig
	_, err := toml.Decode(blob, &conf)
	if err != nil {
		return Config{} // @error
	}

	output := Config{}

	if x, ok := process_regex_array(conf.Inline); ok {
		output.inline = x
	} else {
		return Config{} // @error
	}

	switch strings.ToLower(conf.Default_Path_Mode) {
	case "relative":
		output.path_mode = RELATIVE
	case "absolute":
		output.path_mode = ABSOLUTE
	case "root", "rooted":
		output.path_mode = ROOTED
	default:
		eprintf("config: %q is not a valid path mode\n", conf.Default_Path_Mode)
	}

	if conf.Build_Path == "" {
		output.output_path = PUBLIC_PATH
	} else {
		output.output_path = conf.Build_Path
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

	output.domain      = conf.Domain
	output.sitemap     = conf.Sitemap
	output.port_number = conf.Port_Number

	return output
}

func get_arguments() (Config, bool) {
	args := os.Args[1:]
	conf := load_config()

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

		case "p", "port":
			if b != "" {
				conf.port_number = b
				counter += 1
			}

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

	if conf.port_number == "" {
		conf.port_number = SERVE_PORT
	} else if !strings.HasPrefix(conf.port_number, ":") {
		conf.port_number = ":" + conf.port_number
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