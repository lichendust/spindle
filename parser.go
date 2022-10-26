package main

import "strings"
import "strconv"

type parser struct {
	index int
	array []*lexer_token
}

func parse_stream(array []*lexer_token) []ast_data {
	parser := parser { array: array	}
	return parser.parse_block(EOF)
}

func (parser *parser) parse_block(exit_upon ast_type) []ast_data {
	array := make([]ast_data, 0, 32)

	for {
		parser.eat_whitespace()

		token := parser.next()

		if token == nil {
			break
		}
		if token.ast_type == exit_upon {
			break
		}
		if token.ast_type == NEWLINE {
			if p := parser.prev(); p != nil && p.ast_type == NEWLINE {
				array = append(array, &ast_normal{ ast_type: BLANK })
			}
			continue
		}

		switch token.ast_type {
		case FORWARD_SLASH:
			if p := parser.peek(); p != nil && p.ast_type != WHITESPACE {
				break // not a token, has to be spaced
			}
			parser.eat_comment()
			continue

		case NON_WORD:
			new_tok := &ast_token {
				decl_hash: new_hash(token.field),
			}

			if p := parser.peek(); p != nil && p.ast_type != WHITESPACE {
				break // not a token, has to be spaced
			}

			parser.eat_whitespace()
			new_tok.children = parser.parse_paragraph(NULL)
			new_tok.position = token.position

			array = append(array, new_tok)
			continue

		case AMPERSAND, ANGLE_CLOSE, TILDE:
			new_tok := &ast_normal{}

			switch token.ast_type {
			case AMPERSAND:   new_tok.ast_type = TEMPLATE
			case ANGLE_CLOSE: new_tok.ast_type = PARTIAL // don't exist yet
			case TILDE:       new_tok.ast_type = IMPORT
			}

			parser.eat_whitespace()

			x := parser.peek()

			if x.ast_type.is(WORD, IDENT) {
				parser.next()
				new_tok.field = x.field
			} else {
				panic("bad declaration")
			}

			parser.eat_whitespace()

			if parser.peek().ast_type != NEWLINE {
				panic("unexpected item in the scope area")
			}

			new_tok.position = token.position
			array = append(array, new_tok)
			continue

		case WORD, IDENT:
			x := parser.peek_whitespace()

			if x.ast_type == BRACE_OPEN {
				parser.next()
				parser.eat_whitespace()
				parser.next()

				new_tok := &ast_block {
					decl_hash: new_hash(token.field),
				}
				array = append(array, new_tok)

				new_tok.children = parser.parse_block(BRACE_CLOSE)
				new_tok.position = token.position
				continue
			}

			if x.ast_type == EQUALS {
				parser.next()
				parser.eat_whitespace()
				parser.next()
				parser.eat_whitespace()

				the_token := &ast_declare{
					ast_type: DECL,
					field:    new_hash(token.field),
				}
				the_token.position = token.position

				x = parser.peek()

				if x.ast_type.is(WORD, IDENT) {
					parser.next()
					parser.eat_whitespace()

					if parser.peek().ast_type == BRACE_OPEN {
						parser.next()
						new_block := &ast_block{
							decl_hash: new_hash(x.field),
						}
						new_block.children = parser.parse_block(BRACE_CLOSE)
						the_token.children = []ast_data{new_block}
					} else {
						parser.step_back()
						parser.step_back()
						the_token.children = parser.parse_paragraph(NULL)
					}
				}

				array = append(array, the_token)
				continue
			}
			// if not, we're a a normal line,
			// so we just continue past the switch

		case BRACE_OPEN:
			if p := parser.peek(); p.ast_type.is(WHITESPACE, NEWLINE) {
				new_tok := &ast_block{}
				array = append(array, new_tok)

				new_tok.children = parser.parse_block(BRACE_CLOSE)
				new_tok.position = token.position
				continue
			}
			fallthrough // try for {#} declaration

		case BRACKET_OPEN:
			x := parser.next()

			is_brace := token.ast_type == BRACE_OPEN

			the_token := &ast_declare{}
			the_token.position = token.position

			{
				if x.ast_type > is_non_word {
					the_token.ast_type = DECL_TOKEN
				} else if !is_brace && x.ast_type.is(WORD, IDENT) {
					the_token.ast_type = DECL_BLOCK
				} else {
					if is_brace {
						panic("bad type in {decl}") // @error
					}
					panic("bad type in [decl]")
				}

				the_token.field = new_hash(x.field)

				// if brace instead of square bracket
				if is_brace {
					the_token.field += 1
				}
			}

			x = parser.next()

			if !x.ast_type.is(BRACKET_CLOSE, BRACE_CLOSE) {
				panic("bad closure on declaration")
			}

			parser.eat_whitespace()

			x = parser.next()

			if x.ast_type != EQUALS {
				panic("bad declaration")
			}

			parser.eat_whitespace()

			x = parser.peek()

			if x.ast_type.is(WORD, IDENT) {
				parser.next()
				parser.eat_whitespace()

				if parser.peek().ast_type == BRACE_OPEN {
					parser.next()
					new_block := &ast_block{
						decl_hash: new_hash(x.field),
					}
					new_block.children = parser.parse_block(BRACE_CLOSE)
					the_token.children = []ast_data{new_block}
				} else {
					parser.step_back()
					parser.step_back()
					the_token.children = parser.parse_paragraph(NULL)
				}
			} else if x.ast_type == BRACE_OPEN {
				parser.next()
				the_token.children = parser.parse_block(BRACE_CLOSE)
			} else {
				the_token.children = parser.parse_paragraph(NULL)
			}

			array = append(array, the_token)
			continue
		}

		{
			// parse_p needs to get the first tok again
			parser.step_back()
			tok := ast_normal{
				ast_type: NORMAL,
			}
			tok.children = parser.parse_paragraph(NULL)
			array = append(array, &tok)
		}
	}

	return array
}

func (parser *parser) parse_paragraph(exit_upon ast_type) []ast_data {
	array := make([]ast_data, 0, 32)
	break_all := false

	for {
		buffer := strings.Builder{}
		buffer.Grow(256)

		for {
			token := parser.next()

			if token == nil {
				return array
			}
			if token.ast_type.is(NEWLINE, EOF, exit_upon) {
				break_all = true
				break
			}

			switch token.ast_type {
			default:
				buffer.WriteString(token.field)

			case WHITESPACE:
				buffer.WriteRune(' ')

			case ASTERISK:
				open := parser.prev().ast_type.is(WHITESPACE, NEWLINE)

				x := parser.peek()

				if open && x.ast_type.is(WHITESPACE, NEWLINE, EOF) {
					buffer.WriteString(strings.Repeat("*", len(token.field)))
					continue
				}

				the_type := is_formatter

				switch len(token.field) {
				case 1: the_type = ITALIC_OPEN
				case 2: the_type = BOLD_OPEN
				case 3: the_type = BOLD_ITALIC_OPEN
				}

				if !open {
					the_type += 1
				}

				if buffer.Len() > 0 {
					array = append(array, &ast_normal {
						ast_type: NORMAL,
						field:    buffer.String(),
					})
					buffer.Reset()
					buffer.Grow(256)
				}

				array = append(array, &ast_normal {
					ast_type: the_type,
				})
				continue

			case PERCENT:
				if parser.peek().ast_type == BRACE_OPEN {
					parser.next()

					new_finder := &ast_finder{}

					word := parser.peek()

					if word.ast_type == WORD {
						parser.next()
						switch strings.ToLower(word.field) {
						case "page":  new_finder.finder_type = PAGE
						case "image": new_finder.finder_type = IMAGE
						default: // @error
						}

						// @todo clean up
						if parser.peek().ast_type == COLON {
							parser.next()
							word := parser.next()

							new_finder.path_type = check_path_type(strings.ToLower(word.field))
						} else {
							new_finder.path_type = NO_PATH_TYPE
						}
					}

					{
						parser.eat_whitespace()
						new_finder.children = parser.parse_paragraph(BRACE_CLOSE)

						// @todo need to deal with additional arguments here
					}

					array = append(array, new_finder)
					continue
				}

				new_var := &ast_variable{}

				a := parser.peek()

				the_type := VAR

				if a.ast_type.is(WORD, IDENT) {
					parser.next()

					b := parser.peek()

					if b.ast_type == STOP {
						parser.next()

						c := parser.peek()

						if c.ast_type.is(WORD, IDENT) {
							parser.next()

							new_var.field    = new_hash(a.field + "." + c.field)
							new_var.taxonomy = new_hash(a.field)
							new_var.subname  = new_hash(c.field)
						} else {
							new_var.field = new_hash(a.field)
						}
					} else {
						new_var.field = new_hash(a.field)
					}

				} else if a.ast_type == NUMBER {
					parser.next()
					the_type = VAR_ENUM

					n, err := strconv.ParseInt(a.field, 10, 32)
					if err != nil {
						panic(err)
					}

					new_var.field   = base_hash
					new_var.subname = uint32(n)

				} else if a.ast_type == PERCENT {
					parser.next()
					the_type      = VAR_ANON
					new_var.field = base_hash // just a %
				} else {
					buffer.WriteRune('%')
					continue
				}

				// apply the type
				new_var.ast_type = the_type

				if buffer.Len() > 0 {
					array = append(array, &ast_normal {
						ast_type: NORMAL,
						field:    buffer.String(),
					})
					buffer.Reset()
					buffer.Grow(256)
				}

				{
					a := parser.peek()

					if a.ast_type == COLON {
						parser.next()

						b := parser.peek()

						if b.ast_type == WORD {
							parser.next()

							switch strings.ToUpper(b.field) {
							case "SLUG":       new_var.modifier = SLUG
 							case "UPPER":      new_var.modifier = UPPER
 							case "LOWER":      new_var.modifier = LOWER
 							case "TITLE":      new_var.modifier = TITLE
 							case "EXPAND":     new_var.modifier = EXPAND
 							case "EXPAND_ALL": new_var.modifier = EXPAND_ALL
							}
						}
					}
				}

				array = append(array, new_var)
			}
		}

		if buffer.Len() > 0 {
			array = append(array, &ast_normal {
				ast_type: NORMAL,
				field:    buffer.String(),
			})
		}

		if break_all {
			break
		}
	}

	// rebalancer
	/*{
		for _, entry := range array {

		}
	}*/

	// @todo if any scope contains var_enum, replace all var_anon with var_enum

	return array
}

func (parser *parser) eat_comment() {
	parser.eat_whitespace()

	index := 0
	is_escaped := false
	passed_newline := false

	for i, entry := range parser.array[parser.index:] {
		if entry.ast_type == ESCAPE {
			is_escaped = true
			continue
		}

		switch entry.ast_type {
		case NEWLINE:     passed_newline = true
		case BRACE_OPEN:
			if !is_escaped {
				index++
			}
		case BRACE_CLOSE:
			if !is_escaped {
				index--
			}
			if passed_newline {
				passed_newline = false
			}
		}

		if passed_newline && index == 0 {
			parser.index += i
			return
		}

		is_escaped = false
	}
}

func (parser *parser) step_back() {
	parser.index--
}

func (parser *parser) prev() *lexer_token {
	if parser.index < 2 {
		return nil
	}
	return parser.array[parser.index - 2]
}

func (parser *parser) next() *lexer_token {
	if parser.index > len(parser.array) - 1 {
		return nil
	}
	t := parser.array[parser.index]
	parser.index++
	return t
}

func (parser *parser) peek() *lexer_token {
	if parser.index > len(parser.array) - 1 {
		return nil
	}
	return parser.array[parser.index]
}

func (parser *parser) peek_whitespace() *lexer_token {
	index := 0

	for _, token := range parser.array[parser.index:] {
		if token.ast_type == WHITESPACE {
			index++
			continue
		}
		break
	}

	return parser.array[parser.index + index]
}

func (parser *parser) eat_whitespace() {
	for _, token := range parser.array[parser.index:] {
		if token.ast_type == WHITESPACE {
			parser.index++
			continue
		}
		break
	}
}