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

	CONTROL_FOR
	CONTROL_IF
	OP_AND
	OP_OR
	OP_NOT

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

	WHITESPACE

	is_lexer
	NEWLINE
	WORD
	IDENT
	BRACE_OPEN
	BRACE_CLOSE
	BRACKET_OPEN
	BRACKET_CLOSE
	ANGLE_OPEN
	ANGLE_CLOSE

	ESCAPE
	NUMBER
	FORWARD_SLASH
	MULTIPLY
	AMPERSAND
	COLON
	TILDE
	EOF

	is_non_word
	NON_WORD
	ASTERISK
	EQUALS
	BANG
	STOP
	PLUS
	PERCENT
	PIPE
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
	get_position() *position
}
type ast_base_fields struct {
	position position
	children []ast_data
}

// normal token, which is the real "base object"
type ast_base struct {
	ast_base_fields
	ast_type   ast_type
	field      string
}
func (t *ast_base) type_check() ast_type {
	return t.ast_type
}
func (t *ast_base) get_children() []ast_data {
	return t.children
}
func (t *ast_base) get_position() *position {
	return &t.position
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
func (t *ast_variable) get_position() *position {
	return &t.position
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
func (t *ast_declare) get_position() *position {
	return &t.position
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
func (t *ast_block) get_position() *position {
	return &t.position
}

type ast_token struct {
	ast_base_fields
	decl_hash  uint32
	orig_field string
}
func (t *ast_token) type_check() ast_type {
	return TOKEN
}
func (t *ast_token) get_children() []ast_data {
	return t.children
}
func (t *ast_token) get_position() *position {
	return &t.position
}

type ast_finder struct {
	ast_base_fields
	finder_type    finder_type
	path_type      path_type
	image_settings *image_settings
}
func (t *ast_finder) type_check() ast_type {
	return RES_FINDER
}
func (t *ast_finder) get_children() []ast_data {
	return t.children
}
func (t *ast_finder) get_position() *position {
	return &t.position
}

type ast_for struct {
	ast_base_fields
	iterator_source ast_data
}
func (t *ast_for) type_check() ast_type {
	return CONTROL_FOR
}
func (t *ast_for) get_children() []ast_data {
	return t.children
}
func (t *ast_for) get_position() *position {
	return &t.position
}

type ast_if struct {
	ast_base_fields
	condition_list []ast_data
}
func (t *ast_if) type_check() ast_type {
	return CONTROL_IF
}
func (t *ast_if) get_children() []ast_data {
	return t.children
}
func (t *ast_if) get_position() *position {
	return &t.position
}


type ast_builtin struct {
	ast_base_fields
	ast_type  ast_type
	hash_name uint32
	// target    string
}
func (t *ast_builtin) type_check() ast_type {
	return t.ast_type
}
func (t *ast_builtin) get_children() []ast_data {
	return t.children
}
func (t *ast_builtin) get_position() *position {
	return &t.position
}


type finder_type uint8
const (
	_NO_FINDER finder_type = iota
	_PAGE
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