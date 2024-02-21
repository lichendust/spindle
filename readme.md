# 🧵 Spindle

Spindle is a static site generator.

It's principally based around the idea of a self-editing Markdown, where both templating and content are defined in the same space, allowing for infinite flexibility with the same basic structure as other tools: define some useful stuff with templating tools, then pull on it to format content.

Unlike Markdown, however, Spindle's syntax is extensible; you can add new tokens on the fly, define as many variations of elements as you like.

This new Lua version of Spindle is a partial, simplified rebuild of the previous Go version — it's significantly more powerful than it ever was, despite being 1/5th the size, but it also does less out of the box.  Also, this Lua version is very much not finished[^1] yet.

[^1]: As if this has ever meant anything.  I don't think any version of Spindle has ever been finished.
