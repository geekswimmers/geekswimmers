package utils

import (
	"bytes"
	"fmt"
	"github.com/gomarkdown/markdown"
	"html/template"
	"log"
	"regexp"
)

func GetTemplate(layout, page string) *template.Template {
	html, err := template.New(fmt.Sprintf("page.%s.html", page)).ParseFiles(applyLayout(layout, page)...)
	if err != nil {
		log.Fatalf("GetTemplate(%v, %v): %v", layout, page, err)
	}

	return html
}

func GetTemplateWithFunctions(layout, page string, funcMap template.FuncMap) *template.Template {
	html, err := template.New(fmt.Sprintf("page.%s.html", page)).Funcs(funcMap).
		ParseFiles(applyLayout(layout, page)...)
	if err != nil {
		log.Fatalf("GetTemplate(%v, %v): %v", layout, page, err)
	}

	return html
}

func applyLayout(layout, page string) []string {
	files := []string{
		fmt.Sprintf("web/templates/page.%s.html", page),
		fmt.Sprintf("web/templates/layout.%s.html", layout),
		"web/templates/partial.footer.html",
	}

	return files
}

// ToMarkdown A template funcConvert markdown content to HTML and unescape special characters.
func ToMarkdown(s string) template.HTML {
	md := normalizeNewlines([]byte(s))
	html := markdown.ToHTML(md, nil, nil)

	// Adds target="_blank" to any URL generated from markdown.
	pattern := regexp.MustCompile(`(a href="[^"]+")`)
	htmlWithTargetedUrls := pattern.ReplaceAllString(string(html), "${1} target=\"_blank\"")

	return template.HTML(htmlWithTargetedUrls)
}

func ToHTML(s string) template.HTML {
	return template.HTML(s)
}

func normalizeNewlines(d []byte) []byte {
	// replace CR(13) LF(10) (windows) with LF(10) (unix)
	d = bytes.Replace(d, []byte{13, 10}, []byte{10}, -1)

	// replace CF (mac) with LF (unix)
	d = bytes.Replace(d, []byte{13}, []byte{10}, -1)
	return d
}
