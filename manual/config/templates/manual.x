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

/ define the aside
[aside] = {
	<aside>
		## Topics
		. %%
	</aside>
	<br clear="all">
}

/ tt shorthand
[tt] = <tt>%%</tt>

/ quote for the quote example
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
	<link rel="stylesheet" type="text/css" href="%{find style.css}"/>
	<script type="text/javascript" src="%{find copy.js}" defer></script>

	if %spindle.is_server {
		. %spindle.reload_script
	}
</head>
<body>
	/ different header if we're on the homepage
	if %homepage tt {
		. [Your Site](/) | [Manual](%{find index})
	} else tt {
		. [Your Site](/) | [Spindle %VERSION Manual](%{find index})
	}

	<h1>%title</h1>
	<main>
		. %%

		if !%homepage tt {
			<div style="height:100px"></div>
			. [Your Site](/) | [Top](#)
		}
	</main>

	/ include the sidebar if we're not on the homepage
	if !%homepage aside {
		> sidebar
	}
</body>
</html>