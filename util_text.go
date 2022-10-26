package main

import (
	"strings"

	"unicode"
	"unicode/utf8"
)

var ascii_space = [256]uint8{'\t': 1, '\v': 1, '\f': 1, '\r': 1, ' ': 1}

func make_slug(source string) string {
	buffer := strings.Builder{}

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
	words := strings.Split(input, " ")

	for i, word := range words {
		if i > 0 && short_words[word] {
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
		if c >= utf8.RuneSelf && !unicode.IsSpace(c) {
			break
		}
		if ascii_space[c] == 0 {
			break
		}
		n ++
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
/*
	@todo this should be here
	if !(the_rune == '_' || unicode.IsLetter(the_rune)) {
		return ""
	}
*/
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
const alloc_threshold = 32

// Works on runes (Unicode code points) but does not normalize
// the input strings. See https://blog.golang.org/normalization
// and the golang.org/x/text/unicode/norm package.
func levenshtein_distance(a, b string) int {
	if len(a) == 0 {
		return utf8.RuneCountInString(b)
	}
	if len(b) == 0 {
		return utf8.RuneCountInString(a)
	}
	if a == b {
		return 0
	}

	// We need to convert to []rune if the strings are non-ASCII.
	// This could be avoided by using utf8.RuneCountInString
	// and then doing some juggling with rune indices,
	// but leads to far more bounds checks. It is a reasonable trade-off.
	string_one := []rune(a)
	string_two := []rune(b)

	// swap to save some memory O(min(a,b)) instead of O(a)
	if len(string_one) > len(string_two) {
		string_one, string_two = string_two, string_one
	}

	len_one := len(string_one)
	len_two := len(string_two)

	// Init the row.
	var x []uint16
	if len_one + 1 > alloc_threshold {
		x = make([]uint16, len_one + 1)
	} else {
		// We make a small optimization here for small strings.
		// Because a slice of constant length is effectively an array,
		// it does not allocate. So we can re-slice it to the right length
		// as long as it is below a desired threshold.
		x = make([]uint16, alloc_threshold)
		x = x[:len_one + 1]
	}

	// we start from 1 because index 0 is already 0.
	for i := 1; i < len(x); i++ {
		x[i] = uint16(i)
	}

	// make a dummy bounds check to prevent the 2 bounds check down below.
	// The one inside the loop is particularly costly.
	_ = x[len_one]

	// fill in the rest
	for i := 1; i <= len_two; i++ {
		prev := uint16(i)
		for j := 1; j <= len_one; j++ {
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
	return 0
}