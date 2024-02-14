package web

import (
	"geekswimmers/content"
	"geekswimmers/storage"
	"geekswimmers/utils"
	"html/template"
	"net/http"

	"log"
)

type WebController struct {
	DB                  storage.Database
	BaseTemplateContext *utils.BaseTemplateContext
}

type webContext struct {
	Articles            []*content.Article
	BirthDate           string
	Gender              string
	BaseTemplateContext *utils.BaseTemplateContext
	AcceptedCookies     bool
}

func (wc *WebController) HomeView(res http.ResponseWriter, req *http.Request) {
	birthDate := storage.GetSessionValue(req, "profile", "birthDate")
	gender := storage.GetSessionValue(req, "profile", "gender")
	articles, err := content.FindHighlightedArticles(wc.DB)
	if err != nil {
		log.Printf("Error viewing the articles: %v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	ctx := &webContext{
		Articles:            articles,
		BirthDate:           birthDate,
		Gender:              gender,
		BaseTemplateContext: wc.BaseTemplateContext,
		AcceptedCookies:     storage.GetSessionValue(req, "profile", "acceptedCookies") == "true",
	}

	html := utils.GetTemplateWithFunctions("base", "home", template.FuncMap{"markdown": utils.ToHTML})
	err = html.Execute(res, ctx)
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

	ctx := &webContext{
		BaseTemplateContext: wc.BaseTemplateContext,
		AcceptedCookies:     true,
	}

	template := utils.GetTemplate("base", "not-found")
	err := template.Execute(res, ctx)
	if err != nil {
		log.Print(err)
	}
}

func (wc *WebController) ActivateCookieSession(res http.ResponseWriter, req *http.Request) {
	if storage.SessionAvailable() {
		err := storage.AddSessionEntry(res, req, "profile", "acceptedCookies", "true")
		if err != nil {
			log.Printf("web.ActivateCookieSession: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
		}
		log.Printf("User accepted Cookies: %v", storage.GetSessionValue(req, "profile", "acceptedCookies"))
	}

	res.WriteHeader(http.StatusAccepted)
}
