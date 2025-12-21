package content

import (
	"fmt"
	"geekswimmers/modules"
	"geekswimmers/storage"
	"geekswimmers/utils"
	"html/template"
	"log"
	"net/http"
)

type Controller struct {
	DB               storage.Database
	BaseTemplateData *utils.BaseTemplateData
}

func (wc *Controller) BlogView(res http.ResponseWriter, req *http.Request) {
	ctx := modules.InitializeRequestContext(storage.NewSessionData(req), wc.BaseTemplateData)

	highlightedArticles, err := FindHighlightedArticles(wc.DB)
	if err != nil {
		log.Printf("content.Blog.%v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	highlightedArticle := &Article{
		Reference: "",
	}
	if len(highlightedArticles) > 0 {
		highlightedArticle = highlightedArticles[0]
	}

	articles, err := FindArticlesExcept(highlightedArticle.Reference, wc.DB)
	if err != nil {
		log.Printf("content.Blog.%v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	ctx["Articles"] = articles
	ctx["Highlighted"] = highlightedArticle

	html := utils.GetTemplateWithFunctions("base", "blog", template.FuncMap{
		"Title":    utils.Title,
		"markdown": utils.MarkdownToHTML,
	})

	err = html.Execute(res, ctx)
	if err != nil {
		log.Printf("web.BlogView: %v", err)
	}
}

func (wc *Controller) ArticleView(res http.ResponseWriter, req *http.Request) {
	ctx := modules.InitializeRequestContext(storage.NewSessionData(req), wc.BaseTemplateData)

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

	ctx["Article"] = article
	ctx["OtherArticles"] = otherArticles

	html := utils.GetTemplateWithFunctions("base", "article", template.FuncMap{"markdown": utils.MarkdownToHTML})
	err = html.Execute(res, ctx)
	if err != nil {
		log.Print(err)
	}
}

func (wc *Controller) ArticleRedirectView(res http.ResponseWriter, req *http.Request) {
	reference := req.URL.Query().Get(":reference")

	http.Redirect(res, req, fmt.Sprintf("/content/blog/%s/", reference), http.StatusMovedPermanently)
}
