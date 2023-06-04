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

// filepaths
const (
	EXTENSION = ".x"

	SOURCE_PATH = "source"
	PUBLIC_PATH = "public"
	CONFIG_PATH = "config"
	CONFIG_FILE_PATH = "config/spindle.toml"

	TEMPLATE_PATH = CONFIG_PATH + "/templates"
	PARTIAL_PATH  = CONFIG_PATH + "/partials"
	SCRIPT_PATH   = CONFIG_PATH + "/scripts"
)

const MAIN_TEMPLATE = `/ markdown emulation
/ headings
[#]      = <h1 id="%%:unique_slug">%%</h1>
[##]     = <h2 id="%%:unique_slug">%%</h2>
[###]    = <h3 id="%%:unique_slug">%%</h3>
[####]   = <h4 id="%%:unique_slug">%%</h4>
[#####]  = <h5 id="%%:unique_slug">%%</h5>
[######] = <h6 id="%%:unique_slug">%%</h6>

/ "default" means a regular line with no leading token
[default] = <p>%%</p>

/ images
[!] = <img src="%1" alt="%2">

/ lists
{-} = <ul>%%</ul>
[-] = <li>%%</li>

{+} = <ol>%%</ol>
[+] = <li>%%</li>

/ codeblocks
[code] = <pre><code>%%</code></pre>



<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<title>%title</title>

	/ this is Spindle's internal style.css for its manual pages
	/ don't use this, because it won't build; it's just a nice example
	<script type="text/javascript" src="/_spindle/style.css" defer></script>

	/ this allows you to hotload pages during local development
	if %spindle.is_server {
		. %spindle.reload_script
	}
</head>
<body>%%</body>
</html>`

const INDEX_TEMPLATE = `& main

title = Hello, World!

This is the basic index page that you just created when initialising your project.

When using the local development server, Spindle sneakily serves an additional path, <code>/_spindle</code>, which hosts all of the documentation you need to understand it.  (Don't worry, these pages aren't added to your site.)

Here's a [link](/_spindle).  It's also helpfully shown when Spindle encounters an error while hot-reloading.

- [GitHub](https://github.com/qxoko/spindle)
- [Manual](/_spindle)`

const CONFIG_TEMPLATE = `domain = "https://website.com/"

path_mode  = "absolute"
build_path = "public"
tag_path   = "tag"

port_number = ":3011"

# default settings for image linkers
# applied to any images without inline settings
# image_quality = 90
# image_size    = 1920
# image_format  = "webp"

[[inline]]
pattern  = '\[(.+?)\]\((.+?)\)'
template = '<a href="$2">$1</a>'

[[inline]]
pattern  = '\*\*(\S(.+?)\S)\*\*'
template = '<b>$1</b>'

[[inline]]
pattern  = '\*(\S(.+?)\S)\*'
template = '<i>$1</i>'`