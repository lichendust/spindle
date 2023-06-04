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

// this file was generated by tool/embed.sh: don't modify!

package main

func manual_content(arg string) string {
	switch arg {
	case "copy.js":
		return `function set_copy(){if(!navigator.clipboard)return;const e="⌗";let t=document.querySelectorAll("pre");t.forEach(t=>{let s=document.createElement("button");s.className="copy mono",s.innerText=e,t.appendChild(s),s.addEventListener("click",async()=>{await n(t)})});async function n(e){let t=e.querySelector("code"),n=t.innerText;await navigator.clipboard.writeText(n)}}set_copy()`
	case "hot-reloading":
		return `<!DOCTYPE html><html><head><meta charset="utf-8"><title>Spindle — Hot Reloading</title><link rel="stylesheet" type="text/css" href="/_spindle/style.css"/><script type="text/javascript" src="/_spindle/copy.js" defer></script></head><body><tt><a href="/">Your Site</a> | <a href="/_spindle">Spindle v0.4.2 Manual</a></tt><h1>Hot Reloading</h1><main><p>Spindle injects a number of <a href="/_spindle/implicit-variables">implicit variables</a> for you to make use of.</p><p>In regards to hot-reloading, Spindle provides two relevant values to make use of:</p><ul><li><code>spindle.is_server</code> is a boolean value that allows you to distinguish between local development server mode and a final build output.</li><li><code>spindle.reload_script</code> is a pre-formatted chunk of Javascript that connects to the Spindle server and causes the browser to reload itself if the server detects changes to any files on disk.</li></ul><p>In the <code>&lt;head&gt;</code> of your site, you should add:</p><pre><code>if %spindle.is_server {
    . %spindle.reload_script
}</code></pre><p>The default <code>spindle init</code> template provides this already. Also, the <code>.</code> token here, if you have not overridden it, is used to force the line to have no templating — we don't want to wrap our reload script in a <code>&lt;p&gt;</code> tag.</p><tt><div style="height:100px"></div><a href="/">Your Site</a> | <a href="#">Top</a></tt></main><aside><h3 id="topics">Topics</h3><ul><li><a href="/_spindle/usage">Usage</a></li><li><a href="/_spindle/syntax">Syntax</a></li><li><a href="/_spindle/scripting">Scripting</a></li></ul><ul><li><a href="/_spindle/implicit-variables">Implicit Variables</a></li><li><a href="/_spindle/taginator">Taginator</a></li><li><a href="/_spindle/hot-reloading">Hot Reloading</a></li></ul></aside><br clear="all"></body></html>`
	case "implicit-variables":
		return `<!DOCTYPE html><html><head><meta charset="utf-8"><title>Spindle — Implicit Variables</title><link rel="stylesheet" type="text/css" href="/_spindle/style.css"/><script type="text/javascript" src="/_spindle/copy.js" defer></script></head><body><tt><a href="/">Your Site</a> | <a href="/_spindle">Spindle v0.4.2 Manual</a></tt><h1>Implicit Variables</h1><main><ul><li><a class="nu" href="#spindle">Spindle</a></li><li><a class="nu" href="#pages">Pages</a></li><li><a class="nu" href="#imports">Imports</a></li><li><a class="nu" href="#taginator">Taginator</a></li><li><a class="nu" href="#for-loops">For-Loops</a></li></ul><h2 id="spindle">Spindle</h2><p>The top-level <code>spindle</code> taxonomy provides a number of values relevant to the functioning of Spindle itself.</p><pre><code>%spindle.is_server
%spindle.reload_script</code></pre><p><code>spindle.is_server</code> is only true when Spindle is running its local development server.</p><p><code>spindle.reload_script</code> is a chunk of Javascript that allows browser connections to <a href="/_spindle/hot-reloading">hot-reload</a> when the Spindle server detects changes on disk.</p><h2 id="pages">Pages</h2><pre><code>%page.file_path
%page.url
%page.url_canonical</code></pre><h2 id="imports">Imports</h2><pre><code>%import.file_path
%import.url
%import.url_canonical</code></pre><h2 id="taginator">Taginator</h2><pre><code>%taginator
%taginator.active
%taginator.tag_name
%taginator.all_tags

%taginator.parent_url</code></pre><h2 id="for-loops">For-Loops</h2><p>Inside a for-loop, Spindle injects the variable <code>it</code>, which represents each value within the array being iterated over.</p><p>Because Spindle has no concept of numerical comparison, it also injects <code>last</code> for the final iteration of the loop.</p><pre><code>%it
%last</code></pre><tt><div style="height:100px"></div><a href="/">Your Site</a> | <a href="#">Top</a></tt></main><aside><h3 id="topics">Topics</h3><ul><li><a href="/_spindle/usage">Usage</a></li><li><a href="/_spindle/syntax">Syntax</a></li><li><a href="/_spindle/scripting">Scripting</a></li></ul><ul><li><a href="/_spindle/implicit-variables">Implicit Variables</a></li><li><a href="/_spindle/taginator">Taginator</a></li><li><a href="/_spindle/hot-reloading">Hot Reloading</a></li></ul></aside><br clear="all"></body></html>`
	default:
		return `<!DOCTYPE html><html><head><meta charset="utf-8"><title>Spindle — Spindle v0.4.2 Manual</title><link rel="stylesheet" type="text/css" href=".css}"/><script type="text/javascript" src=".js}" defer></script></head><body><tt><a href="/">Your Site</a> | [Manual]()</tt><h1>Spindle v0.4.2 Manual</h1><main><p>Welcome to the internal Spindle documentation. Spindle is a static site generator designed with total flexibility in mind. It is also not complete.</p><h2 id="contents">Contents</h2><ul><li><a href="usage}">Usage</a></li><li><a href="syntax}">Syntax</a></li><li><a href="scripting}">Scripting</a></li></ul><ul><li><a href="implicit-variables}">Implicit Variables</a></li><li><a href="taginator}">Taginator</a></li><li><a href="hot-reloading}">Hot Reloading</a></li></ul></main></body></html>`
	case "scripting":
		return `<!DOCTYPE html><html><head><meta charset="utf-8"><title>Spindle — Scripting</title><link rel="stylesheet" type="text/css" href="/_spindle/style.css"/><script type="text/javascript" src="/_spindle/copy.js" defer></script></head><body><tt><a href="/">Your Site</a> | <a href="/_spindle">Spindle v0.4.2 Manual</a></tt><h1>Scripting</h1><main><ul><li><a class="nu" href="#builtins">Builtins</a></li><li style="margin-left:2rem"><a class="nu" href="#data">Data</a></li><li style="margin-left:2rem"><a class="nu" href="#page">Page</a></li><li><a class="nu" href="#page-data">Page Data</a></li><li style="margin-left:2rem"><a class="nu" href="#get">Get</a></li><li style="margin-left:2rem"><a class="nu" href="#get-token">Get Token</a></li><li><a class="nu" href="#format-text">Format Text</a></li><li style="margin-left:2rem"><a class="nu" href="#modify">Modify</a></li><li style="margin-left:2rem"><a class="nu" href="#truncate">Truncate</a></li><li style="margin-left:2rem"><a class="nu" href="#format-date">Format Date</a></li></ul><h2 id="builtins">Builtins</h2><p>All builtins are implicitly injected.</p><h3 id="data">Data</h3><pre><code>info {
    line int
}</code></pre><p>Info provides information about this particular instance of a script call:</p><ul><li>its line number on a given page.</li></ul><h3 id="page">Page</h3><pre><code>page {
    url           string
    canonical_url string
    file_path     string
}</code></pre><p>Page provides information about the page on which the current instance has been called from.</p><h2 id="page-data">Page Data</h2><h3 id="get">Get</h3><p>Get a declared variable value from the script call's position.</p><pre><code>get(identifier) string, bool</code></pre><p>Accepts an ID for a non-template declaration (meaning no [square brackets] around it) and returns it as a string, with a secondary success boolean that indicates anything was found.</p><h3 id="get-token">Get Token</h3><p>Accepts an integer depth value and a variable number of strings, which will find all instances of a token on a page.</p><pre><code>get_token(depth, matches...) []token</code></pre><p>Each returned entry is an object with three fields:</p><pre><code>token {
    token string
    text  string
    line  number
}</code></pre><p>Example</p><pre><code>let all_headings = get_token(1, "#", "##", "###")</code></pre><h2 id="format-text">Format Text</h2><h3 id="modify">Modify</h3><pre><code>text_modifier(text, modifier) string</code></pre><p><code>text_modifier</code> accepts a string and a builtin <a href="#modifier">modifier</a>, returning the appropriately formatted string.</p><p>The list of available modifiers is as follows:</p><pre><code>modifier.slug
modifier.unique_slug
modifier.upper
modifier.lower
modifier.title</code></pre><h3 id="truncate">Truncate</h3><pre><code>text_truncate(text, n) string</code></pre><p><code>text_truncate</code> shortens a string to the relevant <code>n</code> number of characters and appends an ellipsis to the end.</p><h3 id="format-date">Format Date</h3><pre><code>current_date(format) string</code></pre><p><code>current_date</code> take the Unix timestamp at time of build or refresh and formats it according to a simple set of formatters. It mostly conforms to Apple's <a href="https://nsdateformatter.com" target="_blank" rel="noopener noreferrer">NSDateFormatter</a> with some minor limitations .</p><tt><div style="height:100px"></div><a href="/">Your Site</a> | <a href="#">Top</a></tt></main><aside><h3 id="topics">Topics</h3><ul><li><a href="/_spindle/usage">Usage</a></li><li><a href="/_spindle/syntax">Syntax</a></li><li><a href="/_spindle/scripting">Scripting</a></li></ul><ul><li><a href="/_spindle/implicit-variables">Implicit Variables</a></li><li><a href="/_spindle/taginator">Taginator</a></li><li><a href="/_spindle/hot-reloading">Hot Reloading</a></li></ul></aside><br clear="all"></body></html>`
	case "style.css":
		return `html{scroll-behavior:smooth}body{font-family:Atkinson Hyperlegible,Helvetica,Arial,sans-serif;margin:5ex;font-size:1.2rem}::selection{background:#66cdaa;color:#000!important}tt,code,.mono{font-family:DM Mono,SF Mono,Source Code Pro,Fira Code,Roboto Mono,monospace;font-size:1.07rem}tt,p{padding:0;margin:0;margin-bottom:.5ex}ul{padding-left:2ex;list-style-type:"×";list-style-position:outside}ul>li{padding-inline-start:1ex}a{color:#000}a:hover{color:#fff;background:#000}main{float:left;width:70ex;margin-right:2vw;margin-bottom:4vh}h2,h3{margin-top:2em;scroll-margin:30px}h2+h3{margin-top:0!important}aside{float:left;max-width:24ex}aside>*:first-child{margin:0!important}main>*:first-child{margin-top:0!important}section:not(:first-child){margin-top:2rem}pre{position:relative;padding:2ex;background:#000}code{white-space:nowrap;background:#eee;padding:.2ex .5ex}pre code{white-space:pre-wrap;background:0 0;padding:0;letter-spacing:0;color:#fff;font-size:1.2rem}.copy{user-select:none;cursor:pointer;position:absolute;right:20px;top:20px;padding:1px 4px;background:#fff;border:none}.copy:hover{background:#66cdaa}.copy:active{background:#fff}`
	case "syntax":
		return `<!DOCTYPE html><html><head><meta charset="utf-8"><title>Spindle — Syntax</title><link rel="stylesheet" type="text/css" href="/_spindle/style.css"/><script type="text/javascript" src="/_spindle/copy.js" defer></script></head><body><tt><a href="/">Your Site</a> | <a href="/_spindle">Spindle v0.4.2 Manual</a></tt><h1>Syntax</h1><main><ul><li><a class="nu" href="#tokens">Tokens</a></li><li><a class="nu" href="#blocks">Blocks</a></li><li><a class="nu" href="#declarations">Declarations</a></li><li style="margin-left:2rem"><a class="nu" href="#variables">Variables</a></li><li style="margin-left:2rem"><a class="nu" href="#templating">Templating</a></li><li><a class="nu" href="#builtins">Builtins</a></li><li style="margin-left:2rem"><a class="nu" href="#a-note-on-comments">A Note on Comments</a></li><li><a class="nu" href="#semi-builtins">Semi-Builtins</a></li><li><a class="nu" href="#resource-finders">Resource Finders</a></li><li style="margin-left:2rem"><a class="nu" href="#images">Images</a></li></ul><h2 id="tokens">Tokens</h2><p>Much like Markdown, Spindle uses non-alphanumeric 'tokens' at the start of a line to determine how that line is rendered.</p><p>The difference is that Spindle's tokens can be declared, either as part of a template, or on the fly. This also means that any token can become anything — if you don't like <code>#</code> as H1, then feel free to change it.</p><p>If a new project is created using <code>spindle init</code>, a simple set of 'Markdown emulations' are pre-filled in the default template to get you moving faster:</p><pre><code># H1
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

Regular lines will become paragraphs.</code></pre><h2 id="blocks">Blocks</h2><p>As you saw in the example above, Spindle's first major deviation from Markdown its block syntax.</p><pre><code>block {
    / something inside
}</code></pre><p>Blocks will wrap their contents in an additional template, allowing for subsections to be created, like a blockquote and citation around this quote:</p><pre><code>quote {
    This is a quote.

    source = https://apple.com
    person = Tim Apple
}</code></pre><p>...which becomes, when combined:</p><blockquote><p>This is a quote.</p><cite>— <a href="https://apple.com">Tim Apple</a></cite></blockquote><h2 id="declarations">Declarations</h2><p>Declarations are the way in which templating is created and manipulated.</p><h3 id="variables">Variables</h3><p>Base variables are declared like so:</p><pre><code>page_title = My Blog Post</code></pre><p>These can then be substituted into text or templates using the corresponding syntax:</p><pre><code>&lt;title&gt;%page_title&lt;/title&gt;</code></pre><p>Missing variables will expand to nothing.</p><h3 id="templating">Templating</h3><pre><code>[#] = &lt;h1&gt;%%&lt;/h1&gt;</code></pre><p>For consistency and to distinguish them from other variables, the named declarations that are used for blocks are also wrapped in square brackets:</p><pre><code>[div] = &lt;div&gt;%%&lt;/div&gt;</code></pre><p>You can also use blocks within the declarations themselves — for example, the quote above:</p><pre><code>[quote] = {
    &lt;blockquote&gt;
        . %%
        &lt;cite&gt;— <a href="%source">%person</a>&lt;/cite&gt;
    &lt;/blockquote&gt;
}</code></pre><p>Tokens can only be declared as repeated instances of the same non-alphanumeric character:</p><pre><code>[!]   = ...
[??]  = ...
[###] = ...</code></pre><p>Blocks can only be declared as 'identifiers', which is to say they can only be composed of letters, numbers and underscores:</p><pre><code>[div] = ...
[group1] = ...
[article_gallery] = ...</code></pre><p>(Note that an identifier cannot *begin* with a number).</p><h2 id="builtins">Builtins</h2><ul><li><code>&</code> — Templates</li><li><code>~</code> — Imports</li><li><code>></code> — Partials</li><li><code>×</code> — Unset</li><li><code>/</code> — Comments</li></ul><h3 id="a-note-on-comments">A Note on Comments</h3><p>The <code>/</code> token acts as both a comment *and* a block comment. If placed ahead of a block, the entire thing, including any children, will be ommitted.</p><pre><code>/ this is a comment

/ block {
    Everything in here is a comment.

    block {
        Even this as well.
    }
}</code></pre><h2 id="semi-builtins">Semi-Builtins</h2><ul><li><code>.</code> — Raw</li></ul><p>The singular <code>.</code> token (this is a regular full-stop/period, by the way) is a regular token available for any use, with the exception the Spindle renderer is hard-coded to not warn about its misuse.</p><p>Normally, when something that *looks* like a token but doesn't a have a corresponding template is used to start a line, Spindle will warn about it: This is because it cannot distinguish between intentional or accidental use of a character in a token syntax, a deliberate lack of associated template declaration or a whole template file being somehow unlinked. If a line *must* start with one or more repeating non-alphanumeric characters, it should be escaped.</p><p>However, the <code>.</code> token will *not* warn. This is so it can be used to force a "raw" line, one which is not wrapped in whatever the chosen default is.</p><p>The <code>.</code> can still be templated with anything you like — it's not reserved like the true builtins, it just serves a default purpose.</p><h2 id="resource-finders">Resource Finders</h2><p>Resource Finders are used to simplify linking between pages and assets within a Spindle site:</p><pre><code>%{some-page}
%{image.jpg}
%{style.css}</code></pre><p>Each of these finders will search through the tree of files in the <code>/source/</code> directory (top-down and breadth-first) and will return the first match it finds. You can, if there are several assets or pages with the same name, provide a hint by adding a bit of the leading path (or even supplying the entire path):</p><pre><code>%{data/style.css}
%{docs/style.css}</code></pre><p>Index pages are implicitly understood, however, and are simply accessed by supplying the directory path, though <code>dir/index</code> will also work.</p><p>Certain asset types have additional options:</p><h3 id="images">Images</h3><pre><code>%{image.jpg 1920x 90 webp}</code></pre><p>These three arguments can be specified in any order, and reflect the following:</p><ul><li><code>1920x</code> the maximum size (long-edge) that image should be reduced to. In this example, images smaller than 1920 pixels would simply be left as is.</li></ul><ul><li><code>90</code> the quality that this image should be reduced or set as. This does not apply to all formats and will be ignored in those cases.</li></ul><ul><li><code>webp</code> change the output format of this file, in this case to webp.</li></ul><p>Spindle supports transforming images *to* the following formats:</p><ul><li>WEBP (requires <a href="https://developers.google.com/speed/webp/download">cwebp</a> to be installed).</li><li>JPEG</li><li>PNG</li></ul><p>It supports transforming images *from* these formats:</p><ul><li>WEBP (does not require cwebp).</li><li>JPEG</li><li>PNG</li><li>TIFF</li></ul><p>Any other image formats will simply be handled normally without transformation.</p><tt><div style="height:100px"></div><a href="/">Your Site</a> | <a href="#">Top</a></tt></main><aside><h3 id="topics">Topics</h3><ul><li><a href="/_spindle/usage">Usage</a></li><li><a href="/_spindle/syntax">Syntax</a></li><li><a href="/_spindle/scripting">Scripting</a></li></ul><ul><li><a href="/_spindle/implicit-variables">Implicit Variables</a></li><li><a href="/_spindle/taginator">Taginator</a></li><li><a href="/_spindle/hot-reloading">Hot Reloading</a></li></ul></aside><br clear="all"></body></html>`
	case "taginator":
		return `<!DOCTYPE html><html><head><meta charset="utf-8"><title>Spindle — Taginator</title><link rel="stylesheet" type="text/css" href="/_spindle/style.css"/><script type="text/javascript" src="/_spindle/copy.js" defer></script></head><body><tt><a href="/">Your Site</a> | <a href="/_spindle">Spindle v0.4.2 Manual</a></tt><h1>Taginator</h1><main><ul><li><a class="nu" href="#setting-up-a-taginator">Setting up a Taginator</a></li><li><a class="nu" href="#caveats">Caveats</a></li><li><a class="nu" href="#utility-variables">Utility Variables</a></li></ul><p>The taginator is an opinionated construct that allows pages to spawn their own sub-pages based on the values of variables.</p><h2 id="setting-up-a-taginator">Setting up a Taginator</h2><p>A taginator can, in its simplest form, be created on a page with the following declaration. This must be placed before any of the relevant sections on the page:</p><pre><code>taginator = tags</code></pre><p>Spindle will immediately seek forward along the remainder of the page to find any instances of the declaration <code>tags</code>.</p><pre><code>taginator = tags

section {
    tags = foo

    Some content.
}

section {
    tags = foo bar

    Some other content.
}</code></pre><p>This page will now generate two sub-pages within a sibling directory to the page (the name of which is defined globally in <code>spindle.toml</code>).</p><p>By default, this directory is <code>/tags/</code>.</p><p>The example above will result in the following structure:</p><pre><code>.
├── index.x
└── tags/
    ├── foo.x
    └── bar.x</code></pre><p>Each of the sub-pages will be a copy of the original page, with the <code>section</code> blocks that do not apply ommitted.</p><p>The taginator is not just limited to blocks: The built-in <code>import</code> construct is also considered.</p><pre><code>taginator = tags

~ template some_page
~ template some_other_page</code></pre><p>The <code>tags</code> declaration must exist at the top-level scope in each of these pages for them to be sorted into the sub-pages.</p><p>You can also mix the two constructs with impunity:</p><pre><code>taginator = tags

{
    section {
        tags  = foo
        image = image.jpg
    }

    ~ section some_page
}</code></pre><p>If you wish for the scope of the taginator to be limited — you have blocks and imports at the bottom of your page that should not be considered, for example — you can limit the scope of the taginator within a block.</p><pre><code>header {
    blah blah
}

{
    taginator = tags

    ~ section some_page
    ~ section some_other_page
}

footer {
    copyright, etc.
}</code></pre><h2 id="caveats">Caveats</h2><p>You cannot create a taginator in a template. Taginators, by design, do not pursue through templating, otherwise most websites would very quickly enter a twilight zone of bizarre omissions and inclusions. Taginators must be specified on the spawning page itself.</p><p>Taginators *do* pursue through partials, but only for imports. Blocks in partials are not considered.</p><h2 id="utility-variables">Utility Variables</h2><p>When a taginator is invoked, it injects a whole host of useful goodies on the spawning page *and* the sub-pages.</p><p>This is includes:</p><ul><li><code>taginator.active</code>: this variable evaluates to true if we're currently on a sub-page and not the original.</li><li><code>taginator.tag_name</code>: this variable exists only on sub-pages and provides the current tag being generated.</li><li><code>taginator.all_tags</code>: this variable is available on all pages and provides an alphabetical array of all the tags found.</li><li><code>taginator.parent_url</code>: always evaluates to the originating page's URL (even on the originating page itself), allowing for consistent lists or clouds to be created.</li></ul><tt><div style="height:100px"></div><a href="/">Your Site</a> | <a href="#">Top</a></tt></main><aside><h3 id="topics">Topics</h3><ul><li><a href="/_spindle/usage">Usage</a></li><li><a href="/_spindle/syntax">Syntax</a></li><li><a href="/_spindle/scripting">Scripting</a></li></ul><ul><li><a href="/_spindle/implicit-variables">Implicit Variables</a></li><li><a href="/_spindle/taginator">Taginator</a></li><li><a href="/_spindle/hot-reloading">Hot Reloading</a></li></ul></aside><br clear="all"></body></html>`
	case "usage":
		return `<!DOCTYPE html><html><head><meta charset="utf-8"><title>Spindle — Usage</title><link rel="stylesheet" type="text/css" href="/_spindle/style.css"/><script type="text/javascript" src="/_spindle/copy.js" defer></script></head><body><tt><a href="/">Your Site</a> | <a href="/_spindle">Spindle v0.4.2 Manual</a></tt><h1>Usage</h1><main><ul><li><a class="nu" href="#getting-started">Getting Started</a></li><li><a class="nu" href="#serve">Serve</a></li><li style="margin-left:2rem"><a class="nu" href="#options">Options</a></li><li><a class="nu" href="#build">Build</a></li><li style="margin-left:2rem"><a class="nu" href="#options-1">Options</a></li></ul><h2 id="getting-started">Getting Started</h2><p>Initialise a new project *within* an existing directory:</p><pre><code>spindle init</code></pre><h2 id="serve">Serve</h2><p>Launch a local server to work on the site in real-time, with hot reloading:</p><pre><code>spindle serve</code></pre><h3 id="options">Options</h3><p>Most options are configured inside of <code>spindle.toml</code>, but there are a few overridable options at the command-line level. In the event of overlap, the command line argument takes precedence, then <code>spindle.toml</code>, then the builtin default value.</p><ul><li><code>-p [number]</code> or <code>--port [number]</code>: override the port number for the local development server.</li></ul><h2 id="build">Build</h2><p>Output a deployment-ready copy of the site to a <code>/public/</code> directory:</p><pre><code>spindle build</code></pre><h3 id="options-1">Options</h3><p>Spindle makes use of a piece of syntax called a <a href="/_spindle/syntax#resource-finders">'Resource Finder'</a>: this streamlines the process of linking between pages — you can just put in the page's name, not its full path — and the resulting link styles can all be changed according to the needs of the site in <code>spindle.toml</code> (absolute, relative, etc.).</p><p>You can of course still manually supply links, but it's not recommended for the second reason: Resource Finders track the usage of any given asset or page, and by default <code>spindle build</code> will only output pages that are not orphaned. The root index is assumed to be 'live' as the starting point, and anything reachable from the subsequent tree is included.</p><p>Practically speaking, if you commented out a link in your navigation bar, this would typically make a chunk of your site unreachable to the average user. If the navigation bar uses Resource Finders to create its links, commenting an entry out will remove that entire branch of the site from the build output.</p><p>All of this is to explain this next flag:</p><ul><li><code>-a</code> or <code>--all</code>: this forces Spindle to always render every asset in the source directory, ignoring their orphan status.</li></ul><p>Spindle typically builds a site in just a few milliseconds, but it can also be directed to perform image conversions and resizing. This process is by nature, *much slower*.</p><ul><li><code>--skip-images</code> will ignore any images located by Resource Finders during the build, which is useful if you've just built the entire site and don't want to wait another thirty seconds for a full rebuild to fix a typo.</li></ul><tt><div style="height:100px"></div><a href="/">Your Site</a> | <a href="#">Top</a></tt></main><aside><h3 id="topics">Topics</h3><ul><li><a href="/_spindle/usage">Usage</a></li><li><a href="/_spindle/syntax">Syntax</a></li><li><a href="/_spindle/scripting">Scripting</a></li></ul><ul><li><a href="/_spindle/implicit-variables">Implicit Variables</a></li><li><a href="/_spindle/taginator">Taginator</a></li><li><a href="/_spindle/hot-reloading">Hot Reloading</a></li></ul></aside><br clear="all"></body></html>`
	}
}

const HELP_TEXT = `
$1Usage$0
-----

    spindle command [--flags]

$1Commands$0
--------

    $1init$0     start a project in the current directory
    $1serve$0    start the local development server
    $1build$0    render the project to a build directory

$1Further Information$0
-------------------

The local development server built into Spindle serves a 
complete browsable manual under the path $1/_spindle/$0.`