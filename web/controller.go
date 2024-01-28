package web

import (
	"geekswimmers/storage"
	"geekswimmers/utils"
	"html/template"
	"net/http"

	"log"
)

type WebController struct {
	DB                  storage.Database
	BaseTemplateContext BaseTemplateContext
}

// HomeView
// get: /
func (wc *WebController) HomeView(res http.ResponseWriter, req *http.Request) {
	html := utils.GetTemplate("base", "home")
	err := html.Execute(res, wc.BaseTemplateContext)
	if err != nil {
		log.Print(err)
	}
}

// CrawlerView
// get: /robots.txt
func (wc *WebController) CrawlerView(res http.ResponseWriter, req *http.Request) {
	txt, err := template.ParseFiles("web/templates/robots.txt")
	if err != nil {
		log.Print(err)
	}

	err = txt.Execute(res, nil)
	if err != nil {
		log.Print(err)
	}
}

func (wc *WebController) NotFoundView(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(http.StatusNotFound)

	template := utils.GetTemplate("base", "not-found")
	err := template.Execute(res, nil)
	if err != nil {
		log.Print(err)
	}
}
