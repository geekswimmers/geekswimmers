package server

import (
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

func CreateServer(db storage.Database, router *pat.PatternServeMux) *Server {
	s := &Server{}
	s.DB = db
	s.Router = router
	s.Routes()
	return s
}

func (s *Server) handleRequest(f Handler) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		f(res, req)
	}
}

func (s *Server) Routes() {
	webController := &web.WebController{
		DB: s.DB,
	}

	swimmersController := &swimmers.SwimmersController{
		DB: s.DB,
	}

	// The order here must be absolutely respected.
	s.Router = pat.New()
	s.Router.Get("/", s.handleRequest(webController.HomeView))
	s.Router.Post("/swimmers/benchmark", s.handleRequest(swimmersController.BenchmarkTime))

	s.Router.Get("/robots.txt", http.HandlerFunc(webController.CrawlerView))
	s.Router.Get("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("./web/static"))))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Router.ServeHTTP(w, r)
}
