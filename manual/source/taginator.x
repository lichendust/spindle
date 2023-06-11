& manual

title = Taginator



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

You cannot declare a taginator in a template: this will throw an error.

Taginators, by design, do not pursue through templating, otherwise most websites would very quickly enter a twilight zone of bizarre omissions and inclusions.  Taginators must be specified on the spawning page itself.

# Implicit Variables

When a taginator is invoked, it injects a whole host of useful goodies on the spawning page *and* the sub-pages.

## taginator.active

Evaluates truthily if we're currently on a sub-page and not the initial page.

## taginator.tag_name

Provides the name of the current page's tag. Only available on sub-pages.

## taginator.all_tags

Provides an alphabetically-sorted, duplicate-free array of all the tags found or connected to the initial page. Available on initial and sub-pages.

## taginator.source_url

Available everywhere, this expands to the initial page's URL (conforms to the current `spindle.toml` URL scheme).  On the initial page, this will be the same as `page.url`.