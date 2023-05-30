& version

/ headings
[#]     = <h2 id="%%:unique_slug">%%</h2>
[##]    = <h3 id="%%:unique_slug">%%</h3>
[###]   = <h4 id="%%:unique_slug">%%</h4>
[####]  = <h5 id="%%:unique_slug">%%</h5>
[#####] = <h6 id="%%:unique_slug">%%</h6>

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

[aside] = {
	<aside>
		## Topics
		. %%
	</aside>
	<br clear="all">
}

[tt] = <tt>%%</tt>

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
	<link rel="stylesheet" type="text/css" href="%{static style.css}"/>
	<script type="text/javascript" src="%{static copy.js}" defer></script>

	if %spindle.is_server {
		. %spindle.reload_script
	}
</head>
<body>
	if %index tt {
		. [Your Site](/) | [Manual](%{page index})
	} else tt {
		. [Spindle %VERSION Manual](%{page index})
	}

	<h1>%title</h1>
	<main>%%</main>

	if !%index aside {
		> sidebar
	}
</body>
</html>