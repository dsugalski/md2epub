// Package walk has code to walk the AST and extract out various things.
package walk

import (
	"fmt"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/microcosm-cc/bluemonday"
)

// ExtractImageFiles walks a tree and for each image node it pulls out
// the linked-to image files. We'll use this later to go remap image
// file names.
func ExtractImageFiles(doc ast.Node) []string {
	ret := []string{}

	ast.WalkFunc(doc, func(node ast.Node, entering bool) ast.WalkStatus {
		if img, ok := node.(*ast.Image); ok && entering {
			ret = append(ret, string(img.Destination))
		}
		return ast.GoToNext
	})

	return ret
}

// RemapImageFiles takes an AST and a filename remapping map. Any
// image node in the doc with a name we can find in the map as a key
// gets remapped to the value of that entry.
//
// The point here is we're sticking the actual image files into the
// epub, but they'll have different names and we need to update the
// AST to note the new name.
func RemapImageFiles(doc ast.Node, remap map[string]string) {
	ast.WalkFunc(doc, func(node ast.Node, entering bool) ast.WalkStatus {
		if img, ok := node.(*ast.Image); ok && entering {
			if replace, ok := remap[string(img.Destination)]; ok {
				img.Destination = []byte(replace)
			}
		}
		return ast.GoToNext
	})
}

// AssignCustomHeaderIDs makes sure every header entry has an ID
// assigned to it. We'll need this later.
func AssignCustomHeaderIDs(doc ast.Node) {
	counter := 1
	ast.WalkFunc(doc, func(node ast.Node, entering bool) ast.WalkStatus {
		if h, ok := node.(*ast.Heading); ok && entering {
			if h.HeadingID == "" {
				h.HeadingID = fmt.Sprintf("header-level%d-%d", h.Level, counter)
				counter++
			}
		}
		return ast.GoToNext
	})
}

type Headers struct {
	Level int
	Text  string
	Link  string
}

func ExtractHeaderIDs(doc ast.Node) []Headers {
	ret := []Headers{}
	ast.WalkFunc(doc, func(node ast.Node, entering bool) ast.WalkStatus {
		if h, ok := node.(*ast.Heading); ok && entering {
			// Snag the contents of the header node's child, since we need
			// to use this for table of contents and navpoint stuff.
			htmlFlags := html.CommonFlags | html.HrefTargetBlank | html.UseXHTML
			opts := html.RendererOptions{Flags: htmlFlags}
			r := html.NewRenderer(opts)
			rawHtml := markdown.Render(h.Children[0], r)
			cookedHtml := string(bluemonday.UGCPolicy().SanitizeBytes(rawHtml))
			ret = append(ret, Headers{h.Level, cookedHtml, h.HeadingID})
		}
		return ast.GoToNext
	})
	return ret
}
