if (args.length === 0) {
	var the_array = get_token(1, "#", "##")
} else {
	var the_array = get_token(1, ...args)
}

let result = ''

for (var i = 0; i < the_array.length; i++) {
	let x = the_array[i]
	let s = text_modifier(x.text, modifier.unique_slug)

	switch(x.token) {
	case "#":
		result += '<li>'
		break
	case "##":
		result += '<li style="margin-left:2rem">'
		break
	}

	result += '<a class="nu" href="#' + s + '">' + x.text + '</a></li>'
}

return "<ul class=\"monospace\">" + result + "</ul>"