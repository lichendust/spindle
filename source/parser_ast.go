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

//go:generate stringer -type=AST_Type,Modifier,File_Type -output=parser_string.go

type AST_Type uint8
const (
	NULL AST_Type = iota

	BLANK
	RAW
	NORMAL
	TOKEN

	IMPORT
	PARTIAL
	TEMPLATE
	SCOPE_UNSET
	SCRIPT

	DECL
	DECL_BLOCK
	DECL_TOKEN
	DECL_REJECT // temp scope blocking

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
	DOLLAR

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

func (t AST_Type) is(comp ...AST_Type) bool {
	for _, c := range comp {
		if t == c {
			return true
		}
	}
	return false
}

// base token data structure
type AST_Data interface {
	type_check()   AST_Type
	get_children() []AST_Data
	get_position() Position
}
type AST_Base struct {
	position Position
	children []AST_Data
	ast_type AST_Type
	field    string
}
func (t *AST_Base) type_check() AST_Type {
	return t.ast_type
}
func (t *AST_Base) get_children() []AST_Data {
	return t.children
}
func (t *AST_Base) get_position() Position {
	return t.position
}



type AST_Variable struct {
	AST_Base
	ast_type AST_Type
	modifier Modifier
	field    uint32
	taxonomy uint32
	subname  uint32
}
func (t *AST_Variable) type_check() AST_Type {
	return t.ast_type
}
func (t *AST_Variable) get_children() []AST_Data {
	return t.children
}
func (t *AST_Variable) get_position() Position {
	return t.position
}



type AST_Declare struct {
	AST_Base
	ast_type   AST_Type
	field      uint32
	taxonomy   uint32
	subname    uint32
	immediate  bool
	is_soft    bool
}
func (t *AST_Declare) type_check() AST_Type {
	return t.ast_type
}
func (t *AST_Declare) get_children() []AST_Data {
	return t.children
}
func (t *AST_Declare) get_position() Position {
	return t.position
}



type AST_Block struct {
	AST_Base
	decl_hash uint32 // zero means anonymous
}
func (t *AST_Block) type_check() AST_Type {
	return BLOCK
}
func (t *AST_Block) get_children() []AST_Data {
	return t.children
}
func (t *AST_Block) get_position() Position {
	return t.position
}



type AST_Token struct {
	AST_Base
	decl_hash  uint32
	orig_field string
}
func (t *AST_Token) type_check() AST_Type {
	return TOKEN
}
func (t *AST_Token) get_children() []AST_Data {
	return t.children
}
func (t *AST_Token) get_position() Position {
	return t.position
}



type AST_Exec struct {
	AST_Base
	exec_type      Exec_Type
	modifier       Modifier
	path_type      Path_Type
	Image_Settings *Image_Settings
}
func (t *AST_Exec) type_check() AST_Type {
	return RES_FINDER
}
func (t *AST_Exec) get_children() []AST_Data {
	return t.children
}
func (t *AST_Exec) get_position() Position {
	return t.position
}



type AST_For struct {
	AST_Base
	iterator_source AST_Data
}
func (t *AST_For) type_check() AST_Type {
	return CONTROL_FOR
}
func (t *AST_For) get_children() []AST_Data {
	return t.children
}
func (t *AST_For) get_position() Position {
	return t.position
}



type AST_If struct {
	AST_Base
	is_else bool
	condition_list []AST_Data
}
func (t *AST_If) type_check() AST_Type {
	return CONTROL_IF
}
func (t *AST_If) get_children() []AST_Data {
	return t.children
}
func (t *AST_If) get_position() Position {
	return t.position
}



type AST_Builtin struct {
	AST_Base
	ast_type  AST_Type
	hash_name uint32
	// target    string
}
func (t *AST_Builtin) type_check() AST_Type {
	return t.ast_type
}
func (t *AST_Builtin) get_children() []AST_Data {
	return t.children
}
func (t *AST_Builtin) get_position() Position {
	return t.position
}



type AST_Script struct {
	AST_Base
	hash_name uint32
}
func (t *AST_Script) type_check() AST_Type {
	return SCRIPT
}
func (t *AST_Script) get_children() []AST_Data {
	return t.children
}
func (t *AST_Script) get_position() Position {
	return t.position
}



type Modifier uint8
const (
	NONE Modifier = iota
	SLUG
	UNIQUE_SLUG
	UPPER
	LOWER
	TITLE
	EXPAND
	EXPAND_ALL
)

type Path_Type uint8
const (
	NO_PATH_TYPE Path_Type = iota
	RELATIVE
	ABSOLUTE
	ROOTED // @todo bad name
)

type Exec_Type uint8
const (
	_NO_EXEC Exec_Type = iota
	_LOCATOR
	_DATE
)