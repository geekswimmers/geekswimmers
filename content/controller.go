package content

import (
	"geekswimmers/storage"
	"geekswimmers/utils"
	"html/template"
	"log"
	"net/http"
)

type ContentController struct {
	DB                  storage.Database
	BaseTemplateContext *utils.BaseTemplateContext
}

type webContext struct {
	Article             *Article
	OtherArticles       []*Article
	BaseTemplateContext *utils.BaseTemplateContext
	AcceptedCookies     bool
}

func (wc *ContentController) ArticleView(res http.ResponseWriter, req *http.Request) {
	article, err := findArticle(req.URL.Query().Get(":reference"), wc.DB)
	if err != nil {
		log.Printf("Error viewing the article: %v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	otherArticles, err := FindArticlesExcept(article.Reference, wc.DB)
	if err != nil {
		log.Printf("Error viewing the articles: %v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	ctx := &webContext{
		Article:             article,
		OtherArticles:       otherArticles,
		BaseTemplateContext: wc.BaseTemplateContext,
		AcceptedCookies:     storage.GetSessionValue(req, "profile", "acceptedCookies") == "true",
	}

	html := utils.GetTemplateWithFunctions("base", "article", template.FuncMap{"markdown": utils.ToHTML})
	err = html.Execute(res, ctx)
	if err != nil {
		log.Print(err)
	}
}
