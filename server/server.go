package server

import (
	"geekswimmers/config"
	"geekswimmers/storage"
	"geekswimmers/swimmers"
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
	sessionAvailable := storage.SessionAvailable()

	btc := web.BaseTemplateContext{
		FeedbackForm:              c.GetString(config.FeedbackForm),
		MonitoringGoogleAnalytics: c.GetString(config.MonitoringGoogleAnalytics),
		AcceptedCookies:           !sessionAvailable,
	}
	s.Routes(btc)
	return s
}

func (s *Server) handleRequest(f Handler) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		f(res, req)
	}
}

func (s *Server) Routes(btc web.BaseTemplateContext) {
	webController := &web.WebController{
		DB:                  s.DB,
		BaseTemplateContext: btc,
	}

	swimmersController := &swimmers.SwimmersController{
		DB:                  s.DB,
		BaseTemplateContext: btc,
	}

	// The order here must be absolutely respected.
	s.Router = pat.New()
	s.Router.Get("/", s.handleRequest(webController.HomeView))
	s.Router.Get("/api/accepted-cookies", s.handleRequest(webController.ActivateCookieSession))

	s.Router.Post("/swimmers/benchmark", s.handleRequest(swimmersController.BenchmarkTime))

	s.Router.Get("/robots.txt", http.HandlerFunc(webController.CrawlerView))
	s.Router.Get("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("./web/static"))))

	s.Router.NotFound = http.HandlerFunc(webController.NotFoundView)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Router.ServeHTTP(w, r)
}
