package main

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// this file is isolated as a constant
// reminder to replace it with something
// faster

var strike  = regexp.MustCompile(`~(\S(.+?)\S)~`)
var italics = regexp.MustCompile(`\*(\S(.+?)\S)\*`)
var bold    = regexp.MustCompile(`\*\*(\S(.+?)\S)\*\*`)
var link    = regexp.MustCompile(`\[(.+?)\]\((.+?)\)`)
var link2   = regexp.MustCompile(`\[\[(.+?)\]\]\((.+?)\)`)
var inline  = regexp.MustCompile(`c\.(.+?){(.+?)}`)

var code = regexp.MustCompile("`(.+?)`")

func format_inlines(v string, vars map[string]string) string {
	input := []byte(v)

	code_temp  := sprint(vars["code"],  "$1")
	link_temp  := sprint(vars["link"],  "$2", "$1")
	link2_temp := sprint(vars["link2"], "$2", "$1")

	input = code.ReplaceAll(input,    []byte(code_temp)) // inline needs ReplaceFunc with inline_sub
	input = link2.ReplaceAll(input,   []byte(link2_temp))
	input = link.ReplaceAll(input,    []byte(link_temp))
	input = bold.ReplaceAll(input,    []byte(`<b>$1</b>`))
	input = italics.ReplaceAll(input, []byte(`<i>$1</i>`))
	input = strike.ReplaceAll(input,  []byte(`<s>$1</s>`))

	return string(input)
}

func strip_inlines(v string) string {
	input := []byte(v)

	input = link.ReplaceAll(input,    []byte(`$1`))
	input = code.ReplaceAll(input,    []byte(`$1`))
	input = bold.ReplaceAll(input,    []byte(`$1`))
	input = italics.ReplaceAll(input, []byte(`$1`))
	input = strike.ReplaceAll(input,  []byte(`$1`))

	return string(input)
}

func inline_code_sub(input string) string {
	buffer := strings.Builder {}
	buffer.Grow(len(input) + 128)

	count := 0

	for _, c := range input {
		count += utf8.RuneLen(c)

		switch c {
		case '\t':
			buffer.WriteString(input[:count-1])
			input = input[count:]
			buffer.WriteString(`    `)
			count = 0

		case '&':
			buffer.WriteString(input[:count-1])
			input = input[count:]
			buffer.WriteString(`&amp;`)
			count = 0

		case '<':
			buffer.WriteString(input[:count-1])
			input = input[count:]
			buffer.WriteString(`&lt;`)
			count = 0

		case '>':
			buffer.WriteString(input[:count-1])
			input = input[count:]
			buffer.WriteString(`&gt;`)
			count = 0
		}
	}

	buffer.WriteString(input)

	return buffer.String()
}

func process_code(input string) string {
	input = inline_code_sub(input)

	indent := 0

	input = input[1:] // leading newline

	for i, r := range input {
		if r != ' ' {
			indent = i
			break
		}
	}

	input = strings.TrimSpace(input)

	buffer := strings.Builder {}
	buffer.Grow(len(input))

	for len(input) > 0 {
		line := extract_to_newline(input)
		n    := len(line)

		if len(input) > n {
			input = input[n+1:]
		} else {
			input = input[n:]
		}

		for i, c := range line {
			if i < indent && c == ' ' {
				continue
			}
			buffer.WriteString(line[i:])
			break
		}

		if len(input) > 0 {
			buffer.WriteRune('\n')
		}
	}

	return buffer.String()
}