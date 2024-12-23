package utils

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// This context is shared globally within the application. Do not put any session-specific data here.
type BaseTemplateData struct {
	Email                     string
	FeedbackForm              string
	MonitoringGoogleAnalytics string
}

func GetTemplate(layout, page string) *template.Template {
	html, err := template.New(fmt.Sprintf("page.%s.html", page)).ParseFiles(applyLayout(layout, page)...)
	if err != nil {
		log.Fatalf("utils.GetTemplate(%v, %v): %v", layout, page, err)
	}

	return html
}

func GetTemplateWithFunctions(layout, page string, funcMap template.FuncMap) *template.Template {
	html, err := template.
		New(fmt.Sprintf("page.%s.html", page)).
		Funcs(funcMap).
		ParseFiles(applyLayout(layout, page)...)
	if err != nil {
		log.Fatalf("utils.GetTemplateWithFunctions(%v, %v): %v", layout, page, err)
	}

	return html
}

func applyLayout(layout, page string) []string {
	files := []string{
		fmt.Sprintf("web/templates/page.%s.html", page),
		fmt.Sprintf("web/templates/layout.%s.html", layout),
	}

	return files
}

func Title(str string) string {
	if len(str) == 0 {
		return str
	}

	words := strings.Split(str, "_")
	separator := ""
	var title string
	for _, word := range words {
		title += separator + cases.Title(language.English, cases.Compact).String(word)
		separator = " "
	}

	return title
}

func Lowercase(str string) string {
	return strings.ToLower(str)
}

func ToHTML(s string) template.HTML {
	return template.HTML(s)
}

// MarkdownToHTML Given a markdown content, converts it to HTML and unescape special characters.
func MarkdownToHTML(s string) template.HTML {
	var html bytes.Buffer
	if err := goldmark.Convert([]byte(s), &html); err != nil {
		log.Printf("Error converting markdown to HTML: %v", err)
		return template.HTML(s)
	}

	// Adds target="_blank" to all URLs
	pattern := regexp.MustCompile(`(a href="[^"]+")`)
	htmlWithTargetedUrls := pattern.ReplaceAllString(html.String(), "${1} target=\"_blank\" rel=\"noopener noreferrer\"")

	// Makes all images responsive.
	pattern = regexp.MustCompile(`(img src="[^"]+")`)
	htmlWithTargetedUrls = pattern.ReplaceAllString(htmlWithTargetedUrls, "${1} class=\"img-fluid\"")

	return template.HTML(htmlWithTargetedUrls)
}
