package main

import (
	"os"
	"fmt"
)

type config struct {
	command uint8
	domain  string
	path    string

	default_path_type path_type
}

func get_arguments() (*config, bool) {
	args   := os.Args[1:]
	config := &config{}

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

				config.path = path
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
			break

		case "version":
			config.command = VERSION
			return config, true

		case "help", "h":
			config.command = HELP
			return config, true

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