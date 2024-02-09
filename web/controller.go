package web

import (
	"geekswimmers/storage"
	"geekswimmers/utils"
	"html/template"
	"net/http"
	"strconv"

	"log"
)

type WebController struct {
	DB                  storage.Database
	BaseTemplateContext BaseTemplateContext
}

func (wc *WebController) HomeView(res http.ResponseWriter, req *http.Request) {
	html := utils.GetTemplate("base", "home")
	err := html.Execute(res, wc.BaseTemplateContext)
	if err != nil {
		log.Print(err)
	}
}

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

func (wc *WebController) ActivateCookieSession(res http.ResponseWriter, req *http.Request) {
	if storage.SessionAvailable() {
		err := storage.AddSessionEntry(res, req, "profile", "acceptedCookies", "true")
		if err != nil {
			log.Printf("WebController: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
		}
		wc.BaseTemplateContext.AcceptedCookies, _ = strconv.ParseBool(storage.GetSessionValue(req, "profile", "acceptedCookies"))
		log.Printf("User accepted Cookies: %v", wc.BaseTemplateContext.AcceptedCookies)
	}

	res.WriteHeader(http.StatusAccepted)
}
