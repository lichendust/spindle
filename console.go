package main

import "fmt"

type console struct {
	last_len int
	scrollback []string
}

func (e *console) flush() {
	if len(e.scrollback) == 0 {
		return
	}

	for i := 0; i < e.last_len; i++ {
		fmt.Printf("\033[F\033[K")
	}

	for _, x := range e.scrollback {
		fmt.Println(x)
	}

	e.last_len   = len(e.scrollback)
	e.scrollback = make([]string, 0, e.last_len)
}

func (e *console) print(base string, args ...interface{}) {
	e.scrollback = append(e.scrollback, fmt.Sprintf(base, args...))
}