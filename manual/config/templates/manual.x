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



<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<title>Spindle — %title</title>
	<link rel="stylesheet" type="text/css" href="%{static style.css}"/>

	/ this allows you to hotload pages during local development
	if %spindle.is_server {
		. %spindle.reload_script
	}
</head>
<body>
	<main>
		if %index {
			<tt><a href="%{page index}">Manual</a></tt>
		} else {
			<tt><a href="%{page index}">Spindle %VERSION Manual</a></tt>
		}

		<h1>%title</h1>

		. %%
	</main>

	if ! %index aside {
		> sidebar
	}
</body>
</html>