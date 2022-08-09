package main

import (
	"fmt"
	"strings"
)

type warning_handler struct {
	warnings []string
	repeated map[string]bool
}

func new_warning_handler() *warning_handler {
	return &warning_handler {
		warnings: make([]string, 0, 64),
		repeated: make(map[string]bool, 64),
	}
}

func (w *warning_handler) new(base string, args ...interface{}) {
	compiled := fmt.Sprintf(base, args...)

	if w.repeated[compiled] {
		return
	}

	w.repeated[compiled] = true
	w.warnings = append(w.warnings, compiled)
}

func (w *warning_handler) print_all() {
	for _, warning := range w.warnings {
		if strings.Contains(warning, "image") && strings.Contains(warning, "@") {
			continue
		}

		fmt.Println(warning)
	}
}