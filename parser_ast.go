package main

type ast_type uint8
const (
	NULL ast_type = iota

	BLANK
	RAW
	NORMAL
	TOKEN

	IMPORT
	PARTIAL
	TEMPLATE
	SCOPE_UNSET

	DECL
	DECL_BLOCK
	DECL_TOKEN

	VAR
	VAR_ENUM
	VAR_ANON

	RES_FINDER
	BLOCK

	is_formatter
	ITALIC_OPEN
	ITALIC_CLOSE
	BOLD_OPEN
	BOLD_CLOSE
	BOLD_ITALIC_OPEN
	BOLD_ITALIC_CLOSE
	MARK_OPEN
	MARK_CLOSE
	STRIKE_OPEN
	STRIKE_CLOSE

	is_lexer
	WHITESPACE
	NEWLINE
	WORD
	IDENT
	EQUALS
	BRACE_OPEN
	BRACE_CLOSE
	BRACKET_OPEN
	BRACKET_CLOSE
	ANGLE_OPEN
	ANGLE_CLOSE

	is_non_word
	NON_WORD // see below
	ESCAPE
	FORWARD_SLASH
	NUMBER
	STOP
	COLON
	ASTERISK
	AMPERSAND
	TILDE
	PLUS
	MULTIPLY
	PERCENT

	EOF
)

/*
	@note
	a NON_WORD is one or more of the same
	non-alphanumeric rune, like ## or &&&
*/

type ast_modifier uint8
const (
	NONE ast_modifier = iota
	SLUG
	UNIQUE_SLUG
	UPPER
	LOWER
	TITLE
	EXPAND
	EXPAND_ALL
)

func (t ast_type) is(comp ...ast_type) bool {
	for _, c := range comp {
		if t == c {
			return true
		}
	}
	return false
}

// base token data structure
type ast_data interface {
	type_check() ast_type
	get_children() []ast_data
}
type ast_base_fields struct {
	position uint16
	children []ast_data
}

// normal token, which is the real "base object"
type ast_normal struct {
	ast_base_fields
	ast_type   ast_type
	field      string
}
func (t *ast_normal) type_check() ast_type {
	return t.ast_type
}
func (t *ast_normal) get_children() []ast_data {
	return t.children
}

type ast_variable struct {
	ast_base_fields
	ast_type ast_type
	modifier ast_modifier
	field    uint32
	taxonomy uint32
	subname  uint32
}
func (t *ast_variable) type_check() ast_type {
	return t.ast_type
}
func (t *ast_variable) get_children() []ast_data {
	return t.children
}

type ast_declare struct {
	ast_base_fields
	ast_type   ast_type
	field      uint32
	taxonomy   uint32
	subname    uint32
}
func (t *ast_declare) type_check() ast_type {
	return t.ast_type
}
func (t *ast_declare) get_children() []ast_data {
	return t.children
}

type ast_block struct {
	ast_base_fields
	decl_hash uint32 // zero means anonymous
}
func (t *ast_block) type_check() ast_type {
	return BLOCK
}
func (t *ast_block) get_children() []ast_data {
	return t.children
}

type ast_token struct {
	ast_base_fields
	decl_hash uint32
}
func (t *ast_token) type_check() ast_type {
	return TOKEN
}
func (t *ast_token) get_children() []ast_data {
	return t.children
}

type finder_type uint8
const (
	_PAGE finder_type = iota
	_IMAGE
	_STATIC
)

type path_type uint8
const (
	NO_PATH_TYPE path_type = iota
	RELATIVE
	ABSOLUTE
	ROOTED // @todo bad name
)

func check_path_type(input string) path_type {
	switch input {
	case "abs", "absolute":
		return ABSOLUTE
	case "rel", "relative":
		return RELATIVE
	case "root", "rooted":
		return ROOTED
	}
	return NO_PATH_TYPE
}

type ast_finder struct {
	ast_base_fields
	finder_type finder_type
	path_type   path_type
}
func (t *ast_finder) type_check() ast_type {
	return RES_FINDER
}
func (t *ast_finder) get_children() []ast_data {
	return t.children
}