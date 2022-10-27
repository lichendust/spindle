package main

import "unicode/utf8"

type lexer_token struct {
	ast_type   ast_type
	position   uint16
	field      string
}

var rune_match = map[rune]ast_type {
	'\n': NEWLINE,
	'\\': ESCAPE,
	'/':  FORWARD_SLASH,
	'|':  PIPE,
	'%':  PERCENT,
	'{':  BRACE_OPEN,
	'}':  BRACE_CLOSE,
	'[':  BRACKET_OPEN,
	']':  BRACKET_CLOSE,
	'<':  ANGLE_OPEN,
	'>':  ANGLE_CLOSE,
	'~':  TILDE,
	'&':  AMPERSAND,
	'=':  EQUALS,
	'.':  STOP,
	':':  COLON,
	'+':  PLUS,
	'×':  MULTIPLY,
}

func lex_blob(input string) []*lexer_token {
	array   := make([]*lexer_token, 0, 512)
	line_no := uint16(1)

	for {
		i := eat_spaces(input) // ignores newlines

		if i > 0 {
			array = append(array, &lexer_token {
				ast_type: WHITESPACE,
				position: line_no,
				field: input[:i],
			})
			input = input[i:]
			continue
		}

		if len(input) == 0 {
			break
		}

		rune, width := utf8.DecodeRuneInString(input)

		if rune == '\n' {
			line_no ++
		}

		if result, ok := rune_match[rune]; ok {
			input = input[width:]
			array = append(array, &lexer_token {
				ast_type: result,
				position: line_no,
				field: string(rune),
			})
			continue
		}

		number := extract_number(input)

		if n := len(number); n > 0 {
			input = input[n:]
			array = append(array, &lexer_token {
				ast_type: NUMBER,
				position: line_no,
				field: number,
			})
			continue
		}

		ident := extract_ident(input)
		word  := extract_word(ident)

		// if the ident is different to the word
		// then the ident is the winner
		if n := len(ident); n != len(word) {
			input = input[n:]
			array = append(array, &lexer_token {
				ast_type: IDENT,
				position: line_no,
				field: ident,
			})
			continue
		}

		// otherwise it's the word, we use that instead
		if n := len(word); n > 0 {
			input = input[n:]
			array = append(array, &lexer_token {
				ast_type: WORD,
				position: line_no,
				field: word,
			})
			continue
		}

		// if the word was length zero, then we'll
		// try for a token (such as ###)
		non_word := extract_repeated_rune(input)

		if len(non_word) > 0 {
			the_type := NON_WORD

			// these are the only formatter char that gets repeated
			if non_word[0] == '*' {
				the_type = ASTERISK
			}

			input = input[len(non_word):]
			array = append(array, &lexer_token {
				ast_type: the_type,
				position: line_no,
				field:    non_word,
			})
			continue
		}

		break
	}

	// this removes leading and trailing whitespace, but it
	// occurs to me that this may be necessary for the "raw"
	// feature if a snippet/chunk is nested inside

	/*for i, entry := range array {
		if entry.ast_type == NEWLINE || entry.ast_type == WHITESPACE {
			continue
		}
		array = array[i:]
		break
	}

	for i := len(array) - 1; i >= 0; i-- {
		type_check := array[i].ast_type
		if type_check == NEWLINE || type_check == WHITESPACE {
			continue
		}
		array = array[:i - 1]
		break
	}*/

	array = append(array, &lexer_token {
		ast_type: EOF,
		position: line_no,
	})

	return array
}