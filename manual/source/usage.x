& manual

title = Usage

$ toc

# Getting Started

Initialise a new project *within* an existing directory:

code raw {
	spindle init
}

# Serve

Launch a local server to work on the site in real-time, with hot reloading:

code raw {
	spindle serve
}

## Options

Most options are configured inside of `spindle.toml`, but there are a few overridable options at the command-line level.  In the event of overlap, the command line argument takes precedence, then `spindle.toml`, then the builtin default value.

- `-p [number]` or `--port [number]`: override the port number for the local development server.

# Build

Output a deployment-ready copy of the site to a `/public/` directory:

code raw {
	spindle build
}

## Options

Spindle makes use of a piece of syntax called a ['Resource Finder'](%{page syntax}#resource-finders): this streamlines the process of linking between pages — you can just put in the page's name, not its full path — and the resulting link styles can all be changed according to the needs of the site in `spindle.toml` (absolute, relative, etc.).

You can of course still manually supply links, but it's not recommended for the second reason: Resource Finders track the usage of any given asset or page, and by default `spindle build` will only output pages that are not orphaned.  The root index is assumed to be 'live' as the starting point, and anything reachable from the subsequent tree is included.

Practically speaking, if you commented out a link in your navigation bar, this would typically make a chunk of your site unreachable to the average user.  If the navigation bar uses Resource Finders to create its links, commenting an entry out will remove that entire branch of the site from the build output.

All of this is to explain this next flag:

- `-a` or `--all`: this forces Spindle to always render every asset in the source directory, ignoring their orphan status.

Spindle typically builds a site in just a few milliseconds, but it can also be directed to perform image conversions and resizing.  This process is by nature, *much slower*.

- `--skip-images` will ignore any images located by Resource Finders during the build, which is useful if you've just built the entire site and don't want to wait another thirty seconds for a full rebuild to fix a typo.