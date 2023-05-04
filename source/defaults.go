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

// hashes
const (
	DEFAULT_HASH uint32 = 2470140894 // "default"
	BASE_HASH    uint32 = 537692064  // "%"
	IT_HASH      uint32 = 1194886160 // "it"
	STOP_HASH    uint32 = 722245873  // "."
	INDEX_HASH   uint32 = 151693739  // "index"

	IS_SERVER_HASH     uint32 = 3202569332 // "spindle.is_server"
	RELOAD_SCRIPT_HASH uint32 = 3668506343 // "spindle.reload_script"

	URL_HASH              uint32 = 3862688697 // "page.url"
	CANONICAL_HASH        uint32 = 168974084  // "page.url_canonical"
	IMPORT_URL_HASH       uint32 = 4180644383 // "import.url"
	IMPORT_CANONICAL_HASH uint32 = 3998801026 // "import.url_canonical"

	TAGINATOR_HASH        uint32 = 296413604  // "taginator"
	TAGINATOR_ACTIVE_HASH uint32 = 1983695352 // "taginator.active"
	TAGINATOR_TAG_HASH    uint32 = 3373312024 // "taginator.tag_name"
	TAGINATOR_ALL_HASH    uint32 = 3084771407 // "taginator.all_tags"

	TAGINATOR_PARENT_HASH uint32 = 2722466740 // "taginator.parent_url"
)

const main_template = `/ markdown emulation
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
[code] = <pre><code>%%:raw</code></pre>



<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<title>%title</title>
	/ <script type="text/javascript" src="" defer></script>
	/ <link rel="stylesheet" type="text/css" href=""/>

	/ this allows you to hotload pages during local development
	if %spindle.is_server {
		. %spindle.reload_script
	}
</head>
<body>%%</body>
</html>`

const index_template = `& main

title = Hello, World!

# Welcome to your new Spindle site!

The server you're currently accessing also hosts Spindle's [documentation](/_spindle/manual).`