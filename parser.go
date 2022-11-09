package main

import (
	"strings"
	"strconv"
	"unicode/utf8"
)

type parser struct {
	index  int
	unwind bool
	errors *error_handler
	array  []*lexer_token
}

func parse_stream(errors *error_handler, array []*lexer_token, allow_anon bool) []ast_data {
	parser := parser {
		array:  array,
		errors: errors,
	}
	return parser.parse_block(0, allow_anon)
}

func (parser *parser) get_non_word(token *lexer_token) (string, int) {
	if token.ast_type == NON_WORD {
		return token.field, 1
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
	return strings.Repeat(token.field, count + 1), count
}

func (parser *parser) parse_block(max_depth int, allow_anon bool) []ast_data {
	if parser.unwind {
		return []ast_data{}
	}

	array := make([]ast_data, 0, 32)

	main_loop: for {
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
				t.position = token.position
				array = append(array, t)
			}
			continue
		}

		if token.ast_type > is_non_word {
			word, count := parser.get_non_word(token)

			if p := parser.peek(); p != nil && p.ast_type.is(WHITESPACE, NEWLINE) {
				new_tok := &ast_token{
					decl_hash:  new_hash(word),
					orig_field: word,
				}

				parser.eat_whitespace()
				new_tok.children = parser.parse_paragraph(allow_anon, NULL)
				new_tok.position = token.position

				array = append(array, new_tok)
				continue
			} else if count > 1 {
				parser.step_backn(count)
			}
		}

		switch token.ast_type {
		case WHITESPACE:
			whitespace := &ast_base{
				ast_type: WHITESPACE,
				field:    token.field,
			}
			whitespace.position = token.position
			array = append(array, whitespace)
			continue

		case FORWARD_SLASH:
			if p := parser.peek(); p != nil && p.ast_type != WHITESPACE {
				break // not a token, has to be spaced
			}
			parser.eat_comment()
			continue

		case TILDE, MULTIPLY, AMPERSAND, ANGLE_CLOSE:
			the_builtin := &ast_builtin{}
			the_type    := NULL

			switch token.ast_type {
			case TILDE:
				the_type = IMPORT
			case MULTIPLY:
				the_type = SCOPE_UNSET
			case AMPERSAND:
				the_type = TEMPLATE
			case ANGLE_CLOSE:
				the_type = PARTIAL
			}

			the_builtin.ast_type = the_type

			parser.eat_whitespace()

			if the_type == IMPORT {
				x := parser.parse_paragraph(allow_anon, WHITESPACE)

				if len(x) == 0 {
					panic("no children")
				}

				the_builtin.children = x
				parser.eat_whitespace()
			}

			peeked := parser.peek()

			if peeked.ast_type.is(WORD, IDENT) {
				parser.next()
				the_builtin.hash_name = new_hash(peeked.field)
			} else {
				parser.errors.new_pos(PARSER_WARNING, token.position, "ambiguous token %q should be escaped", token.field)
				break
			}

			parser.eat_whitespace()

			if !parser.peek().ast_type.is(NEWLINE, EOF) {
				if the_type == IMPORT {
					parser.errors.new_pos(PARSER_FAILURE, token.position, "malformed import (or unescaped ~ at start of line)")
					parser.unwind = true
				} else {
					parser.step_backn(3)
					parser.errors.new_pos(PARSER_WARNING, token.position, "ambiguous token %q should be escaped", token.field)
				}
				break
			}

			the_builtin.position = token.position
			array = append(array, the_builtin)
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

				the_if.children = parser.parse_block(1, allow_anon)
				the_if.position = token.position


				the_if.position.end += 1 // the closing brace

				array = append(array, the_if)
				continue
			}

			if token.field == "for" {
				the_for := &ast_for{}
				parser.eat_whitespace()

				x := parser.parse_paragraph(allow_anon, WHITESPACE)

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

				the_for.children = parser.parse_block(1, allow_anon)
				the_for.position = token.position


				the_for.position.end += 1 // the closing brace

				array = append(array, the_for)
				continue
			}

			parser.step_back() // revert for first part of variable (if any)
			field, taxonomy, subname := parser.parse_variable_ident()

			x := parser.peek_whitespace()

			needs_raw := false

			if x.ast_type == WORD && x.field == "raw" {
				parser.next()
				parser.eat_whitespace()
				parser.next()

				if n := parser.peek_whitespace(); n.ast_type == BRACE_OPEN {
					needs_raw = true
				} else {
					parser.step_backn(2)
				}

				x = parser.peek_whitespace()
			}

			if x.ast_type == BRACE_OPEN {
				parser.next()
				parser.eat_whitespace()
				parser.next()

				the_block := &ast_block {
					decl_hash: new_hash(token.field),
				}

				if needs_raw {
					x := parser.parse_raw_block()

					the_block.children = []ast_data{x}

					// @todo update positions, etc.
					array = append(array, the_block)
					continue
				}

				the_block.children = parser.parse_block(0, allow_anon)
				the_block.position = token.position


				the_block.position.end += 1 // the closing brace

				array = append(array, the_block)
				continue

			} else if x.ast_type == EQUALS {
				parser.next()
				parser.eat_whitespace()
				parser.next()
				parser.eat_whitespace()

				the_decl := &ast_declare{
					ast_type: DECL,
					field:    field,
					taxonomy: taxonomy,
					subname:  subname,
				}
				the_decl.position = token.position

				x = parser.peek()

				if x.ast_type.is(WORD, IDENT) {
					parser.next()
					did_eat := parser.eat_whitespace()

					if parser.peek().ast_type == BRACE_OPEN {
						parser.next()
						new_block := &ast_block{
							decl_hash: new_hash(x.field),
						}
						new_block.children = parser.parse_block(0, allow_anon)
						new_block.position = x.position

						new_block.position.end += 1 // the closing brace

						the_decl.children = []ast_data{new_block}
						the_decl.position = new_block.position

					} else {
						if did_eat {
							parser.step_backn(2)
						} else {
							parser.step_back()
						}

						parser.eat_whitespace()
						the_decl.children = parser.parse_paragraph(allow_anon, NULL)
					}
				} else {
					the_decl.children = parser.parse_paragraph(allow_anon, NULL)
				}

				array = append(array, the_decl)
				continue
			}

			// if not, we're a a normal line,
			// so we just continue past the switch
			if taxonomy > 0 {
				parser.step_backn(2)
			}

		case BRACE_OPEN:
			if p := parser.peek(); p.ast_type.is(WHITESPACE, NEWLINE) {
				the_block := &ast_block{}

				the_block.children = parser.parse_block(0, allow_anon)
				the_block.position = token.position


				the_block.position.end += 1 // the closing brace

				array = append(array, the_block)
				continue
			}
			fallthrough // try for {#} declaration

		case BRACKET_OPEN:
			inner_text := parser.next()

			is_brace := token.ast_type == BRACE_OPEN

			the_decl := &ast_declare{}
			the_decl.position = token.position

			isnt_valid := false

			{
				if inner_text.ast_type > is_non_word {
					the_decl.ast_type = DECL_TOKEN
					word, _ := parser.get_non_word(inner_text)
					the_decl.field = new_hash(word)

				} else if !is_brace && inner_text.ast_type.is(WORD, IDENT) {
					the_decl.ast_type = DECL_BLOCK

					parser.step_back()
					field, taxonomy, subname := parser.parse_variable_ident()

					the_decl.field    = field
					the_decl.taxonomy = taxonomy
					the_decl.subname  = subname

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

			{
				did_any := parser.eat_whitespace()
				equals  := parser.next()

				if equals.ast_type != EQUALS {
					n := 3
					if did_any {
						n = 4
					}
					parser.step_backn(n)
					break
				}
			}

			// now that we're certain the user intended a declaration, we're killing it
			if isnt_valid {
				if is_brace {
					parser.errors.new_pos(PARSER_FAILURE, the_decl.position, "bad type in {declaration}: %q cannot be used as a token character", inner_text.field)
				} else {
					parser.errors.new_pos(PARSER_FAILURE, the_decl.position, "bad type in [declaration]: %q cannot be used as a block template name", inner_text.field)
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
					new_block.children = parser.parse_block(0, true)
					new_block.position = x.position


					new_block.position.end += 1 // the closing brace

					the_decl.children = []ast_data{new_block}
					the_decl.position.start = new_block.position.start
				} else {
					parser.step_backn(2)
					the_decl.children = parser.parse_paragraph(true, NULL)
				}
			} else if x.ast_type == BRACE_OPEN {
				parser.next()
				the_decl.children = parser.parse_block(0, true)
			} else {
				the_decl.children = parser.parse_paragraph(true, NULL)
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
			the_para.position = token.position
			the_para.children = parser.parse_paragraph(allow_anon, NULL)


			// @todo sanitise any escaped special characters that fall down here

			if token.ast_type == ANGLE_OPEN {
				the_para.ast_type = RAW
			}

			array = append(array, &the_para)
		}
	}

	return array
}

func (parser *parser) parse_paragraph(allow_anon bool, exit_upon ...ast_type) []ast_data {
	if parser.unwind {
		return []ast_data{}
	}

	array := make([]ast_data, 0, 32)

	const alloc = 256

	buffer := strings.Builder{}
	buffer.Grow(alloc)

	for {
		token := parser.next()

		if token == nil {
			return array
		}
		if token.ast_type.is(NEWLINE, EOF) {
			break
		}
		if token.ast_type.is(exit_upon...) {
			parser.step_back()
			break
		}

		switch token.ast_type {
		default:
			buffer.WriteString(token.field)
			/*n := &ast_base{
				ast_type: NORMAL,
				field:    token.field,
			}
			n.position = token.position
			array = append(array, n)*/

		case WHITESPACE:
			buffer.WriteRune(' ')
			/*n := &ast_base{
				ast_type: WHITESPACE,
				field:    token.field,
			}
			n.position = token.position
			array = append(array, n)*/

		/*case ASTERISK:
			open := parser.prev().ast_type.is(WHITESPACE, NEWLINE)

			x := parser.peek()

			if open && x.ast_type.is(WHITESPACE, NEWLINE, EOF) {
				n := &ast_base{
					ast_type: WHITESPACE,
					field:    token.field,
				}
				n.position = token.position
				array = append(array, n)
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

			the_formatter := &ast_base{
				ast_type: the_type,
			}
			the_formatter.position = token.position

			array = append(array, the_formatter)
			continue*/

		case PERCENT:
			if parser.peek().ast_type == BRACE_OPEN {
				parser.next()

				the_finder := &ast_finder{}
				the_finder.position = token.position

				word := parser.peek()

				if word.ast_type == WORD {
					parser.next()
					switch strings.ToLower(word.field) {
					case "page":   the_finder.finder_type = _PAGE
					case "image":  the_finder.finder_type = _IMAGE
					case "static": the_finder.finder_type = _STATIC
					}

					if the_finder.finder_type == _NO_FINDER {
						parser.step_backn(2)
						n := &ast_base{
							ast_type: NORMAL,
							field:    token.field,
						}
						n.position = token.position
						array = append(array, n)
						continue
					}

					if parser.peek().ast_type == COLON {
						parser.next()
						word := parser.next()
						the_finder.path_type = check_path_type(strings.ToLower(word.field)) // @todo revert colon if not word
					} else {
						the_finder.path_type = NO_PATH_TYPE
					}
				}

				{
					parser.eat_whitespace()
					the_finder.children = parser.parse_paragraph(allow_anon, WHITESPACE, BRACE_CLOSE)
					// @error error if more than one child

					parser.eat_whitespace()

					if parser.peek().ast_type == BRACE_CLOSE {
						parser.next()
					} else {
						if the_finder.finder_type == _IMAGE {
							the_finder.image_settings = parser.parse_image_settings()

							if parser.unwind {
								return array
							}
						}
						if parser.peek().ast_type == BRACE_CLOSE {
							parser.next()
						}
					}
				}

				if buffer.Len() > 0 {
					n := &ast_base{
						ast_type: NORMAL,
						field:    buffer.String(),
					}
					array = append(array, n)
					buffer.Reset()
					buffer.Grow(alloc)
				}

				array = append(array, the_finder)
				continue
			}

			new_var := parser.parse_variable(allow_anon)

			if new_var == nil {
				buffer.WriteString(token.field)
				continue
			}

			if buffer.Len() > 0 {
				n := &ast_base{
					ast_type: NORMAL,
					field:    buffer.String(),
				}
				array = append(array, n)
				buffer.Reset()
				buffer.Grow(alloc)
			}

			new_var.position = token.position
			array = append(array, new_var)
		}
	}

	if buffer.Len() > 0 {
		n := &ast_base{
			ast_type: NORMAL,
			field:    buffer.String(),
		}
		array = append(array, n)
		buffer.Reset()
	}

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
			new_var := parser.parse_variable(false)

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

func (parser *parser) parse_variable_ident() (uint32, uint32, uint32) {
	a := parser.next()
	b := parser.peek()

	if b.ast_type == STOP {
		parser.next()

		c := parser.peek()

		if c.ast_type.is(WORD, IDENT) {
			parser.next()
			return new_hash(a.field + c.field), new_hash(a.field), new_hash(c.field)
		}

		parser.step_back()
		return new_hash(a.field), 0, 0
	}

	return new_hash(a.field), 0, 0
}

func (parser *parser) parse_variable(allow_anon bool) *ast_variable {
	new_var := &ast_variable{}

	a := parser.peek()

	the_type := VAR

	if a.ast_type.is(WORD, IDENT) {
		field, taxonomy, subname := parser.parse_variable_ident()

		new_var.field    = field
		new_var.taxonomy = taxonomy
		new_var.subname  = subname

	} else if allow_anon && a.ast_type == NUMBER {
		parser.next()
		the_type = VAR_ENUM

		n, err := strconv.ParseInt(a.field, 10, 32)
		if err != nil {
			panic(err)
		}

		new_var.field   = base_hash
		new_var.subname = uint32(n)

	} else if allow_anon && a.ast_type == PERCENT {
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
				/*case "raw", "r":
					new_var.modifier = RAW_SUB*/
				/*case "expand", "e":
					new_var.modifier = EXPAND
				case "expand_all", "ea":
					new_var.modifier = EXPAND_ALL*/
				default:
					parser.errors.new_pos(PARSER_WARNING, b.position, "unknown variable modifier %q", b.field)
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

func (parser *parser) parse_raw_block() ast_data {
	buffer := strings.Builder{}
	buffer.Grow(512)

	brace_balance := 1
	is_escaped := false

	main_loop: for i, token := range parser.array[parser.index:] {
		if token.ast_type == ESCAPE {
			is_escaped = true
			continue
		}

		switch token.ast_type {
		case ANGLE_OPEN:
			buffer.WriteString("&lt;")
			continue

		case ANGLE_CLOSE:
			buffer.WriteString("&gt;")
			continue

		case AMPERSAND:
			if len(parser.array) - 1 > parser.index + i + 1 {
				if parser.array[parser.index + i + 1].ast_type != WORD {
					buffer.WriteString("&amp;")
					continue
				}
			}

		// @todo add double + single quotes

		case BRACE_OPEN:
			if !is_escaped {
				brace_balance++
			}
		case BRACE_CLOSE:
			if !is_escaped {
				brace_balance--
			}
		case EOF:
			parser.index += i + 1
			break main_loop
		}

		if brace_balance <= 0 {
			parser.index += i + 1
			break main_loop
		}

		buffer.WriteString(token.field)

		is_escaped = false
	}

	str := reindent_text(buffer.String())

	return &ast_base{
		ast_type: RAW,
		field:    str,
	}
}

func (parser *parser) parse_image_settings() *image_settings {
	settings := &image_settings{}

	got_anything := false

	main_loop: for {
		parser.eat_whitespace()
		token := parser.next()

		switch token.ast_type {
		case NUMBER:
			a, err := strconv.ParseInt(token.field, 10, 64)
			if err != nil {
				panic(err) // somehow a number isn't a number, this is horrendously bad (programmer error)
			}

			if parser.peek().ast_type == WORD {
				settings.width = uint(a)
			} else {
				settings.quality = int(a)
			}

			got_anything = true

		case WORD:
			field := token.field

			rune, _ := utf8.DecodeRuneInString(field)

			if rune == 'x' {
				field = field[1:]

				b, err := strconv.ParseInt(field, 10, 64)
				if err != nil {
					panic(err) // @error (this is user error)
				}

				settings.height = uint(b)
				got_anything = true
				continue
			}

			switch field {
			case "png":
				settings.file_type = IMG_PNG
				got_anything = true
			case "webp":
				settings.file_type = IMG_WEB
				got_anything = true
			case "jpeg", "jpg":
				settings.file_type = IMG_JPG
				got_anything = true
			default:
				parser.errors.new_pos(PARSER_FAILURE, token.position, "image format %q is unsupported", field)
				parser.unwind = true
				break main_loop
			}

		default:
			parser.step_back()
			break main_loop
		}
	}

	if got_anything {
		if settings.quality == 0 {
			settings.quality = 100
		}
		return settings
	}

	return nil
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

func (parser *parser) eat_whitespace() bool {
	did_any := false

	for _, token := range parser.array[parser.index:] {
		if token.ast_type == WHITESPACE {
			did_any = true
			parser.index++
			continue
		}
		break
	}

	return did_any
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

func arrange_top_scope(content []ast_data) []ast_data {
	array := make([]ast_data, 0, 8)

	for _, entry := range content {
		if entry.type_check().is(DECL, DECL_TOKEN, DECL_BLOCK, TEMPLATE) {
			array = append(array, entry)
		}
	}

	return array
}

func reindent_text(input string) string {
	input = strings.ReplaceAll(input, "\t", "    ")
	lines := strings.Split(input, "\n")

	shortest_indent := len(input)

	for i, c := range input {
		if c != '\n' {
			input = input[i:]
			break
		}
	}

	for _, line := range lines {
		count := 0

		if len(line) == 0 {
			continue
		}

		for _, c := range line {
			if c != ' ' {
				break
			}
			count ++
		}

		if count < shortest_indent {
			shortest_indent = count
		}
	}

	if shortest_indent == 0 {
		return input
	}

	buffer := strings.Builder{}
	buffer.Grow(len(input))

	for _, line := range lines {
		if len(line) == 0 {
			buffer.WriteRune('\n')
			continue
		}
		buffer.WriteString(line[shortest_indent:])
		buffer.WriteRune('\n')
	}

	render := buffer.String()

	for i, c := range render {
		if c != '\n' {
			render = render[i:]
			break
		}
	}

	// not utf8 aware, but should be fine
	// because we're only interested in a
	// single-width char
	for i := len(render) - 1; i >= 0; i-- {
		c := render[i]
		if c != '\n' {
			render = render[:i + 1]
			break
		}
	}

	return render
}

// @todo extremely primitive
func is_ext_url(input string) bool {
	for {
		if len(input) == 0 {
			break
		}
		r, width := utf8.DecodeRuneInString(input)
		input = input[width:]

		if r == ':' {
			a, w := utf8.DecodeRuneInString(input)
			input = input[w:]
			b, _ := utf8.DecodeRuneInString(input)

			if a == b && a == '/' {
				return true
			}
		}
	}
	return false
}