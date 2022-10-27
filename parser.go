package main

import "strings"
import "strconv"

type parser struct {
	index  int
	unwind bool
	errors *error_handler
	array  []*lexer_token
	location string
}

func parse_stream(who_am_i string, errors *error_handler, array []*lexer_token) []ast_data {
	parser := parser {
		array:    array,
		errors:   errors,
		location: who_am_i,
	}
	return parser.parse_block(0)
}

func (parser *parser) get_non_word(token *lexer_token) string {
	if token.ast_type == NON_WORD {
		return token.field
	}

	count := 0

	for i, c := range parser.array[parser.index:] {
		if c.ast_type != token.ast_type {
			count = i
			break
		}
	}

	parser.index += count

	// count is + 1 because the first of the series has already
	// been docked by the parser, which is also why we don't add
	// + 1 to the index above
	return strings.Repeat(token.field, count + 1)
}

func (parser *parser) parse_block(max_depth int) []ast_data {
	array := make([]ast_data, 0, 32)

	if parser.unwind {
		return array
	}

	main_loop: for {
		parser.eat_whitespace()

		token := parser.next()

		if token == nil || token.ast_type.is(EOF, BRACE_CLOSE) {
			break
		}
		if max_depth > 0 && len(array) >= max_depth {
			break
		}

		if token.ast_type == NEWLINE {
			if p := parser.prev(); p != nil && p.ast_type == NEWLINE {
				t := &ast_base{
					ast_type: BLANK,
				}
				t.position = &position{token.position, parser.location}
				array = append(array, t)
			}
			continue
		}

		if token.ast_type > is_non_word {
			word := parser.get_non_word(token)

			new_tok := &ast_token {
				decl_hash:  new_hash(word),
				orig_field: word,
			}

			if p := parser.peek(); p != nil && p.ast_type != WHITESPACE {
				break // not a token, has to be spaced
			}

			parser.eat_whitespace()
			new_tok.children = parser.parse_paragraph(NULL)
			new_tok.position = &position{token.position, parser.location}

			array = append(array, new_tok)
			continue
		}

		switch token.ast_type {
		case FORWARD_SLASH:
			if p := parser.peek(); p != nil && p.ast_type != WHITESPACE {
				break // not a token, has to be spaced
			}
			parser.eat_comment()
			continue

		case AMPERSAND, ANGLE_CLOSE, TILDE, MULTIPLY:
			new_tok := &ast_base{}

			switch token.ast_type {
			case AMPERSAND:   new_tok.ast_type = TEMPLATE
			case ANGLE_CLOSE: new_tok.ast_type = PARTIAL // don't exist yet
			case TILDE:       new_tok.ast_type = IMPORT
			case MULTIPLY:    new_tok.ast_type = SCOPE_UNSET
			}

			parser.eat_whitespace()

			x := parser.peek()

			if x.ast_type.is(WORD, IDENT) {
				parser.next()
				new_tok.field = x.field
			} else {
				parser.errors.new_pos(PARSER_WARNING, position{token.position, parser.location}, "ambiguous token %q should be escaped", token.field)
				break
			}

			parser.eat_whitespace()

			if parser.peek().ast_type != NEWLINE {
				parser.step_backn(3)
				parser.errors.new_pos(PARSER_WARNING, position{token.position, parser.location}, "ambiguous token %q should be escaped", token.field)
				break
			}

			new_tok.position = &position{token.position, parser.location}
			array = append(array, new_tok)
			continue

		case WORD, IDENT:
			if token.field == "if" {
				the_if := &ast_if{}

				the_if.condition_list = parser.parse_if()

				if len(the_if.condition_list) == 0 {
					parser.step_back()
					break
				}

				parser.eat_whitespace()

				the_if.children = parser.parse_block(1)
				the_if.position = &position{token.position, parser.location}

				array = append(array, the_if)
				continue
			}

			if token.field == "for" {
				the_for := &ast_for{}
				parser.eat_whitespace()

				x := parser.parse_paragraph(WHITESPACE)

				if len(x) == 0 {
					parser.step_back()
					break
				}

				n := x[0]

				if !n.type_check().is(VAR, VAR_ANON, VAR_ENUM, RES_FINDER) {
					parser.step_backn(3)
					break
				}

				the_for.iterator_source = n

				parser.eat_whitespace()

				the_for.children = parser.parse_block(1)
				the_for.position = &position{token.position, parser.location}

				array = append(array, the_for)
				continue
			}

			x := parser.peek_whitespace()

			if x.ast_type == BRACE_OPEN {
				parser.next()
				parser.eat_whitespace()
				parser.next()

				the_block := &ast_block {
					decl_hash: new_hash(token.field),
				}

				the_block.children = parser.parse_block(0)
				the_block.position = &position{token.position, parser.location}

				array = append(array, the_block)
				continue
			}

			if x.ast_type == EQUALS {
				parser.next()
				parser.eat_whitespace()
				parser.next()
				parser.eat_whitespace()

				the_decl := &ast_declare{
					ast_type: DECL,
					field:    new_hash(token.field),
				}
				the_decl.position = &position{token.position, parser.location}

				x = parser.peek()

				if x.ast_type.is(WORD, IDENT) {
					parser.next()
					parser.eat_whitespace()

					if parser.peek().ast_type == BRACE_OPEN {
						parser.next()
						new_block := &ast_block{
							decl_hash: new_hash(x.field),
						}
						new_block.children = parser.parse_block(0)
						the_decl.children = []ast_data{new_block}
					} else {
						parser.step_back()
						the_decl.children = parser.parse_paragraph(NULL)
					}
				}

				array = append(array, the_decl)
				continue
			}
			// if not, we're a a normal line,
			// so we just continue past the switch

		case BRACE_OPEN:
			if p := parser.peek(); p.ast_type.is(WHITESPACE, NEWLINE) {
				the_block := &ast_block{}

				the_block.children = parser.parse_block(0)
				the_block.position = &position{token.position, parser.location}

				array = append(array, the_block)
				continue
			}
			fallthrough // try for {#} declaration

		case BRACKET_OPEN:
			inner_text := parser.next()

			is_brace := token.ast_type == BRACE_OPEN

			the_decl := &ast_declare{}
			the_decl.position = &position{token.position, parser.location}

			isnt_valid := false

			{
				if inner_text.ast_type > is_non_word {
					the_decl.ast_type = DECL_TOKEN
					the_decl.field = new_hash(parser.get_non_word(inner_text))

				} else if !is_brace && inner_text.ast_type.is(WORD, IDENT) {
					the_decl.ast_type = DECL_BLOCK
					the_decl.field = new_hash(inner_text.field)
				} else {
					isnt_valid = true
				}

				// if brace instead of square bracket
				if is_brace {
					the_decl.field += 1
				}
			}

			{
				bracket_close := parser.next()

				if !bracket_close.ast_type.is(BRACKET_CLOSE, BRACE_CLOSE) {
					parser.step_backn(2)
					break
				}
			}

			parser.eat_whitespace()

			{
				equals := parser.next()

				if equals.ast_type != EQUALS {
					parser.step_backn(4)
					break
				}
			}

			// now that we're certain the user intended a declaration, we're killing it
			if isnt_valid {
				if is_brace {
					parser.errors.new_pos(PARSER_FAILURE, *the_decl.position, "bad type in {declaration}: %q cannot be used as a token character", inner_text.field)
				} else {
					parser.errors.new_pos(PARSER_FAILURE, *the_decl.position, "bad type in [declaration]: %q cannot be used as a block template name", inner_text.field)
				}
				parser.unwind = true
				break main_loop
			}

			parser.eat_whitespace()

			x := parser.peek()

			if x.ast_type.is(WORD, IDENT) {
				parser.next()
				parser.eat_whitespace()

				if parser.peek().ast_type == BRACE_OPEN {
					parser.next()
					new_block := &ast_block{
						decl_hash: new_hash(x.field),
					}
					new_block.children = parser.parse_block(0)
					the_decl.children = []ast_data{new_block}
				} else {
					parser.step_backn(2)
					the_decl.children = parser.parse_paragraph(NULL)
				}
			} else if x.ast_type == BRACE_OPEN {
				parser.next()
				the_decl.children = parser.parse_block(0)
			} else {
				the_decl.children = parser.parse_paragraph(NULL)
			}

			array = append(array, the_decl)
			continue
		}

		{
			// parse_p needs to get the first tok again
			parser.step_back()
			the_para := ast_base{
				ast_type: NORMAL,
			}
			the_para.position = &position{token.position, parser.location}
			the_para.children = parser.parse_paragraph(NULL)

			// @todo sanitise any special characters that fall down here

			if token.ast_type == ANGLE_OPEN {
				the_para.ast_type = RAW

				// simplify if possible
				if len(the_para.children) == 1 && the_para.children[0].type_check() == NORMAL {
					the_para.field = the_para.children[0].(*ast_base).field
					the_para.children = nil
				}
			}

			array = append(array, &the_para)
		}
	}

	return array
}

func (parser *parser) parse_paragraph(exit_upon ast_type) []ast_data {
	array := make([]ast_data, 0, 32)
	break_all := false

	if parser.unwind {
		return array
	}

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
					array = append(array, &ast_base{
						ast_type: NORMAL,
						field:    buffer.String(),
					})
					buffer.Reset()
					buffer.Grow(256)
				}

				array = append(array, &ast_base{
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
						case "page":   new_finder.finder_type = _PAGE
						case "image":  new_finder.finder_type = _IMAGE
						case "static": new_finder.finder_type = _STATIC
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

					if buffer.Len() > 0 {
						array = append(array, &ast_base {
							ast_type: NORMAL,
							field:    buffer.String(),
						})
						buffer.Reset()
						buffer.Grow(256)
					}

					array = append(array, new_finder)
					continue
				}

				new_var := parser.parse_variable()

				if new_var == nil {
					buffer.WriteRune('%')
					continue
				}

				if buffer.Len() > 0 {
					array = append(array, &ast_base {
						ast_type: NORMAL,
						field:    buffer.String(),
					})
					buffer.Reset()
					buffer.Grow(256)
				}

				array = append(array, new_var)
			}
		}

		if buffer.Len() > 0 {
			array = append(array, &ast_base {
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

func (parser *parser) parse_if() []ast_data {
	array := make([]ast_data, 0, 8)

	for {
		parser.eat_whitespace()
		token := parser.next()

		switch token.ast_type {
		case BANG:
			array = append(array, &ast_base{
				ast_type: OP_NOT,
			})

		case PLUS:
			array = append(array, &ast_base{
				ast_type: OP_AND,
			})

		case PIPE:
			array = append(array, &ast_base{
				ast_type: OP_OR,
			})

		case PERCENT:
			new_var := parser.parse_variable()

			if new_var == nil {
				panic("bad thing in if")
			}
			if new_var.ast_type != VAR {
				panic("bad variable in if")
			}

			array = append(array, new_var)

		default:
			parser.step_back()
			return array
		}
	}

	return array
}

func (parser *parser) parse_variable() *ast_variable {
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
				parser.step_back()
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
		return nil
	}

	// apply the type
	new_var.ast_type = the_type

	{
		a := parser.peek()

		if a.ast_type == COLON {
			parser.next()

			b := parser.peek()

			if b.ast_type.is(WORD, IDENT) {
				parser.next()

				switch strings.ToLower(b.field) {
				case "slug", "s":
					new_var.modifier = SLUG
				case "unique_slug", "uslug", "us":
					new_var.modifier = UNIQUE_SLUG
					case "upper", "u":
						new_var.modifier = UPPER
					case "lower", "l":
						new_var.modifier = LOWER
					case "title", "t":
						new_var.modifier = TITLE
				// @todo
					/*case "expand", "e":
						new_var.modifier = EXPAND
					case "expand_all", "ea":
						new_var.modifier = EXPAND_ALL*/
				}
			} else {
				parser.step_back() // revert the colon
			}
		}
	}

	return new_var
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
		case NEWLINE:
			passed_newline = true

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
		case EOF:
			parser.index += i
			return
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

func (parser *parser) step_backn(n int) {
	parser.index -= n
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

func recursive_anon_count(children []ast_data) int {
	count := 0
	for _, entry := range children {
		if entry.type_check().is(DECL, DECL_TOKEN, DECL_BLOCK) {
			continue
		}
		if entry.type_check().is(VAR_ANON, VAR_ENUM) {
			count++
			continue
		}
		if x := entry.get_children(); len(x) > 0 {
			count += recursive_anon_count(x)
		}
	}
	return count
}

func immediate_decl_count(children []ast_data) int {
	count := 0
	for _, entry := range children {
		if entry.type_check().is(DECL, DECL_TOKEN, DECL_BLOCK) {
			count++
		}
		if entry.type_check() == TEMPLATE {
			count += 8
		}
	}
	return count
}