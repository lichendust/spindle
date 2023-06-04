# Build Tools

All of Spindle's build tools are in this directory.  There aren't many and they're mostly time-savers rather than direct necessities.

Given a full copy of the source, Spindle can always be compiled using only —

	go build -ldflags "-s -w" -trimpath ./source

`-ldflags` strips out compiler cruft and debug symbols, halving the size of the binary.  Why this is not the default expression of `go build` is utterly maddening.

`-trimpath` makes sure the Go compiler doesn't leave *information about your personal filesystem on your computer in the binary for everyone to see in the event of a crash*.

## Usage

All of the following tools must be run with the working directory set as the root of the project, therefore called in the form —

	tool/build.sh

## Build

`build.sh` cross-compiles Spindle for several platforms, packages them with licenses and readmes, then outputs those packages' checksums ready for distribution.

## Embed

`embed.sh` takes the Spindle project in the `/manual/` directory *and* the `help_*` files from the `/text/` directory and embeds them into the Spindle codebase, where they act as the built-in user manual and the text for the `help` command respectively.

This is designed to allow consistent maintenance of the program's internal help messages, helping ensure that all presentation and formatting matches going forwards, and changes are easily visible outside the context of the code itself.

## Builtin IDs

`builtin_ids.go` generates Spindle's built-in identifiers, creating the hashes and embedding them as constants in the codebase for Spindle to make use of.