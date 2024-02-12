package content

import (
	"geekswimmers/storage"
	"geekswimmers/utils"
	"log"
	"net/http"
)

type ContentController struct {
	DB                  storage.Database
	BaseTemplateContext *utils.BaseTemplateContext
}

type webContext struct {
	Article             Article
	BaseTemplateContext *utils.BaseTemplateContext
	AcceptedCookies     bool
}

func (wc *ContentController) ArticleView(res http.ResponseWriter, req *http.Request) {
	article, err := findArticle(req.URL.Query().Get(":reference"), wc.DB)
	if err != nil {
		log.Printf("Error viewing the article: %v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	ctx := &webContext{
		Article:             *article,
		BaseTemplateContext: wc.BaseTemplateContext,
		AcceptedCookies:     storage.GetSessionValue(req, "profile", "acceptedCookies") == "true",
	}

	html := utils.GetTemplate("base", "article")
	err = html.Execute(res, ctx)
	if err != nil {
		log.Print(err)
	}
}
