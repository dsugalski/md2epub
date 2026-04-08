package main

import (
	"flag"
	"log"
	"os"
	"regexp"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/microcosm-cc/bluemonday"
	"github.com/writingtoole/epub"
)

var (
	version  = flag.Int("version", 3, "EPub version to generate, 2 or 3 is valid. (3 is default)")
	output   = flag.String("output", "markdown.epub", "Name of the output file")
	input    = flag.String("input", "markdown.md", ".md file to read")
	author   = flag.String("author", "", "Author, if any")
	title    = flag.String("title", "", "Title, if any")
	language = flag.String("language", "en", "Doc language, en by default")
)

// sub performs a global search and replace, because Go's regex
// package is kind of sucky.
func sub(in, regex, repl string) string {
	r := regexp.MustCompile(regex)
	return r.ReplaceAllString(in, repl)
}

func wrapXHTML(title, body string) string {
	title = sub(title, "<[^>]*>", "")
	doc := `<?xml version='1.0' encoding='utf-8'?>
	<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.1//EN" "http://www.w3.org/TR/xhtml11/DTD/xhtml11.dtd">
	<html xmlns="http://www.w3.org/1999/xhtml">
	<head>
	<title>
	`
	doc += title
	doc += `
</title>
</head>    
  <body>
`
	doc += body
	doc += `    </body>
</html>
`

	return doc
}

func mdToHTML(md []byte) string {
	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank | html.UseXHTML
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	rawHtml := markdown.Render(doc, renderer)
	html := bluemonday.UGCPolicy().SanitizeBytes(rawHtml)
	return wrapXHTML(*title, string(html))
}

func main() {
	flag.Parse()

	md, err := os.ReadFile(*input)
	if err != nil {
		log.Fatalf("Can't read input file %v: %v", *input, err)
	}

	html := mdToHTML([]byte(md))
	book := epub.New()
	err = book.SetVersion(float64(*version))
	if err != nil {
		log.Fatalf("Unable to set version: %v", err)
	}

	if *author != "" {
		book.AddAuthor(*author)
	}
	if *title != "" {
		book.SetTitle(*title)
	}

	book.AddLanguage(*language)

	_, err = book.AddXHTML("markdown.xhtml", string(html))
	if err != nil {
		log.Fatalf("Unable to add markdown file: %v", err)
	}

	book.AddNavpoint("start", "markdown.xhtml", 1)

	err = book.Write(*output)
	if err != nil {
		log.Fatalf("Unable to write epub: %v", err)
	}
}
