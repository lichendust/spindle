& manual

title = Scripting

$ toc

# Builtins

All builtins are implicitly injected.

## Data

code raw {
	info {
		line int
	}
}

Info provides information about this particular instance of a script call:

- its line number on a given page.

## Page

code raw {
	page {
		url           string
		canonical_url string
		file_path     string
	}
}

Page provides information about the page on which the current instance has been called from.

# Page Data

## Get

Get a declared variable value from the script call's position.

code raw {
	get(identifier) string, bool
}

Accepts an ID for a non-template declaration (meaning no [square brackets] around it) and returns it as a string, with a secondary success boolean that indicates anything was found.

## Get Token

Accepts an integer depth value and a variable number of strings, which will find all instances of a token on a page.

code raw {
	get_token(depth, matches...) []token
}

Each returned entry is an object with three fields:

code raw {
	token {
		token string
		text  string
		line  number
	}
}

Example

code raw {
	let all_headings = get_token(1, "#", "##", "###")
}

# Format Text

## Modify

code raw {
	text_modifier(text, modifier) string
}

`text_modifier` accepts a string and a builtin [modifier](#modifier), returning the appropriately formatted string.

The list of available modifiers is as follows:

code raw {
	modifier.slug
	modifier.unique_slug
	modifier.upper
	modifier.lower
	modifier.title
}

## Truncate

code raw {
	text_truncate(text, n) string
}

`text_truncate` shortens a string to the relevant `n` number of characters and appends an ellipsis to the end.

## Format Date

code raw {
	current_date(format) string
}

`current_date` take the Unix timestamp at time of build or refresh and formats it according to a simple set of formatters.  It mostly conforms to Apple's [[NSDateFormatter]](https://nsdateformatter.com) with some minor limitations .