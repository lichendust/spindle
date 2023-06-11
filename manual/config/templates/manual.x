& version

[#]  = <h2 id="%%:unique_slug">%%</h2>
[##] = <h3 id="%%:unique_slug">%%</h3>

[!] = <img src="%1" alt="%2">

{-} = <ul>%%</ul>
[-] = <li>%%</li>

{+} = <ol>%%</ol>
[+] = <li>%%</li>

[default] = <p>%%</p>
[code]    = <pre><code>%%</code></pre>
[tt]      = <tt>%%</tt>

[aside] = {
	<aside>
		if !%homepage {
			## Contents

			$ toc # ##

			## Other Topics
		} else {
			## Topics
		}

		. %%
	</aside>
	<br clear="all">
}

[quote] = {
	<blockquote>
		. %%
		<cite>— <a href="%source">%person</a></cite>
	</blockquote>
}



<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<title>Spindle — %title</title>
	<link rel="stylesheet" type="text/css" href="%{link style.css}"/>
	<script type="text/javascript" src="%{link copy.js}" defer></script>

	if %spindle.is_server {
		. %spindle.reload_script
	}
</head>
<body>
	version = <a href="%{link index}">Spindle %VERSION Manual</a>

	<tt><a href="/">Your Site</a> | %version</tt>
	<h1>%title</h1>
	<main>
		. %%
		<div style="height:100px"></div>
		<tt><a href="#">Top</a> | %version</tt>
	</main>

	if !%homepage aside {
		> sidebar
	}
</body>
</html>