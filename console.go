package main

import "fmt"

type console struct {
	scrollback []string
}

func (e *console) flush() {
	if len(e.scrollback) == 0 {
		return
	}
	e.scrollback = make([]string, 0, len(e.scrollback))
}

func (e *console) print(base string, args ...interface{}) {
	e.scrollback = append(e.scrollback, fmt.Sprintf(base, args...))
}