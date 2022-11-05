package main

const (
	default_hash uint32 = 2470140894 // "default"
	base_hash    uint32 = 537692064  // "%"
	it_hash      uint32 = 1194886160 // "it"
	stop_hash    uint32 = 722245873  // "."
	index_hash   uint32 = 151693739  // "index"
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
</head>
<body>%%</body>
</html>`

const index_template = `& main

title = Hello, World!

# Welcome to your new Spindle site!

The server you're currently accessing also hosts Spindle's [documentation](/_spindle/manual).`