package main

import (
	"net/url"
	"strings"
	"unicode"
	"unicode/utf8"
)

var ascii_space = [256]uint8{'\t': 1, '\n': 1, '\v': 1, '\f': 1, '\r': 1, ' ': 1}

func consume_whitespace(input string) string {
	start := 0

	for ; start < len(input); start++ {
		c := input[start]

		if c >= utf8.RuneSelf {
			return strings.TrimFunc(input[start:], unicode.IsSpace)
		}

		if ascii_space[c] == 0 {
			break
		}
	}

	return input[start:]
}

func extract_ident(input string) string {
	for i, c := range input {
		if !(unicode.IsLetter(c) || unicode.IsNumber(c) || c == '_' || c == '.') {
			return input[:i]
		}
	}
	return input
}

func extract_arbitrary_word(input string) string {
	for i, c := range input {
		if unicode.IsSpace(c) {
			return input[:i]
		}
	}
	return input
}

func extract_to_newline(input string) string {
	for i, c := range input {
		if c == '\n' {
			return input[:i]
		}
	}
	return input
}

func count_rune(input string, r rune) int {
	count := 0
	for _, c := range input {
		if c != r {
			return count
		}
		count++
	}
	return count
}

func sprint(source string, v ...string) string {
	if len(v) == 1 {
		return strings.ReplaceAll(source, `%s`, v[0])
	}

	for _, x := range v {
		source = strings.Replace(source, `%s`, x, 1)
	}

	return source
}

func split_rune(input string, r rune) []string {
	for i, c := range input {
		if c == r {
			if i == len(input) {
				return []string {input}
			}
			return []string {input[:i], input[i + 1:]}
		}
	}
	return []string {input}
}

func make_element_id(source string) string {
	new := strings.Builder {}

	inside_element := false

	for _, c := range source {
		if c == '<' {
			inside_element = true
			continue
		}

		if c == '>' {
			inside_element = false
			new.WriteRune('-')
			continue
		}

		if inside_element {
			continue
		}

		if unicode.IsLetter(c) || unicode.IsNumber(c) {
			new.WriteRune(c)
			continue
		}

		if unicode.IsSpace(c) || c == '-' {
			new.WriteRune('-')
			continue
		}
	}

	return strings.ToLower(new.String())
}

func unix_args(input string) []string {
	input = strings.TrimSpace(input)

	last_quote := 'x'
	is_quote   := false

	args := make([]string, 0, 4)

	for {
		if len(input) == 0 {
			break
		}

		if is_quote {
			for i, c := range input {
				if last_quote == c {
					last_quote = 'x'
					is_quote = false

					args = append(args, input[:i])

					input = input[i:]
					continue
				}
			}
		}

		input = consume_whitespace(input)

		prefix := rune(input[0])

		if prefix == '"' || prefix == '\'' {
			is_quote = true
			last_quote = prefix
			input = input[1:]
			continue
		}

		word := extract_arbitrary_word(input)
		args = append(args, word)
		input = input[len(word):]
	}

	return args
}

func clean_html(input string) string {
	buffer := strings.Builder {}
	buffer.Grow(len(input))

	inside_pre := false
	last := 'x'

	for len(input) > 0 {
		if strings.HasPrefix(input, "<pre>") {
			inside_pre = true
		}
		if strings.HasPrefix(input, "</pre>") {
			inside_pre = false
		}

        c, width := utf8.DecodeRuneInString(input)

		input = input[width:]

		if !inside_pre && (c == '\n' || c == '\t') {
			continue
		}

		// collapse repeat spaces to singles
		if c == ' ' && last == ' ' {
			continue
		}

		// rewrite escaped braces
		if (c == '}' || c == '}') && last == '\\' {
			buffer.WriteRune(c)
		}

		// replace all tabs with four spaces
		if c == '\t' {
			buffer.WriteString("    ") // 4
			continue
		}

		last = c

		buffer.WriteRune(c)
	}

	return buffer.String()
}

func complex_key_mapper(source string, vars map[string]string) string {
	source = format_inlines(source, vars)

	if strings.IndexRune(source, '$') < 0 {
		return source
	}

	buffer := strings.Builder {}
	buffer.Grow(len(source) * 2)

	for {
		i := strings.IndexRune(source, '$')

		if i < 0 {
			buffer.WriteString(source)
			break
		}

		buffer.WriteString(source[:i])

		if source[i + 1] == '{' {
			var_text := extract_code_block(source[i + 2:]) // +1 trailing brace
			source = source[i + len(var_text) + 3:]
			var_text = strings.TrimSpace(var_text)

			the_var := extract_ident(var_text)

			var_text = consume_whitespace(var_text[len(the_var):])

			sub_text, ok := vars[the_var]

			if len(var_text) == 0 {
				if ok {
					buffer.WriteString(simple_key_mapper(sub_text, vars))
				}
				continue
			}

			switch var_text[0] {
			case ':':
				var_text = strings.TrimSpace(var_text[1:])

				parts := split_rune(var_text, '|')

				if len(parts) == 1 {
					if ok {
						buffer.WriteString(simple_key_mapper(parts[0], vars))
					}
				} else {
					if ok {
						buffer.WriteString(simple_key_mapper(strings.TrimSpace(parts[0]), vars))
					} else {
						buffer.WriteString(simple_key_mapper(strings.TrimSpace(parts[1]), vars))
					}
				}

			case '|':
				var_text = strings.TrimSpace(var_text[1:])

				if ok {
					buffer.WriteString(sub_text)
				} else {
					buffer.WriteString(var_text)
				}
			}

		} else {
			buffer.WriteRune('$')
			source = source[i+1:]
		}
	}

	return buffer.String()
}

func simple_key_mapper(source string, vars map[string]string) string {
	if strings.IndexRune(source, '$') < 0 {
		return source
	}

	buffer := strings.Builder {}
	buffer.Grow(len(source) * 2)

	for {
		i := strings.IndexRune(source, '$')

		if i < 0 {
			buffer.WriteString(source)
			break
		}

		buffer.WriteString(source[:i])

		if source[i + 1] == '{' {
			var_text := extract_code_block(source[i + 2:]) // +1 trailing brace
			source = source[i + len(var_text) + 3:]
			var_text = strings.TrimSpace(var_text)

			the_var := extract_ident(var_text)

			var_text = consume_whitespace(var_text[len(the_var):])

			sub_text, ok := vars[the_var]

			if len(var_text) == 0 {
				if ok {
					buffer.WriteString(sub_text)
				}
				continue
			}

		} else {
			buffer.WriteRune('$')
			source = source[i+1:]
		}
	}

	return buffer.String()
}

func join_url(a, b string) string {
	u, err := url.Parse(b)

	if err != nil {
		panic(err)
	}

	base, err := url.Parse(a)

	if err != nil {
		panic(err)
	}

	return base.ResolveReference(u).String()
}