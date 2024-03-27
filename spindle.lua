--[[
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
]]

-- 'spindle' is a globally initialised table created by the spindle
-- backing executable; you can just start accessing it right away

spindle.handlers     = {} -- file types
spindle.tokens       = {} -- line-level tokens; headings, images, etc.
spindle.inlines      = {} -- inline syntax; links, bold, etc.
spindle.modifiers    = {} -- %variable:modifiers
spindle.all_pages    = {}
spindle.all_files    = {}
spindle.lock_file    = {} -- temp map for 'other' files @todo
spindle.markup_cache = {}

spindle.output_path = "_site/"

spindle.handlers[".x"] = function(file_path)
	spindle.parse_markup = spindle.__parse_markup
	local page = spindle.load_page(file_path)
	return page._output_path, page.canonical_url
end

spindle.tokens[">"] = function(page, arg)
	spindle.parse_markup = spindle.__parse_markup
	local part = spindle.load_markup("_data/" .. arg .. ".x")
	return spindle.render(page, part.syntax_tree)
end

-- @todo make this safe my goodness
spindle.tokens["~"] = function(page, arg)
	spindle.parse_markup = spindle.__parse_markup
	local args = spindle.split_quoted(arg)
	local part = spindle.load_markup("_data/" .. args[1] .. ".x")
	local path = spindle.find_file(args[2])
	return spindle.render(spindle.load_page(path), part.syntax_tree)
end

spindle.tokens["#"]   = '<h1 id="%0:slug">%0</h1>'
spindle.tokens["##"]  = '<h2 id="%0:slug">%0</h2>'
spindle.tokens["###"] = '<h3 id="%0:slug">%0</h3>'

spindle.tokens["-"] = { main = '<li>%0</li>', wrap = '<ul>%0</ul>' }
spindle.tokens["+"] = { main = '<li>%0</li>', wrap = '<ol>%0</ol>' }

spindle.tokens["!"] = function(page, arg)
	local _, url = spindle.process_any_file(arg)
	return spindle.template_expand('<img src="%0">', url)
end

spindle.tokens.block_default = '%0'
spindle.tokens.default       = '<p>%0</p>'

-- @hack to ensure these exist
-- but skip the typecheck
spindle.tokens["$"]  = 0
spindle.tokens["$$"] = 0

spindle.to_upper = string.upper
spindle.to_lower = string.lower
spindle.to_quote = function(arg)
	return string.format("%q", arg)
end

spindle.modifiers.slug  = spindle.to_slug
spindle.modifiers.title = spindle.to_title
spindle.modifiers.upper = spindle.to_upper
spindle.modifiers.lower = spindle.to_lower
spindle.modifiers.quote = spindle.to_quote

table.insert(spindle.inlines, function(page, line)
	return line:gsub("%[(.-)%]%((.-)%)", spindle.make_anchor_tag)
end)

table.insert(spindle.inlines, function(page, line)
	return line:gsub("%%{(.-)}", function(x)
		local _, url = spindle.process_any_file(x)
		return url
	end)
end)



function spindle.make_anchor_tag(label, path)
	if spindle.has_protocol(path) then
		return string.format('<a href="%s">%s</a>', path, label)
	end

	if path:sub(1, 1) == '#' then
		return string.format('<a href="%s">%s</a>', path, label)
	end

	local frag = path:match("#.+$")
	if frag == nil then
		frag = ""
	else
		path = path:gsub("#.+$", "")
	end

	local _, url = spindle.process_any_file(path)
	return string.format('<a href="%s%s">%s</a>', url, frag, label)
end

-- [[STRING UTILS]]
function spindle.has_protocol(text)
	return text:match("^%a+://") ~= nil
end

function spindle.url_from_path(path)
	return (spindle.domain .. path):gsub("/index.*$", ""):gsub("%.x$", ""):gsub("%.html$", "") -- @todo hard-coded
end

function spindle.short_ext(text)
	return text:match("%.[^%./]+$")
end

function spindle.long_ext(text)
	return text:match("%.[^/]+$")
end

function spindle.dir(text)
	return text:gsub("(.*/)(.*)", "%1")
end

function spindle.basename(text)
	return text:gsub("(.*/)(.*)", "%2")
end

function spindle.root_basename(text) -- @todo might rename
	return spindle.basename(text):sub(1, -spindle.long_ext(text):len() - 1)
end

function spindle.trim(s)
	return string.gsub(s, "^%s*(.-)%s*$", "%1")
end

-- modified from https://stackoverflow.com/a/36958689
local function escape_magic(s)
	local magic_set = '[()%%.[^$%]*+%-?]'
	if s == nil then return end
	return (s:gsub(magic_set, '%%%1'))
end

function string:gsplit(delimiter)
	if self:sub(-#delimiter) ~= delimiter then
		self = self .. delimiter
	end
	return self:gmatch('(.-)' .. escape_magic(delimiter))
end
-- end stackoverflow

function string:split(delimiter)
	local t = {}
	for item in self:gsplit(delimiter) do
		table.insert(t, item)
	end
	return t
end

-- this is a workaround hack for Lua's lack of non-inclusive negative matches
-- essentially, you define the escaped condition *and* the correct condition
-- and it will return either the now-unescaped correct text or the full replacement
function spindle.replace_escaped(template, level_one, level_two, replace)
	return template:gsub(level_one, function(x)
		return x:gsub(level_two, function(x) return string.format(replace, x) end)
	end)
end

function spindle.write_file(path, content)
	spindle.make_directory(path)
	local file = io.open(path, 'w')
	file:write(content)
	file:close()
end

function spindle.register_inline(func)
	table.insert(spindle.inlines, func)
end

function spindle.load_markup(file_path)
	if file_path == nil then
		return nil
	end

	local page = {}
	for k, v in pairs(spindle.tokens) do
		page[k] = v
	end

	if spindle.markup_cache[file_path] then
		page.syntax_tree = spindle.markup_cache[file_path]
	else
		local m = spindle.parse_markup(page, file_path)
		spindle.markup_cache[file_path] = m
		page.syntax_tree = m
	end

	return page
end

function spindle.load_page(file_path)
	if file_path == nil then
		return nil
	end

	if not spindle.file_exists(file_path) then
		return nil
	end

	if spindle.all_pages[file_path] then
		return spindle.all_pages[file_path]
	end

	local curl = spindle.url_from_path(file_path)
	local page = spindle.load_markup(file_path)

	page.canonical_url = curl
	page._source_path  = file_path
	page._output_path  = spindle.output_path .. file_path:gsub("%..+$", ".html")

	-- must be cached before pre-render to avoid loops
	spindle.all_pages[file_path] = page

	page.content = spindle.render(page, page.syntax_tree)
	if page.template then
		spindle.parse_markup = spindle.__parse_markup
		local template = spindle.load_markup("_data/" .. page.template .. ".x")
		spindle.render(page, template.syntax_tree)
	end

	return page
end

function spindle.load_file(file_path)
	local f = io.open(file_path)
	if f == nil then
		return ""
	end

	local s = f:read('*a')
	f:close()
	return s
end

function spindle.expand(template, page)
	-- variables: %var:modifiers
	local result = template:gsub("%%([%a_]+):([%a_]+)", function (x, y)
		local z = page[x]
		if z == nil then
			return x
		end
		if spindle.modifiers[y] then
			return spindle.modifiers[y](z)
		end
		return z
	end)

	-- variables: %var
	result = result:gsub("%%([%a_]+)", function (x)
		return page[x]
	end)

	-- spindle.inlines, like [links](#) or **bolded**
	for i, inline_exec in ipairs(spindle.inlines) do
		result = inline_exec(page, result)
	end

	return result
end

function spindle.template_expand(template, value)
	-- %placeholder:modifier
	local result = template:gsub("%%(%d+):([%w_]+)", function(x, y)
		local n = x + 0
		if n == 0 then
			x = value
		else
			x = spindle.split_quoted(value)[n]
		end
		if spindle.modifiers[y] then
			return spindle.modifiers[y](x)
		end
		return x
	end)

	-- %placeholder
	result = result:gsub("%%(%d+)", function(x)
		local n = x + 0
		if n == 0 then
			return value
		end
		return spindle.split_quoted(value)[n]
	end)

	return result
end

--[[function print_syntax_tree(level, syntax_tree)
	for i, entry in ipairs(syntax_tree) do
		print(level, i, entry.token or entry.block, entry.text or '')
		if entry.token and entry.block then
			print("    ", entry.token, entry.block)
		end
		if entry.block then
			print_syntax_tree(level + 1, entry)
		end
	end
end]]

function spindle.find_tokens(block, ...)
	local t = {}

	for _, entry in ipairs(block) do
		for _, v in ipairs {...} do
			if entry.token == v then
				table.insert(t, entry)
				break
			end
		end

		if entry.block then
			local n = spindle.find_tokens(entry, ...)
			for _, v in ipairs(n) do
				table.insert(t, v)
			end
			n = nil
		end
	end

	return t
end

function spindle.parse_markup(page, file_path)
	local syntax_tree = {}
	local active = syntax_tree
	local stack = {syntax_tree}

	local is_block_comment = false
	local is_block_raw     = false

	for line in io.lines(file_path) do
		local active = stack[#stack]

		if line == "" then
			if is_block_raw then
				active.text = active.text .. '\n'
				goto next_line
			end
			active[#active + 1] = { token = 'ws' }
			goto next_line
		end

		if line:match('^%s*}') then
			if not is_block_comment and not is_block_raw then
				stack[#stack] = nil
			end
			if is_block_comment then is_block_comment = false end
			if is_block_raw     then is_block_raw     = false end
			goto next_line
		end

		if is_block_comment then
			goto next_line
		end
		if is_block_raw then
			active.text = active.text .. '\n' .. line
			goto next_line
		end

		if line:match('^%s*/%s+') then
			if line:match('{$') then
				is_block_comment = true
			end
			goto next_line
		end

		-- check for variables
		local x = line:match('^%s*([%w_]+)%s*=')
		if x then
			local y = line:match("=%s*(.+)")

			if line:match("{$") then
				y = spindle.trim(y:sub(1, -2))

				new_block = {}
				new_block.token = x
				new_block.block = y

				active[#active + 1] = new_block
				stack[#stack + 1]   = new_block
				goto next_line
			end

			if y == 'false' then y = false end
			if y == 'true'  then y = true end

			if #stack == 1 then
				page[x] = y
			else
				active[x] = y
			end
			goto next_line
		end

		local ifs = line:match('^%s*if%s+%%(.-)%s+.-{$')
		if ifs then
			local block_id = line:match("%s+([%w_]+)%s+{$")
			local new_block = {
				token  = block_id or 'block_default',
				ifstmt = ifs,
				block  = true,
			}

			active[#active + 1] = new_block
			stack[#stack + 1]   = new_block
			goto next_line
		end

		-- open block
		local is_block = line:match('^%s*{$')
		if is_block then
			local new_block = {
				token = 'block_default',
				block = true
			}
			active[#active + 1] = new_block
			stack[#stack + 1]   = new_block
			goto next_line
		end

		local label = line:match('^%s*(.-)%s+{$')
		if label then
			local new_block = {
				token = label,
				block = true
			}
			active[#active + 1] = new_block
			stack[#stack + 1]   = new_block
			goto next_line
		end

		-- check for token/value pair
		local x, y = line:match('^%s*(%S+)%s+(.-)$')

		if page[x] then
			active[#active + 1] = {
				token = x,
				text  = spindle.trim(y)
			}
			goto next_line
		end

		line = spindle.trim(line)

		local char = string.sub(line, 1, 1)

		if char == '<' then
			active[#active + 1] = {
				token = 'html',
				text  = line
			}
		else
			active[#active + 1] = {
				token = 'default',
				text  = line
			}
		end

		::next_line::
	end

	return syntax_tree
end
spindle.__parse_markup = spindle.parse_markup -- here to assist with hacking

function spindle.render(page, active_block)
	local content = ""
	local index   = 0

	while true do
		index = index + 1
		if index > #active_block then
			break
		end

		local entry = active_block[index]

		if entry.token == "ws" then
			goto render_continue
		end

		if entry.token == "html" then
			content = content .. spindle.expand(entry.text, page)
			goto render_continue
		end

		if entry.block and entry.token and page[entry.token] then
			if entry.ifstmt then
				if not page[entry.ifstmt] then
					if index < #active_block and active_block[index + 1].elsestmt then
						entry = active_block[index + 1]
						index = index + 1
					else
						goto render_continue
					end
				end
			end

			local t = page[entry.token]
			local x = type(t)

			if x == 'string' then
				local w = spindle.expand(page[entry.token], page)
				content = content .. spindle.template_expand(w, spindle.render(page, entry))
			elseif x == 'function' then
				local v = t(page, spindle.render(page, entry))
				if v ~= nil then
					content = content .. (entry.raw and v or spindle.expand(v, page))
				end
			end

			goto render_continue
		end

		if entry.token and page[entry.token] then
			local t = page[entry.token]
			local x = type(t)

			if x == 'table' then
				local tok = entry.token
				local sub_content = ""
				local count = 0

				while true do
					local entry = active_block[index]
					if entry.token ~= tok then
						index = index - 1
						break
					end

					if entry.raw then
						sub_content = sub_content .. spindle.template_expand(t.main, entry.text)
					else
						sub_content = sub_content .. spindle.expand(spindle.template_expand(t.main, entry.text), page)
					end
					count = count + 1

					index = index + 1
					if index > #active_block then
						break
					end
				end

				if (not t.minimum) or t.minimum <= count then
					if entry.raw then
						content = content .. spindle.template_expand(t.wrap, sub_content)
					else
						content = content .. spindle.expand(spindle.template_expand(t.wrap, sub_content), page)
					end
				else
					content = content .. sub_content
				end

				goto render_continue

			elseif x == 'string' then
				if entry.raw then
					content = content .. spindle.template_expand(t, entry.text)
				else
					content = content .. spindle.expand(spindle.template_expand(t, entry.text), page)
				end
				goto render_continue

			elseif x == 'function' then
				local v = t(page, entry.raw and entry.text or spindle.expand(entry.text, page))
				if v ~= nil then
					content = content .. (entry.raw and v or spindle.expand(v, page))
				end
				goto render_continue
			end

			if entry.token == '$' then
				if _ENV[entry.text] then
					local v = _ENV[entry.text](page)
					if v ~= nil then
						content = content .. v
					end
					goto render_continue
				end

				local f = spindle.load_file("_data/" .. entry.text .. ".lua")
				local x = "local page = select(1, ...); " .. f
				local v = load(x)(page)
				if v ~= nil then
					content = content .. v
				end
				goto render_continue
			end

			if entry.token == '$$' then
				local x = "local page = select(1, ...); " .. entry.text
				local v = load(x)(page)
				if v ~= nil then
					content = content .. spindle.expand(v, page)
				end
			end
		end

		::render_continue::
	end

	return content
end

function spindle.export_file(path)
	local file_path = spindle.find_file(path)
	if file_path == nil then
		return false
	end

	spindle.copy_file(file_path, spindle.output_path .. file_path)
	return true
end

-- returns filepath + url
function spindle.process_any_file(path)
	if path == '/' then
		return path, spindle.domain
	end
	if path == '' then
		return nil, ""
	end

	if not spindle.has_protocol(path) then
		local file_path = spindle.find_file(path)

		if file_path == nil then
			return nil, ""
		end

		local ext = spindle.long_ext(file_path)

		if spindle.handlers[ext] then
			return spindle.handlers[ext](file_path)
		else
			return spindle.default_handler(file_path)
		end
	end

	return nil, path
end

function spindle.default_handler(file_path)
	local new_path = spindle.output_path .. file_path
	local file = {
		_source_path  = file_path,
		_output_path  = new_path,
		canonical_url = spindle.url_from_path(file_path)
	}
	spindle.all_files[file._output_path] = file
	return file._output_path, file.canonical_url
end

function spindle.has_run(path)
	if spindle.lock_file[path] == true then
		return true
	end

	spindle.lock_file[path] = true
	return false
end

function spindle.generate_sitemap(file_list)
	local sitemap = [[<?xml version="1.0" encoding="utf-8" standalone="yes"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:xhtml="http://www.w3.org/1999/xhtml">]]
	local schema  = [[<url><loc>%s</loc></url>]]

	for i, entry in ipairs(file_list) do
		file_list[i] = string.format(schema, entry)
	end

	return sitemap .. table.concat(file_list, "") .. [[</urlset>]]
end

function spindle.make_page_list()
	local file_list = {}

	for i, page in pairs(spindle.all_pages) do
		if not page._source_path:match("^404") then
			table.insert(file_list, page.canonical_url)
		end
	end

	for i, file in pairs(spindle.all_files) do
		if spindle.long_ext(file._output_path) == ".html" then
			table.insert(file_list, file.canonical_url)
		end
	end

	-- we do this extra sort so the output is idempotent
	table.sort(file_list)
	return file_list
end

function main(starting_file)
	if spindle.file_exists("_data/config.lua") then
		dofile("_data/config.lua")
	end

	spindle.process_any_file(starting_file)

	local file_list = {}

	for i, page in pairs(spindle.all_pages) do
		page.content = spindle.render(page, page.syntax_tree)

		if page.template then
			spindle.parse_markup = spindle.__parse_markup
			local template = spindle.load_markup("_data/" .. page.template .. ".x")
			spindle.write_file(page._output_path, spindle.render(page, template.syntax_tree))
		else
			spindle.write_file(page._output_path, page.content)
		end

		table.insert(file_list, page._source_path)
	end

	for i, file in pairs(spindle.all_files) do
		spindle.make_directory(file._output_path)
		if not spindle.copy_file(file._source_path, file._output_path) then
			print("failed to copy", file._source_path)
		end

		table.insert(file_list, file._source_path)
	end

	spindle.write_file(spindle.output_path .. "sitemap.xml", spindle.generate_sitemap(spindle.make_page_list()))
	table.insert(file_list, "sitemap.xml")

	table.sort(file_list)
	for _, item in ipairs(file_list) do
		print(item)
	end
end
