package main

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

	"img":       "<img src='%s'>",
	"link":      "<a href='%s'>%s</a>",
	"paragraph": "<p>%s</p>",

	"ol": "<ol>%s</ol>",
	"ul": "<ul>%s</ul>",
	"li": "<li>%s</li>",

	"hr": "<hr>",

	"code":      "<code>%s</code>",
	"codeblock": "<pre><code>%s</code></pre>",
}

var html_defaults = map[string]string{
	"abbr":       "<abbr>%s</abbr>",
	"address":    "<address>%s</address>",
	"area":       "<area>%s</area>",
	"article":    "<article>%s</article>",
	"aside":      "<aside>%s</aside>",
	"audio":      "<audio>%s</audio>",
	"b":          "<b>%s</b>",
	"base":       "<base>%s</base>",
	"bdo":        "<bdo>%s</bdo>",
	"blockquote": "<blockquote>%s</blockquote>",
	"body":       "<body>%s</body>",
	"button":     "<button>%s</button>",
	"canvas":     "<canvas>%s</canvas>",
	"caption":    "<caption>%s</caption>",
	"cite":       "<cite>%s</cite>",
	"code":       "<code>%s</code>",
	"col":        "<col>%s</col>",
	"colgroup":   "<colgroup>%s</colgroup>",
	"command":    "<command>%s</command>",
	"datalist":   "<datalist>%s</datalist>",
	"dd":         "<dd>%s</dd>",
	"del":        "<del>%s</del>",
	"details":    "<details>%s</details>",
	"dfn":        "<dfn>%s</dfn>",
	"div":        "<div>%s</div>",
	"dl":         "<dl>%s</dl>",
	"dt":         "<dt>%s</dt>",
	"em":         "<em>%s</em>",
	"embed":      "<embed>%s</embed>",
	"fieldset":   "<fieldset>%s</fieldset>",
	"figcaption": "<figcaption>%s</figcaption>",
	"figure":     "<figure>%s</figure>",
	"footer":     "<footer>%s</footer>",
	"form":       "<form>%s</form>",
	"h1":         "<h1>%s</h1>",
	"head":       "<head>%s</head>",
	"header":     "<header>%s</header>",
	"html":       "<html>%s</html>",
	"i":          "<i>%s</i>",
	"iframe":     "<iframe>%s</iframe>",
	"img":        "<img>%s</img>",
	"input":      "<input>%s</input>",
	"ins":        "<ins>%s</ins>",
	"kbd":        "<kbd>%s</kbd>",
	"label":      "<label>%s</label>",
	"legend":     "<legend>%s</legend>",
	"li":         "<li>%s</li>",
	"link":       "<link>%s</link>",
	"map":        "<map>%s</map>",
	"mark":       "<mark>%s</mark>",
	"menu":       "<menu>%s</menu>",
	"meta":       "<meta>%s</meta>",
	"meter":      "<meter>%s</meter>",
	"nav":        "<nav>%s</nav>",
	"noscript":   "<noscript>%s</noscript>",
	"object":     "<object>%s</object>",
	"ol":         "<ol>%s</ol>",
	"optgroup":   "<optgroup>%s</optgroup>",
	"option":     "<option>%s</option>",
	"output":     "<output>%s</output>",
	"p":          "<p>%s</p>",
	"param":      "<param>%s</param>",
	"pre":        "<pre>%s</pre>",
	"progress":   "<progress>%s</progress>",
	"q":          "<q>%s</q>",
	"rp":         "<rp>%s</rp>",
	"rt":         "<rt>%s</rt>",
	"ruby":       "<ruby>%s</ruby>",
	"s":          "<s>%s</s>",
	"samp":       "<samp>%s</samp>",
	"script":     "<script>%s</script>",
	"section":    "<section>%s</section>",
	"select":     "<select>%s</select>",
	"small":      "<small>%s</small>",
	"source":     "<source>%s</source>",
	"span":       "<span>%s</span>",
	"strong":     "<strong>%s</strong>",
	"sub":        "<sub>%s</sub>",
	"sup":        "<sup>%s</sup>",
	"table":      "<table>%s</table>",
	"tbody":      "<tbody>%s</tbody>",
	"td":         "<td>%s</td>",
	"textarea":   "<textarea>%s</textarea>",
	"tfoot":      "<tfoot>%s</tfoot>",
	"th":         "<th>%s</th>",
	"thead":      "<thead>%s</thead>",
	"time":       "<time>%s</time>",
	"title":      "<title>%s</title>",
	"tr":         "<tr>%s</tr>",
	"track":      "<track>%s</track>",
	"u":          "<u>%s</u>",
	"ul":         "<ul>%s</ul>",
	"var":        "<var>%s</var>",
	"video":      "<video>%s</video>",
	"wbr":        "<wbr>%s</wbr>",
	"main":       "<main>%s</main>",
}

// new project
const config_template = `domain = https://website.com`
const index_template  = `# Index

Welcome to your new Spindle project!`

const startup_error = `not a spindle project!`
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
	meta_source       = "<meta property='%s' content='%s'>"
	meta_description  = "<meta name='description' content='%s'>"
	meta_canonical    = "<link rel='canonical' href='%s'>"
	meta_favicon      = "<link rel='icon' type='image/%s' href='%s'>"
	meta_viewport     = "<meta name='viewport' content='width=device-width, initial-scale=%s'>"
	meta_theme        = "<meta name='theme-color' content='%s'>"
)

const (
	script_template = "<script type='text/javascript' src='%s' defer></script>"
	style_template  = "<link rel='stylesheet' type='text/css' href='%s'/>"
)

const (
	sitemap_template = `<?xml version="1.0" encoding="utf-8" standalone="yes"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:xhtml="http://www.w3.org/1999/xhtml">`
	sitemap_entry = `<url><loc>%s</loc></url>`
)

const (
	media_vimeo_template   = `<div class='video'><iframe src='https://player.vimeo.com/video/%s?color=0&title=0&byline=0&portrait=0' frameborder='0' allow='fullscreen' allowfullscreen></iframe></div>`
	media_youtube_template = `<div class='video'><iframe src='https://www.youtube-nocookie.com/embed/%s?rel=0&controls=1' frameborder='0' allow='accelerometer; encrypted-media; gyroscope; picture-in-picture' allowfullscreen></iframe></div>`
)