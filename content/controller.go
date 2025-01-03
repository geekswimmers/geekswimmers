package content

import (
	"geekswimmers/storage"
	"geekswimmers/utils"
	"html/template"
	"log"
	"net/http"
)

type ContentController struct {
	DB               storage.Database
	BaseTemplateData *utils.BaseTemplateData
}

func (wc *ContentController) ArticleView(res http.ResponseWriter, req *http.Request) {
	sessionData := storage.NewSessionData(req)

	ctx := &articleViewData{
		BaseTemplateData: wc.BaseTemplateData,
		SessionData:      sessionData,
	}

	reference := req.URL.Query().Get(":reference")
	article, err := getArticle(reference, wc.DB)

	if err != nil || article == nil {
		log.Printf("Error retrieving the article %s: %v", reference, err)
		utils.ErrorHandler(res, req, ctx, http.StatusNotFound)
		return
	}

	otherArticles, err := FindArticlesExcept(article.Reference, wc.DB)
	if err != nil {
		log.Printf("Error retrieving other articles: %v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	ctx.Article = article
	ctx.OtherArticles = otherArticles

	html := utils.GetTemplateWithFunctions("base", "article", template.FuncMap{"markdown": utils.MarkdownToHTML})
	err = html.Execute(res, ctx)
	if err != nil {
		log.Print(err)
	}
}
