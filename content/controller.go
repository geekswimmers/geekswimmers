package content

import (
	"geekswimmers/storage"
	"geekswimmers/utils"
	"geekswimmers/web"
	"log"
	"net/http"
)

type ContentController struct {
	DB                  storage.Database
	BaseTemplateContext web.BaseTemplateContext
}

func (wc *ContentController) ArticlesView(res http.ResponseWriter, req *http.Request) {
	articles, err := findArticles(wc.DB)
	if err != nil {
		log.Printf("Error viewing the article: %v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	html := utils.GetTemplate("base", "articles")
	err = html.Execute(res, nil)
	if err != nil {
		log.Print(err)
	}
}

func (wc *ContentController) ArticleView(res http.ResponseWriter, req *http.Request) {
	article, err := findArticle(req.URL.Query().Get(":reference"), wc.DB)
	if err != nil {
		log.Printf("Error viewing the article: %v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	html := utils.GetTemplate("base", "home")
	err = html.Execute(res, nil)
	if err != nil {
		log.Print(err)
	}
}
