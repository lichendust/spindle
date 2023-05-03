# 🧵 Spindle

Spindle is a static site generator.

It's principally based around the idea of a self-editing Markdown, where both templating and content are defined in the same space, allowing for infinite flexibility with the same basic structure as other tools: define some useful stuff with templating tools, then pull on it to format content.

But what about a one-off edit?  What about a quick override here, or a page that's *slightly* different but doesn't merit a full template rebuild?  What if you could make a bunch of templating and then hack it up or override it or ignore it, as needed, at any level.

Spindle is a personal project.  It will never be the cleanest code-base and it's not *really* intended for widespread use, though it may get there at some point.  Its design choices are sometimes odd or poorly thought out.  It's an infinite learning project, designed as much to have a useful tool as it is to have a place to make mistakes in designing a complicated 'thing'.

In short, attempt to use it at your own peril.

<!-- MarkdownTOC autolink=true -->

- [Dependencies](#dependencies)
	- [cwebp](#cwebp)

<!-- /MarkdownTOC -->

## Dependencies

Spindle depends on a number of libraries, some of which are presently only necessary during development.  It also depends on some external utilities for specific processes.

### cwebp

While Spindle can *read* WEBP files (and indeed convert them to JPEG and PNG), it cannot output WEBP if the project calls for images to be converted or resized or re-encoded.

Spindle makes external calls to the `cwebp` executable [from Google](https://developers.google.com/speed/webp/download) to do this.  This isn't required out of the box: Spindle will also warn if any conversions have been requested and `cwebp` isn't found during builds.