package main

const (
	meta_source       = `<meta property='%s' content='%s'>`
	meta_description  = `<meta name='description' content='%s'>`
	meta_canonical    = `<link rel='canonical' href='%s'>`
	meta_favicon      = `<link rel='icon' type='image/%s' href='%s'>`
)

const (
	script_template = `<script type='text/javascript' src='%s' defer></script>`
	style_template  = `<link rel='stylesheet' type='text/css' href='%s'/>`
)

var valid_twitter_card = map[string]bool {
	"summary": true,
	"summary_large_image": true,
	"app": true,
	"player": true,
}

var tag_defaults = map[string]string {
	"h1": "<h1 id='%s'>%s</h1>",
	"h2": "<h2 id='%s'>%s</h2>",
	"h3": "<h3 id='%s'>%s</h3>",
	"h4": "<h4 id='%s'>%s</h4>",
	"h5": "<h5 id='%s'>%s</h5>",
	"h6": "<h6 id='%s'>%s</h6>",

	"img": "<img src='%s'>",
	"link": "<a href='%s'>%s</a>",
	"paragraph": "<p>%s</p>",

	"ol": "<ol>%s</ol>",
	"ul": "<ul>%s</ul>",
	"li": "<li>%s</li>",

	"hr": "<hr>",

	"code": "<code>%s</code>",
	"codeblock": "<pre><code>%s</code></pre>",
}

var html_defaults = map[string]bool {
	"abbr": true,
	"address": true,
	"area": true,
	"article": true,
	"aside": true,
	"audio": true,
	"b": true,
	"base": true,
	"bdo": true,
	"blockquote": true,
	"body": true,
	"button": true,
	"canvas": true,
	"caption": true,
	"cite": true,
	"code": true,
	"col": true,
	"colgroup": true,
	"command": true,
	"datalist": true,
	"dd": true,
	"del": true,
	"details": true,
	"dfn": true,
	"div": true,
	"dl": true,
	"dt": true,
	"em": true,
	"embed": true,
	"fieldset": true,
	"figcaption": true,
	"figure": true,
	"footer": true,
	"form": true,
	"h1": true,
	"head": true,
	"header": true,
	"html": true,
	"i": true,
	"iframe": true,
	"img": true,
	"input": true,
	"ins": true,
	"kbd": true,
	"label": true,
	"legend": true,
	"li": true,
	"link": true,
	"map": true,
	"mark": true,
	"menu": true,
	"meta": true,
	"meter": true,
	"nav": true,
	"noscript": true,
	"object": true,
	"ol": true,
	"optgroup": true,
	"option": true,
	"output": true,
	"p": true,
	"param": true,
	"pre": true,
	"progress": true,
	"q": true,
	"rp": true,
	"rt": true,
	"ruby": true,
	"s": true,
	"samp": true,
	"script": true,
	"section": true,
	"select": true,
	"small": true,
	"source": true,
	"span": true,
	"strong": true,
	"sub": true,
	"sup": true,
	"table": true,
	"tbody": true,
	"td": true,
	"textarea": true,
	"tfoot": true,
	"th": true,
	"thead": true,
	"time": true,
	"title": true,
	"tr": true,
	"track": true,
	"u": true,
	"ul": true,
	"var": true,
	"video": true,
	"wbr": true,
	"main": true,
}

// new project
const config_template = `domain = https://website.com`
const index_template  = `# Index

Welcome to your new Spindle project!`

const new_project_message = `created new spindle project!

first steps:

    1. fill out config/config.x
    2. put some content into source/index.x
    3. run a spindle command

commands:

	spindle serve   start a local development server
	spindle build   render your project to /public/

initial layout:

    .
    ├── config
    │   ├── config.x
    │   ├── chunks
    │   └── plates
    └── source
        └── index.x`

const help_message = `commands:

    spindle serve   start a local development server
    spindle build   render your project to /public/
`

const (
	sitemap_template = `<?xml version="1.0" encoding="utf-8" standalone="yes"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:xhtml="http://www.w3.org/1999/xhtml">`
	sitemap_entry = `<url><loc>%s</loc></url>`
)