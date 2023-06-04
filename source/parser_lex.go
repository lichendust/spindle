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

import "unicode"
import "unicode/utf8"

type Position struct {
	line       int
	start      int
	end        int
	file_path  string
}

type Lexer_Token struct {
	ast_type   AST_Type
	position   Position
	field      string
}

func rune_match(r rune) (AST_Type, bool) {
	switch r {
	case '\n': return NEWLINE,       true
	case '\\': return ESCAPE,        true
	case '/':  return FORWARD_SLASH, true
	case '|':  return PIPE,          true
	case '%':  return PERCENT,       true
	case '{':  return BRACE_OPEN,    true
	case '}':  return BRACE_CLOSE,   true
	case '[':  return BRACKET_OPEN,  true
	case ']':  return BRACKET_CLOSE, true
	case '<':  return ANGLE_OPEN,    true
	case '>':  return ANGLE_CLOSE,   true
	case '~':  return TILDE,         true
	case '&':  return AMPERSAND,     true
	case '=':  return EQUALS,        true
	case '!':  return BANG,          true
	case '.':  return STOP,          true
	case ':':  return COLON,         true
	case '+':  return PLUS,          true
	case '×':  return MULTIPLY,      true
	case '$':  return DOLLAR,        true
	}
	return '0', false
}

func lex_blob(path, input string) []*Lexer_Token{
	array   := make([]*Lexer_Token, 0, 512)
	line_no := 1

	start := 0
	end   := 0

	for {
		i := eat_spaces(input) // ignores newlines

		if i > 0 {
			end += i
			array = append(array, &Lexer_Token{
				ast_type: WHITESPACE,
				position: Position{line_no, start, end, path},
				field:    input[:i],
			})
			input = input[i:]
			start = end
			continue
		}

		if len(input) == 0 {
			break
		}

		rune, width := utf8.DecodeRuneInString(input)

		if rune == '\n' {
			line_no += 1
		}

		if result, ok := rune_match(rune); ok {
			end += width
			array = append(array, &Lexer_Token{
				ast_type: result,
				position: Position{line_no, start, end, path},
				field: string(rune),
			})
			input = input[width:]
			start = end
			continue
		}

		number := extract_number(input)

		if n := len(number); n > 0 {
			end += len(number)
			array = append(array, &Lexer_Token{
				ast_type: NUMBER,
				position: Position{line_no, start, end, path},
				field:    number,
			})
			input = input[n:]
			start = end
			continue
		}

		ident := extract_ident(input)
		word  := extract_word(ident)

		// if the ident is different to the word
		// then the ident is the winner
		if n := len(ident); n != len(word) {
			end += n
			array = append(array, &Lexer_Token{
				ast_type: IDENT,
				position: Position{line_no, start, end, path},
				field:    ident,
			})
			input = input[n:]
			start = end
			continue
		}

		// otherwise it's the word, we use that instead
		if n := len(word); n > 0 {
			end += n
			array = append(array, &Lexer_Token{
				ast_type: WORD,
				position: Position{line_no, start, end, path},
				field:    word,
			})
			input = input[n:]
			start = end
			continue
		}

		// if the word was length zero, then we'll
		// try for a token (such as ###)
		non_word := extract_repeated_rune(input)

		if n := len(non_word); n > 0 {
			the_type := NON_WORD

			if non_word[0] == '*' {
				the_type = ASTERISK // @todo
			}

			{
				r, _ := utf8.DecodeRuneInString(non_word)
				if unicode.IsSymbol(r) {
					the_type = WORD
				}
			}

			end += n

			array = append(array, &Lexer_Token{
				ast_type: the_type,
				position: Position{line_no, start, end, path},
				field:    non_word,
			})
			input = input[n:]
			start = end
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

	for i := len(array) - 1; i >= 0; i -= 1 {
		type_check := array[i].ast_type
		if type_check == NEWLINE || type_check == WHITESPACE {
			continue
		}
		array = array[:i - 1]
		break
	}*/

	array = append(array, &Lexer_Token{
		ast_type: EOF,
		position: Position{line_no, start, end, path},
	})

	return array
}