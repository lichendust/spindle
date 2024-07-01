# 🧵 Spindle

Spindle is a self-contained static-site generator.

It emerged from the desire for a kind of self-editing Markdown — what if you could redefine what `#` did on the fly? Spindle executes on this idea, with templating and content being defined in the same space and markup language, rather than several distinct spaces and structural paradigms.

Spindle's syntax is extensible; you can add new tokens on the fly, define as many variations of elements as you like and built consistent, reusable templating that can even change based on context.

Spindle is designed to live in the non-existent no man's land between the ultra-minimal, like Brian Crabtree's clever but constrained [sh](https://nnnnnnnn.co/sh.html), and the ultra-complex, like the very powerful but inflexible [Hugo](https://gohugo.io).  It also wants to do this while being *way cooler than any of them*.

> [!CAUTION]
> Spindle is a bit rough right now. It works, but it's designed for people who are okay getting a bit *nasty*.

## Architecture

Spindle is a portable executable which hosts a Lua script. All of Spindle's core logic is executed by that script — the underlying program contains a Lua runtime and provides some additional functions; mostly those relating to filesystem handling that Lua isn't capable of performing natively.

Spindle's Lua construction allows for total modularity. Components and functions may be overloaded and modified at any time. In addition to regular templating, Lua can be executed inline to allow for any kind of complex structure.

Spindle's core script can also be vendored into the project itself, to be hacked and modified as needed.

Certain components are actually designed with 'modding support' in mind — for instance, the Spindle markup parser, which accepts a string and returns a syntax tree, is protected internally so that you can overload it with your own and always recover it. This means you can take any input format you like and, as long as you can conform it to spit out Spindle's syntax tree, build any number of interfaces to different markups and languages without needing to make significant changes to the core.

> My own [website](https://lichendust.com) is built with Spindle and uses a very simple drop-in Markdown parser I wrote in an hour or so to render notes from my Obsidian workspace. In addition to most of Markdown (at least the bits I use), it also supports wikilinks and tables.
>
> I use this parser overload trick so that I can simply load Markdown files *exactly* as I load native Spindle files, with two additional lines of code on either side.
