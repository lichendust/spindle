& manual

title = Hot Reloading

Spindle injects a number of [implicit variables](%{find implicit-variables}) for you to make use of.

In regards to hot-reloading, Spindle provides two relevant values to make use of:

- `spindle.is_server` is a boolean value that allows you to distinguish between local development server mode and a final build output.
- `spindle.reload_script` is a pre-formatted chunk of Javascript that connects to the Spindle server and causes the browser to reload itself if the server detects changes to any files on disk.

In the `&lt;head&gt;` of your site, you should add:

code raw {
	if %spindle.is_server {
		. %spindle.reload_script
	}
}

The default `spindle init` template provides this already.  Also, the `.` token here, if you have not overridden it, is used to force the line to have no templating — we don't want to wrap our reload script in a `&lt;p&gt;` tag.