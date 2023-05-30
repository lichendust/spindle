& manual

title = Syntax

$ toc

# Tokens

Much like Markdown, Spindle uses non-alphanumeric 'tokens' at the start of a line to determine how that line is rendered.

The difference is that Spindle's tokens can be declared, either as part of a template, or on the fly.  This also means that any token can become anything — if you don't like `#` as H1, then feel free to change it.

If a new project is created using `spindle init`, a simple set of 'Markdown emulations' are pre-filled in the default template to get you moving faster:

code raw {
	# H1
	## H2
	### H2
	#### ...

	! image.jpg

	- Unordered list
	- Unordered list

	+ Ordered list
	+ Ordered list

	code raw {
		some_code_thing();
	}

	Regular lines will become paragraphs.
}

# Blocks

As you saw in the example above, Spindle's first major deviation from Markdown its block syntax.

code raw {
	block {
		/ something inside
	}
}

Blocks will wrap their contents in an additional template, allowing for subsections to be created, like a blockquote and citation around this quote:

code raw {
	quote {
		This is a quote.

		source = https://apple.com
		person = Tim Apple
	}
}

...which becomes, when combined:

quote {
	This is a quote.

	source = https://apple.com
	person = Tim Apple
}

# Declarations

Declarations are the way in which templating is created and manipulated.

## Variables

Base variables are declared like so:

code raw {
	page_title = My Blog Post
}

These can then be substituted into text or templates using the corresponding syntax:

code raw {
	<title>%page_title</title>
}

Missing variables will expand to nothing.

## Templating

code raw {
	[#] = <h1>%%</h1>
}

For consistency and to distinguish them from other variables, the named declarations that are used for blocks are also wrapped in square brackets:

code raw {
	[div] = <div>%%</div>
}

You can also use blocks within the declarations themselves — for example, the quote above:

code raw {
	[quote] = {
		<blockquote>
			. %%
			<cite>— [%person](%source)</cite>
		</blockquote>
	}
}

Tokens can only be declared as repeated instances of the same non-alphanumeric character:

code raw {
	[!]   = ...
	[??]  = ...
	[###] = ...
}

Blocks can only be declared as 'identifiers', which is to say they can only be composed of letters, numbers and underscores:

code raw {
	[div] = ...
	[group1] = ...
	[article_gallery] = ...
}

(Note that an identifier cannot *begin* with a number).

# Builtins

- `&` — Templates
- `~` — Imports
- `>` — Partials
- `×` — Unset
- `/` — Comments

# Semi-Builtins

- `.` — Raw

The singular `.` token (this is a regular full-stop/period, by the way) is a regular token available for any use, with the exception the Spindle renderer is hard-coded to not warn about its misuse.

Normally, when something that *looks* like a token but doesn't a have a corresponding template is used to start a line, Spindle will warn about it:  This is because it cannot distinguish between intentional or accidental use of a character in a token syntax, a deliberate lack of associated template declaration or a whole template file being somehow unlinked.  If a line *must* start with one or more repeating non-alphanumeric characters, it should be escaped.

However, the `.` token will *not* warn.  This is so it can be used to force a "raw" line, one which is not wrapped in whatever the chosen default is.

The `.` can still be templated with anything you like — it's not reserved like the true builtins, it just serves a default purpose.

# Resource Finders

Resource Finders are used to simplify linking between pages and assets within a Spindle site:

code raw {
	%{some-page}
	%{image.jpg}
	%{style.css}
}

Each of these finders will search through the tree of files in the `/source/` directory (top-down and breadth-first) and will return the first match it finds.  You can, if there are several assets or pages with the same name, provide a hint by adding a bit of the leading path (or even supplying the entire path):

code raw {
	%{data/style.css}
	%{docs/style.css}
}

Index pages are implicitly understood, however, and are simply accessed by supplying the directory path, though `dir/index` will also work.

Certain asset types have additional options:

## Images

code raw {
	%{image.jpg 1920x 90 webp}
}

These three arguments can be specified in any order, and reflect the following:

- `1920x` the maximum size (long-edge) that image should be reduced to.  In this example, images smaller than 1920 pixels would simply be left as is.

- `90` the quality that this image should be reduced or set as.  This does not apply to all formats and will be ignored in those cases.

- `webp` change the output format of this file, in this case to webp.

Spindle supports transforming images *to* the following formats:

- WEBP (requires [cwebp](https://developers.google.com/speed/webp/download) to be installed).
- JPEG
- PNG

It supports transforming images *from* these formats:

- WEBP (does not require cwebp).
- JPEG
- PNG
- TIFF

Any other image formats will simply be handled normally without transformation.