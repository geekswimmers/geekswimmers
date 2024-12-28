package reporting

import (
	"fmt"
	"geekswimmers/utils"
	"log"
	"text/template"
)

func GetReportTemplate(name string) *template.Template {
	svg, err := template.New(fmt.Sprintf("%s.svg", name)).Funcs(template.FuncMap{
		"FormatMiliseconds": utils.FormatMiliseconds,
	}).ParseFiles(fmt.Sprintf("web/templates/reports/%s.svg", name))
	if err != nil {
		log.Print(err)
	}

	return svg
}
