& manual

title = Taginator

$ toc

The taginator is an opinionated construct that allows pages to spawn their own sub-pages based on the values of variables.

# Setting up a Taginator

A taginator can, in its simplest form, be created on a page with the following declaration.  This must be placed before any of the relevant sections on the page:

code raw {
	taginator = tags
}

Spindle will immediately seek forward along the remainder of the page to find any instances of the declaration `tags`.

code raw {
	taginator = tags

	section {
		tags = foo

		Some content.
	}

	section {
		tags = foo bar

		Some other content.
	}
}

This page will now generate two sub-pages within a sibling directory to the page (the name of which is defined globally in `spindle.toml`).

By default, this directory is `/tags/`.

The example above will result in the following structure:

code raw {
	.
	├── index.x
	└── tags/
	    ├── foo.x
	    └── bar.x
}

Each of the sub-pages will be a copy of the original page, with the `section` blocks that do not apply ommitted.

The taginator is not just limited to blocks:  The built-in `import` construct is also considered.

code raw {
	taginator = tags

	~ template some_page
	~ template some_other_page
}

The `tags` declaration must exist at the top-level scope in each of these pages for them to be sorted into the sub-pages.

You can also mix the two constructs with impunity:

code raw {
	taginator = tags

	{
		section {
			tags  = foo
			image = image.jpg
		}

		~ section some_page
	}
}

If you wish for the scope of the taginator to be limited — you have blocks and imports at the bottom of your page that should not be considered, for example — you can limit the scope of the taginator within a block.

code raw {
	header {
		blah blah
	}

	{
		taginator = tags

		~ section some_page
		~ section some_other_page
	}

	footer {
		copyright, etc.
	}
}

# Caveats

You cannot create a taginator in a template.  Taginators, by design, do not pursue through templating, otherwise most websites would very quickly enter a twilight zone of bizarre omissions and inclusions.  Taginators must be specified on the spawning page itself.

Taginators *do* pursue through partials, but only for imports.  Blocks in partials are not considered.

# Utility Variables

When a taginator is invoked, it injects a whole host of useful goodies on the spawning page *and* the sub-pages.

This is includes:

- `taginator.active`: this variable evaluates to true if we're currently on a sub-page and not the original.
- `taginator.tag_name`: this variable exists only on sub-pages and provides the current tag being generated.
- `taginator.all_tags`: this variable is available on all pages and provides an alphabetical array of all the tags found.
- `taginator.parent_url`: always evaluates to the originating page's URL (even on the originating page itself), allowing for consistent lists or clouds to be created.