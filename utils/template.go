package utils

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"regexp"

	"github.com/yuin/goldmark"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// This context is shared globally within the application. Do not put any session-specific data here.
type BaseTemplateContext struct {
	FeedbackForm              string
	MonitoringGoogleAnalytics string
}

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

func Title(str string) string {
	if len(str) == 0 {
		return str
	}

	return cases.Title(language.English, cases.Compact).String(str)
}

// ToHTML A template funcConvert markdown content to HTML and unescape special characters.
func ToHTML(s string) template.HTML {
	var html bytes.Buffer
	if err := goldmark.Convert([]byte(s), &html); err != nil {
		log.Printf("Error converting markdown to HTML: %v", err)
		return template.HTML(s)
	}

	// Adds target="_blank" to any URL generated from markdown.
	pattern := regexp.MustCompile(`(a href="[^"]+")`)
	htmlWithTargetedUrls := pattern.ReplaceAllString(html.String(), "${1} target=\"_blank\"")

	return template.HTML(htmlWithTargetedUrls)
}
