& manual

title = Scripting

$ toc

# Implicit Values and Types

## `data`

code raw {
	data.line
}

Gives the position, in terms of line number, of this script's call location on the current page.

## `script_token`

code raw {
	struct script_token {
		token string
		text  string
		line  number
	}
}

(Only supplied by Spindle through other functions).

## `modifier`

code raw {
	modifier.slug
	modifier.unique_slug
	modifier.upper
	modifier.lower
	modifier.title
}

An enum map corresponding to the Spindle variable modifiers, used for text manipulation.

# Functions

## `text_modifier`

code raw {
	text_modifier(text, modifier) string
}

Accepts a string and a builtin [modifier](#modifier), returning the appropriately formatted string.

## Page Contents

code raw {
	get(identifier) string, bool
}

Accepts an ID for a non-template declaration (meaning no [square brackets] around it) and returns it as a string, with a secondary success boolean that indicates anything was found.

code raw {
	get_token(depth, matches...) []script_token
}

Accepts an integer depth value and a variable number of strings, which will find all instances of a token on a page.