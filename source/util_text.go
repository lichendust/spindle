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

import "time"
import "strings"
import "strconv"
import "unicode"
import "unicode/utf8"

var ascii_space = [256]uint8{'\t': 1, '\v': 1, '\f': 1, '\r': 1, ' ': 1}

// utility to simplify sequential buffer writes
func write_to(buffer *strings.Builder, text ...string) {
	for _, x := range text {
		buffer.WriteString(x)
	}
}

func make_slug(source string) string {
	buffer := strings.Builder{}
	buffer.Grow(len(source))

	inside_element := false

	for _, c := range source {
		if c == '<' {
			inside_element = true
			continue
		}
		if c == '>' {
			inside_element = false
			buffer.WriteRune('-')
			continue
		}
		if inside_element {
			continue
		}

		if unicode.IsLetter(c) || unicode.IsNumber(c) {
			buffer.WriteRune(unicode.ToLower(c))
			continue
		}
		if unicode.IsSpace(c) || c == '-' {
			buffer.WriteRune('-')
			continue
		}
	}

	return buffer.String()
}

func truncate(input string, l int) string {
	counter := 0
	for i := range input {
		counter += 1
		if counter >= l {
			input = input[:i]
			break
		}
	}
	return input + "..."
}

// https://www.grammarly.com/blog/capitalization-in-the-titles/
// @todo complete this
var short_words = map[string]bool {
	"a":    true,
	"an":   true,
	"and":  true,
	"the":  true,
	"on":   true,
	"to":   true,
	"in":   true,
	"for":  true,
	"nor":  true,
	"or":   true,
	"from": true,
	"but":  true,
	"is":   true,
}

// a title caser that actually works!
func make_title(input string) string {
	short_words := func(x string) bool {
		switch x {
		case "a":    return true
		case "an":   return true
		case "and":  return true
		case "the":  return true
		case "on":   return true
		case "to":   return true
		case "in":   return true
		case "for":  return true
		case "nor":  return true
		case "or":   return true
		case "from": return true
		case "but":  return true
		case "is":   return true
		}
		return false
	}

	words := strings.Split(input, " ")

	for i, word := range words {
		if i > 0 && short_words(word) {
			continue
		}

		buffer := strings.Builder {}
		buffer.Grow(len(word))

		for len(word) > 0 {
			c, width := utf8.DecodeRuneInString(word)

			if buffer.Len() == 0 {
				buffer.WriteRune(unicode.ToUpper(c))
				word = word[width:]
				continue
			}

			if c == '-' || c == '—' {
				buffer.WriteRune(unicode.ToLower(c))
				word = word[width:]

				c, width = utf8.DecodeRuneInString(word)

				buffer.WriteRune(unicode.ToUpper(c))
				word = word[width:]
				continue
			}

			buffer.WriteRune(unicode.ToLower(c))
			word = word[width:]
		}

		words[i] = buffer.String()
	}

	return strings.Join(words, " ")
}

func eat_spaces(s string) int {
	n := 0

	for _, c := range s {
		if c >= utf8.RuneSelf {
			if !unicode.IsSpace(c) {
				break
			}
			continue
		}
		if ascii_space[c] == 0 {
			break
		}
		n += 1
	}

	return n
}

func extract_word(input string) string {
	r, _ := utf8.DecodeRuneInString(input)

	if unicode.IsNumber(r) {
		return ""
	}

	for i, c := range input {
		if !(unicode.IsLetter(c) || unicode.IsNumber(c)) {
			return input[:i]
		}
	}

	return input
}

func extract_ident(input string) string {
	the_rune, _ := utf8.DecodeRuneInString(input)

	if unicode.IsNumber(the_rune) {
		return ""
	}

	if !(the_rune == '_' || unicode.IsLetter(the_rune)) {
		return ""
	}

	for i, c := range input {
		if !(c == '_' || unicode.IsLetter(c) || unicode.IsNumber(c)) {
			return input[:i]
		}
	}

	return input
}

func extract_number(input string) string {
	for i, c := range input {
		if !unicode.IsNumber(c) {
			return input[:i]
		}
	}

	return input
}

func extract_repeated_rune(input string) string {
	r, _ := utf8.DecodeRuneInString(input)

	for i, c := range input {
		if c == r {
			continue
		}
		return input[:i]
	}
	return input
}

func extract_non_space_word(input string) string {
	for i, c := range input {
		if c >= utf8.RuneSelf && !unicode.IsSpace(c) {
			return input[:i]
		}
		if ascii_space[c] != 0 {
			return input[:i]
		}
	}
	return input
}

func unix_args(input string) []string {
	is_quote := false

	args := make([]string, 0, 4)

	for {
		if len(input) == 0 {
			break
		}

		if is_quote {
			for i, c := range input {
				if c == '"' {
					is_quote = false
					args = append(args, input[:i])
					input = input[i + 1:]
					break
				}
			}
			continue
		}

		input = strings.TrimSpace(input)

		prefix := rune(input[0])

		if prefix == '"' {
			is_quote = true
			input = input[1:]
			continue
		}

		word := extract_non_space_word(input)
		args = append(args, word)
		input = input[len(word):]
	}

	return args
}

// levenshtein implementation taken from
// https://github.com/agnivade/levenshtein [MIT]

// Works on runes (Unicode code points) but does not normalize
// the input strings. See https://blog.golang.org/normalization
// and the golang.org/x/text/unicode/norm package.
func levenshtein_distance(a, b string) int {
	const alloc_threshold = 32

	if len(a) == 0 {
		return utf8.RuneCountInString(b)
	}
	if len(b) == 0 {
		return utf8.RuneCountInString(a)
	}
	if a == b {
		return 0
	}

	string_one := []rune(a)
	string_two := []rune(b)

	// swap to save some memory O(min(a,b)) instead of O(a)
	if len(string_one) > len(string_two) {
		string_one, string_two = string_two, string_one
	}

	len_one := len(string_one)
	len_two := len(string_two)

	var x []uint16
	if len_one + 1 > alloc_threshold {
		x = make([]uint16, len_one + 1)
	} else {
		x = make([]uint16, alloc_threshold)
		x = x[:len_one + 1]
	}

	for i := 1; i < len(x); i += 1 {
		x[i] = uint16(i)
	}

	_ = x[len_one]

	for i := 1; i <= len_two; i += 1 {
		prev := uint16(i)
		for j := 1; j <= len_one; j += 1 {
			current := x[j - 1] // match
			if string_two[i - 1] != string_one[j - 1] {
				current = min(min(x[j - 1] + 1, prev + 1), x[j] + 1)
			}
			x[j - 1] = prev
			prev = current
		}
		x[len_one] = prev
	}
	return int(x[len_one])
}

func min(a, b uint16) uint16 {
	if a < b {
		return a
	}
	return b
}

func count_rune(input string, r rune) int {
	for i, c := range input {
		if c != r {
			return i
		}
	}
	return len(input)
}

func nsdate(input string) string {
	nsconvert := func(x string) string {
		switch x {
		case "M":    return "1"
		case "MM":   return "01"
		case "MMM":  return "Jan"
		case "MMMM": return "January"
		case "d":    return "2"
		case "dd":   return "02"
		case "E":    return "Mon"
		case "EEEE": return "Monday"
		case "h":    return "3"
		case "hh":   return "03"
		case "HH":   return "15"
		case "a":    return "PM"
		case "m":    return "4"
		case "mm":   return "04"
		case "s":    return "5"
		case "ss":   return "05"
		case "SSS":  return ".000"
		}
		return ""
	}

	clamp := func(m int) int {
		if m < 0 {
			return 0
		}
		return m
	}

	final := strings.Builder{}
	final.Grow(128)

	t := time.Now()

	input = strings.TrimSpace(input)

	for {
		if len(input) == 0 {
			break
		}

		for _, c := range input {
			if unicode.IsLetter(c) {
				n := count_rune(input, c)
				repeat := input[:n]

				// years
				if c == 'y' {
					switch n {
					case 1:
						final.WriteString(strconv.Itoa(t.Year()))
					case 2:
						final.WriteString(t.Format("06"))
					default:
						y := strconv.Itoa(t.Year())
						final.WriteString(strings.Repeat("0", clamp(n - len(y))))
						final.WriteString(y)
					}
					input = input[n:]
					break
				}

				// H - unpadded hour
				if c == 'H' && n == 1 {
					final.WriteString(strconv.Itoa(t.Hour()))
				}
				// MMMMM - single letter month
				if c == 'M' && n == 5 {
					final.WriteString(t.Month().String()[:1])
				}
				// EEEEE - single letter week
				if c == 'E' && n == 5 {
					final.WriteString(t.Weekday().String()[:1])
				}
				// EEEEEE - two letter week
				if c == 'E' && n == 6 {
					final.WriteString(t.Weekday().String()[:2])
				}

				if x := nsconvert(repeat); x != "" {
					final.WriteString(t.Format(x))
				} else {
					final.WriteString(repeat)
				}

				input = input[n:]
				break
			}

			final.WriteRune(c)
			input = input[1:]
		}
	}

	return final.String()
}