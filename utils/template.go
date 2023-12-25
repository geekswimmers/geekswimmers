package utils

import (
	"fmt"
	"html/template"
	"log"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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

func Title(str string) string {
	if len(str) == 0 {
		return str
	}

	return cases.Title(language.English, cases.Compact).String(str)
}
