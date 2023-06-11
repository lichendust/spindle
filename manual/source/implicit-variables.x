& manual

title = Implicit Variables



# Spindle

The top-level `spindle` taxonomy provides a number of values relevant to the functioning of Spindle itself.

code raw {
	%spindle.is_server
	%spindle.reload_script
}

`spindle.is_server` is only true when Spindle is running its local development server.
`spindle.reload_script` is a chunk of Javascript that allows browser connections to [hot-reload](%{link hot-reloading}) when the Spindle server detects changes on disk.

# Pages

code raw {
	%page.file_path
	%page.url
	%page.canonical_url
}

# Imports

code raw {
	%import.file_path
	%import.url
	%import.canonical_url
}

# Taginator

code raw {
	%taginator
	%taginator.active
	%taginator.tag_name
	%taginator.all_tags

	%taginator.parent_url
}

# For-Loops

Inside a for-loop, Spindle injects the variable `it`, which represents each value within the array being iterated over.

Because Spindle has no concept of numerical comparison, it also injects `last` for the final iteration of the loop.

code raw {
	%it
	%last
}
