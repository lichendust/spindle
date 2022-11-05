package main

import (
	"os"
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
)

type config struct {
	command uint8
	domain  string

	default_path_mode path_type

	build_drafts    bool
	build_only_used bool

	output_path string

	inline []*regex_entry
}

type TOMLConfig struct {
	Domain            string          `toml:"domain"`
	Default_Path_Mode string          `toml:"default_path_mode"`
	Build_Path        string          `toml:"build_path"`
	Inline            []*Regex_Config `toml:"inline"`
}

func load_config() (*config, bool) {
	blob, ok := load_file(config_file_path)
	if !ok {
		return nil, false // @error
	}

	var conf TOMLConfig
	_, err := toml.Decode(blob, &conf)
	if err != nil {
		return nil, false // @error
	}

	output := config{}

	if x, ok := process_regex_array(conf.Inline); ok {
		output.inline = x
	} else {
		return nil, false // @error
	}

	switch strings.ToLower(conf.Default_Path_Mode) {
	case "relative":
		output.default_path_mode = RELATIVE
	case "absolute":
		output.default_path_mode = ABSOLUTE
	case "root", "rooted":
		output.default_path_mode = ROOTED
	default:
		panic("not a valid path mode in config")
	}

	if conf.Build_Path == "" {
		output.output_path = public_path
 	} else {
		output.output_path = conf.Build_Path
 	}

	output.domain      = conf.Domain

	return &output, true
}

func get_arguments() (*config, bool) {
	args := os.Args[1:]

	config, ok := load_config()
	if !ok {
		panic("failed to load config")
	}

	config.build_only_used = true

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
				counter++
				config.command = BUILD
				continue

			case "serve":
				counter++
				config.command = SERVE
				continue

			case "init":
				config.command = INIT

				if len(args) < 2 {
					return config, true
				}

				path := args[1]

				if !is_valid_path(path) {
					// @error
					return nil, false
				}

				config.output_path = path
				return config, true

			case "help":
				config.command = HELP
				return config, true

			case "version":
				config.command = VERSION
				return config, true
			}
		}

		a, b := pull_argument(args[counter:])

		counter++

		switch a {
		case "":
		case "version":
			config.command = VERSION
			return config, true

		case "help", "h":
			config.command = HELP
			return config, true

		case "all", "a":
			config.build_only_used = false

		default:
			fmt.Fprintf(os.Stderr, "args: %q flag is unknown\n", a)
			has_errors = true

			if b != "" {
				counter ++
			}
		}
	}

	return config, !has_errors
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