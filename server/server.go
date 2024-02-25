package server

import (
	"geekswimmers/config"
	"geekswimmers/content"
	"geekswimmers/storage"
	"geekswimmers/times"
	"geekswimmers/utils"
	"geekswimmers/web"
	"net/http"

	"github.com/bmizerany/pat"
)

type Server struct {
	DB     storage.Database
	Router *pat.PatternServeMux
}

type Handler func(res http.ResponseWriter, req *http.Request)

func CreateServer(c config.Config, db storage.Database) *Server {
	s := &Server{}
	s.DB = db
	s.Router = pat.New()

	btc := utils.BaseTemplateContext{
		FeedbackForm:              c.GetString(config.FeedbackForm),
		MonitoringGoogleAnalytics: c.GetString(config.MonitoringGoogleAnalytics),
	}
	s.Routes(btc)
	return s
}

func (s *Server) handleRequest(f Handler) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		f(res, req)
	}
}

func (s *Server) Routes(btc utils.BaseTemplateContext) {
	webController := &web.WebController{
		DB:                  s.DB,
		BaseTemplateContext: &btc,
	}

	contentController := &content.ContentController{
		DB:                  s.DB,
		BaseTemplateContext: &btc,
	}

	swimmersController := &times.SwimmersController{
		DB:                  s.DB,
		BaseTemplateContext: &btc,
	}

	// The order here must be absolutely respected.
	s.Router = pat.New()
	s.Router.Get("/", s.handleRequest(webController.HomeView))
	s.Router.Get("/api/accepted-cookies", s.handleRequest(webController.ActivateCookieSession))

	s.Router.Get("/content/articles/:reference/", s.handleRequest(contentController.ArticleView))

	s.Router.Get("/times/benchmark", s.handleRequest(swimmersController.BenchmarkTime))
	s.Router.Get("/times/standards/event/", s.handleRequest(swimmersController.StandardsEventView))
	s.Router.Get("/times/standards/:id/", s.handleRequest(swimmersController.TimeStandardView))
	s.Router.Get("/times/standards", s.handleRequest(swimmersController.TimeStandardsView))

	s.Router.Get("/robots.txt", http.HandlerFunc(webController.CrawlerView))
	s.Router.Get("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("./web/static"))))

	s.Router.NotFound = http.HandlerFunc(webController.NotFoundView)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Router.ServeHTTP(w, r)
}
