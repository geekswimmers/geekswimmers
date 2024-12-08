package server

import (
	"geekswimmers/config"
	"geekswimmers/content"
	"geekswimmers/storage"
	"geekswimmers/swimming"
	"geekswimmers/times"
	"geekswimmers/user"
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
	userController := &user.UserController{
		DB:                  s.DB,
		BaseTemplateContext: &btc,
	}

	webController := &web.WebController{
		DB:                  s.DB,
		BaseTemplateContext: &btc,
	}

	contentController := &content.ContentController{
		DB:                  s.DB,
		BaseTemplateContext: &btc,
	}

	benchmarkController := &times.BenchmarkController{
		DB:                  s.DB,
		BaseTemplateContext: &btc,
	}

	standardsController := &times.StandardsController{
		DB:                  s.DB,
		BaseTemplateContext: &btc,
	}

	recordsController := &times.RecordsController{
		DB:                  s.DB,
		BaseTemplateContext: &btc,
	}

	meetController := &swimming.MeetController{
		DB:                  s.DB,
		BaseTemplateContext: &btc,
	}

	// The order here must be absolutely respected.
	s.Router = pat.New()
	s.Router.Get("/", s.handleRequest(webController.HomeView))
	s.Router.Get("/api/accepted-cookies", s.handleRequest(webController.ActivateCookieSession))

	s.Router.Get("/signup/", http.HandlerFunc(userController.SignUpView))
	s.Router.Post("/signup/", s.handleRequest(userController.SignUp))

	s.Router.Get("/auth/confirm/:confirmation", s.handleRequest(userController.PasswordView))
	s.Router.Get("/auth/password/reset/", http.HandlerFunc(userController.ResetPasswordView))
	s.Router.Post("/auth/password/reset/", s.handleRequest(userController.ResetPassword))
	s.Router.Post("/auth/password/", s.handleRequest(userController.SetNewPassword))

	s.Router.Get("/auth/signout/", http.HandlerFunc(userController.SignOut))
	s.Router.Get("/auth/signin/", http.HandlerFunc(userController.SignInView))
	s.Router.Post("/auth/signin/", s.handleRequest(userController.SignIn))

	s.Router.Get("/content/articles/:reference/", s.handleRequest(contentController.ArticleView))

	s.Router.Get("/times/benchmark", s.handleRequest(benchmarkController.BenchmarkTime))
	s.Router.Get("/times/records/:id/history/", s.handleRequest(recordsController.RecordHistoryView))
	s.Router.Get("/times/records/:id/", s.handleRequest(recordsController.RecordsView))
	s.Router.Get("/times/records", s.handleRequest(recordsController.RecordsListView))
	s.Router.Get("/times/standards/event/", s.handleRequest(standardsController.StandardsEventView))
	s.Router.Get("/times/standards/:id/", s.handleRequest(standardsController.TimeStandardView))
	s.Router.Get("/times/standards", s.handleRequest(standardsController.TimeStandardsView))

	s.Router.Get("/swimming/styles", s.handleRequest(meetController.SwimStylesView))
	s.Router.Get("/swimming/styles/:stroke/", s.handleRequest(meetController.SwimStyleView))

	s.Router.Get("/robots.txt", http.HandlerFunc(webController.CrawlerView))
	s.Router.Get("/sitemap.xml", http.HandlerFunc(webController.SitemapView))
	s.Router.Get("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("./web/static"))))

	s.Router.NotFound = http.HandlerFunc(webController.NotFoundView)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Router.ServeHTTP(w, r)
}
