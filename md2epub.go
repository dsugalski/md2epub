package main

import (
	"flag"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/dsugalski/md2epub/walk"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/microcosm-cc/bluemonday"
	"github.com/writingtoole/epub"
)

var (
	version  = flag.Int("version", 3, "EPub version to generate, 2 or 3 is valid. (3 is default)")
	output   = flag.String("output", "markdown.epub", "Name of the output file")
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

func parseMd(md []byte) ast.Node {
	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	return doc
}

func astToHTML(doc ast.Node) string {
	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank | html.UseXHTML
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	rawHtml := markdown.Render(doc, renderer)
	cookedHtml := bluemonday.UGCPolicy().SanitizeBytes(rawHtml)
	return wrapXHTML(*title, string(cookedHtml))
}

func main() {
	flag.Parse()

	files := flag.Args()
	if len(files) == 0 {
		log.Fatalf("No input files specified")
	}

	// set up the empty book file.
	book := epub.New()
	err := book.SetVersion(float64(*version))
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

	// Run through the files on the command line
	for _, f := range files {

		if !strings.HasSuffix(f, ".md") {
			log.Fatalf("%v is not a markdown file")
		}
		md, err := os.ReadFile(f)
		if err != nil {
			log.Fatalf("Can't read input file %v: %v", f, err)
		}

		ast := parseMd(md)
		//images := walk.ExtractImageFiles(ast)

		// Make sure the headers have unique IDs
		walk.AssignCustomHeaderIDs(ast)

		// Get the headers for later TOC creation
		headers := walk.ExtractHeaderIDs(ast)

		html := astToHTML(ast)

		xmlFilename := strings.TrimSuffix(f, ".md") + ".xhtml"

		_, err = book.AddXHTML(xmlFilename, string(html))
		if err != nil {
			log.Fatalf("Unable to add markdown file: %v", err)
		}

		for _, h := range headers {
			if h.Level != 1 {
				continue
			}
			book.AddNavpoint(h.Text, xmlFilename+"#"+h.Link, 0)
		}
	}

	err = book.Write(*output)
	if err != nil {
		log.Fatalf("Unable to write epub: %v", err)
	}
}
