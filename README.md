# md2epub

A quick little utility that turns a markdown file into an ePub format book

This tool is small and not particularly smart. It currently can't
handle multiple markdown files at once, or include image files in the
generated book. Yet.

The generated epub file has no table of contents, navpoints, or other
niceties. Maybe tomorrow.

## Building

`go build`

Really simple. Any version of Go should work, though the module file
specifies whatever was current when I made this thing.

## usage

`md2epub --output=mybook.epub --author=me --title="My thing" <list of .md files>`

Usage subject to change, this is kind of awkward.

## flags

### output

Name of the output file to write, `markdown.epub` by default.

### input

Name of the input file. None by default.

### title

Title of the epub book to create.

### version

Version of epub to write. Can be 2 (which is old, 2007) or 3 (which is
the current version, as of 2010). Defaults to 3, but can be set to 2
if you have a *really* antique ereader.
