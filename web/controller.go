package web

import (
	"geekswimmers/content"
	"geekswimmers/storage"
	"geekswimmers/swimming"
	"geekswimmers/times"
	"geekswimmers/utils"
	"html/template"
	"net/http"

	"log"
)

type WebController struct {
	DB               storage.Database
	BaseTemplateData *utils.BaseTemplateData
}

func (wc *WebController) HomeView(res http.ResponseWriter, req *http.Request) {
	quoteOfTheDay, err := content.GetQuoteOfTheDay(utils.DayOfTheYear(), wc.DB)
	if err != nil {
		log.Printf("home.quoteOfTheDay.%v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	jurisdictions, err := times.FindJurisdictionsByLevel(times.JurisdictionLevelRegion, wc.DB)
	if err != nil {
		log.Printf("home.jurisdictions.%v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	events, err := swimming.FindEvents(wc.DB)
	if err != nil {
		log.Printf("home.events.%v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	articles, err := content.FindHighlightedArticles(wc.DB)
	if err != nil {
		log.Printf("home.Articles.%v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	updates, err := content.FindUpdates(wc.DB)
	if err != nil {
		log.Printf("home.Updates.%v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	sessionData := storage.NewSessionData(req)
	ctx := &homeViewData{
		QuoteOfTheDay:    quoteOfTheDay,
		Articles:         articles,
		Updates:          updates,
		Jurisdictions:    jurisdictions,
		Events:           events,
		BaseTemplateData: wc.BaseTemplateData,
		SessionData:      sessionData,
	}

	html := utils.GetTemplateWithFunctions("home", "home", template.FuncMap{
		"Title":    utils.Title,
		"markdown": utils.MarkdownToHTML,
	})

	err = html.Execute(res, ctx)
	if err != nil {
		log.Printf("web.HomeView: %v", err)
	}
}

func (wc *WebController) CrawlerView(res http.ResponseWriter, req *http.Request) {
	txt, err := template.ParseFiles("web/templates/robots.txt")
	if err != nil {
		log.Printf("html.template.ParseFiles: %v", err)
	}

	err = txt.Execute(res, nil)
	if err != nil {
		log.Printf("html.template.Template: %v", err)
	}
}

func (wc *WebController) SitemapView(res http.ResponseWriter, req *http.Request) {
	articles, err := content.FindArticlesExcept("", wc.DB)
	if err != nil {
		log.Printf("home.Articles.%v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	sessionData := storage.NewSessionData(req)
	ctx := &sitemapViewData{
		Articles:    articles,
		SessionData: sessionData,
	}

	txt, err := template.ParseFiles("web/templates/sitemap.xml")
	if err != nil {
		log.Printf("html.template.ParseFiles: %v", err)
	}

	err = txt.Execute(res, ctx)
	if err != nil {
		log.Printf("html.template.Template: %v", err)
	}
}

func (wc *WebController) NotFoundView(res http.ResponseWriter, req *http.Request) {
	sessionData := storage.NewSessionData(req)
	ctx := &notFoundViewData{
		BaseTemplateData: wc.BaseTemplateData,
		SessionData:      sessionData,
	}

	utils.ErrorHandler(res, req, ctx, http.StatusNotFound)
}

func (wc *WebController) ActivateCookieSession(res http.ResponseWriter, req *http.Request) {
	if storage.SessionStoreAvailable() {
		if err := storage.AddSessionEntry(res, req, "profile", "acceptedCookies", "true"); err != nil {
			log.Printf("storage.%v", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Printf("User accepted Cookies: %v", storage.GetSessionEntryValue(req, "profile", "acceptedCookies"))
	}

	res.WriteHeader(http.StatusAccepted)
}
