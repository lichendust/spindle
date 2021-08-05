# Spindle

Spindle is a static site builder based around repeatable templating and core web development principles — HTML, CSS and JavaScript.

## Philosophy

Spindle is designed to outlast framework races. This program (and indeed Golang itself) may not last forever, but another tool supporting Spindle's features could be written in a week.

Spindle or its descendants should be designed as portable, single binaries that can are fully accessible across multiple platforms. With this version being pure Go and refusing any dependencies on `cgo` or similar, it is eminently possible to compile and run Spindle on any modern operating system.

## Commands

### New

	spindle new

Spindle will generate a project file structure _in the current directory_.

### Serve

	spindle serve <path>

Spindle will start a local webserver and serve up pages from the application. It will automatically open the root index page in a browser.

The second argument can also be used to open a different page on startup, in the form `folder/page`, without any extension. This is a time-saving feature for quickly previewing the page the user is most interested in seeing, like the blog post they just wrote.

### Build

	spindle build <output_path>

Spindle will render or copy all content from `source` into a newly created `public` directory alongside `config` and `source`.

The output location can be overridden by an additional argument if necessary.

## Documentation

Full documentation will be available in the `/docs/` section of this repository.