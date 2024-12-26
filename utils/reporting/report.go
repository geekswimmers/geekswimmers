package reporting

import (
	"fmt"
	"log"
	"text/template"
)

func GetReportTemplate(name string) *template.Template {
	svg, err := template.ParseFiles(fmt.Sprintf("web/templates/reports/%s.svg", name))
	if err != nil {
		log.Print(err)
	}

	return svg
}
