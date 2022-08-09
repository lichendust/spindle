# Spindle

Spindle is a static site builder based around repeatable templating and core web development principles — HTML, CSS and JavaScript.

## Updates Coming!

⚠ Spindle is being rebuilt from the ground up on the `dev` branch.  A lot is changing.

## Philosophy

Spindle is designed to outlast framework races. This program (and indeed Golang itself) may not last forever, but another tool supporting Spindle's features could be written in a week.

Spindle or its descendants should be designed as portable, single binaries that can are fully accessible across multiple platforms. With this version being pure Go and refusing any dependencies on `cgo` or similar, it is eminently possible to compile and run Spindle on any modern operating system.

## Commands

### New

	spindle new

Spindle will generate a project file structure _in the current directory_.

### Serve

	spindle serve

Spindle will start a local webserver and serve up pages from the application. It will automatically open the root index page in a browser.

### Build

	spindle build <output_path>

Spindle will render or copy all content from `source` into a newly created `public` directory alongside `config` and `source`.

The optional argument allows for the output location to be overridden.

### Test ⚠

	spindle test <output_path>

Spindle will start a local webserver and serve the "public" build directory instead with no additional handling, allowing the final build to be tested locally without needing additional steps, tools or deployment.

The optional argument allows specifying a non-standard directory for output.

## Documentation

Full documentation will be available in the `/docs/` section of this repository.