package server

import (
	"geekswimmers/config"
	"geekswimmers/modules/content"
	"geekswimmers/modules/swimming"
	"geekswimmers/modules/times"
	"geekswimmers/modules/user"
	"geekswimmers/modules/web"
	"geekswimmers/storage"
	"geekswimmers/utils"
	"net/http"

	"github.com/bmizerany/pat"
)

type Server struct {
	DB     storage.Database
	Router *pat.PatternServeMux
}

type Handler func(res http.ResponseWriter, req *http.Request)

type AuthHandler func(res http.ResponseWriter, req *http.Request, session *storage.SessionData)

func CreateServer(c config.Config, db storage.Database) *Server {
	s := &Server{}
	s.DB = db
	s.Router = pat.New()

	btc := utils.BaseTemplateData{
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

func (s *Server) handleAuthRequest(f AuthHandler) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		sessionData := storage.NewSessionData(req)
		if !sessionData.IsAuthenticated() {
			http.Redirect(res, req, "/auth/signin/", http.StatusSeeOther)
			return
		}

		f(res, req, sessionData)
	}
}

func (s *Server) Routes(btc utils.BaseTemplateData) {
	userController := &user.Controller{
		DB:               s.DB,
		BaseTemplateData: &btc,
	}

	webController := &web.Controller{
		DB:               s.DB,
		BaseTemplateData: &btc,
	}

	contentController := &content.Controller{
		DB:               s.DB,
		BaseTemplateData: &btc,
	}

	benchmarkController := &times.BenchmarkController{
		DB:               s.DB,
		BaseTemplateData: &btc,
	}

	standardsController := &times.StandardsController{
		DB:               s.DB,
		BaseTemplateData: &btc,
	}

	recordsController := &times.RecordsController{
		DB:               s.DB,
		BaseTemplateData: &btc,
	}

	swimmingController := &swimming.Controller{
		DB:               s.DB,
		BaseTemplateData: &btc,
	}

	// The order here must be absolutely respected.
	// Web Content
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

	s.Router.Post("/profile/swimmers/form/search/", s.handleAuthRequest(userController.SwimmerFormSearch))
	s.Router.Post("/profile/swimmers/form/link/", s.handleAuthRequest(userController.SwimmerFormLink))
	s.Router.Get("/profile/swimmers/form/", s.handleAuthRequest(userController.SwimmerFormView))
	s.Router.Post("/profile/swimmers/form/", s.handleAuthRequest(userController.SwimmerForm))
	s.Router.Get("/profile/swimmers/:id/besttimes/form/", s.handleAuthRequest(userController.SwimmerBestTimeFormView))
	s.Router.Post("/profile/swimmers/:id/besttimes/form/", s.handleAuthRequest(userController.SwimmerBestTimeForm))
	s.Router.Get("/profile/swimmers/:id/besttimes/:bestId/form/", s.handleAuthRequest(userController.SwimmerBestTimeFormView))
	s.Router.Post("/profile/swimmers/:id/besttimes/:bestId/form/", s.handleAuthRequest(userController.SwimmerBestTimeForm))
	s.Router.Del("/profile/swimmers/:id/besttimes/:bestId/", s.handleAuthRequest(userController.SwimmerBestTimeDelete))
	s.Router.Get("/profile/swimmers/:id/besttimes/:bestId/", s.handleAuthRequest(userController.SwimmerBestTimeView))
	s.Router.Get("/profile/swimmers/:id/edit/", s.handleAuthRequest(userController.ProfileSwimmerFormView))
	s.Router.Post("/profile/swimmers/:id/edit/", s.handleAuthRequest(userController.ProfileSwimmerForm))
	s.Router.Get("/profile/swimmers/:id/", s.handleAuthRequest(userController.ProfileSwimmerView))
	s.Router.Del("/profile/swimmers/:id/", s.handleAuthRequest(userController.SwimmerDelete))
	s.Router.Get("/profile/swimmers/", s.handleAuthRequest(userController.ProfileSwimmerView))
	s.Router.Get("/profile/edit/", s.handleAuthRequest(userController.ProfileEditView))
	s.Router.Post("/profile/edit/", s.handleAuthRequest(userController.ProfileEditSave))
	s.Router.Get("/profile/", s.handleAuthRequest(userController.ProfileView))

	s.Router.Get("/content/articles/:reference/", s.handleRequest(contentController.ArticleView))

	s.Router.Get("/times/benchmark", s.handleRequest(benchmarkController.BenchmarkTime))
	s.Router.Get("/times/records/:id/history/", s.handleRequest(recordsController.RecordHistoryView))
	s.Router.Get("/times/records/:id/poster/", s.handleRequest(recordsController.RecordPosterView))
	s.Router.Get("/times/records/:id/", s.handleRequest(recordsController.RecordsView))
	s.Router.Get("/times/records", s.handleRequest(recordsController.RecordsListView))
	s.Router.Get("/times/standards/event/", s.handleRequest(standardsController.StandardsEventView))
	s.Router.Get("/times/standards/:id/", s.handleRequest(standardsController.TimeStandardView))
	s.Router.Get("/times/standards", s.handleRequest(standardsController.TimeStandardsView))

	s.Router.Get("/swimming/styles", s.handleRequest(swimmingController.SwimStylesView))
	s.Router.Get("/swimming/styles/:stroke/", s.handleRequest(swimmingController.SwimStyleView))

	s.Router.Get("/robots.txt", http.HandlerFunc(webController.CrawlerView))
	s.Router.Get("/sitemap.xml", http.HandlerFunc(webController.SitemapView))
	s.Router.Get("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("./web/static"))))

	// BFF API
	s.Router.Get("/api/swimmers/:id/events/", s.handleRequest(userController.EventsResource))
	s.Router.Get("/api/clubs/", s.handleRequest(swimmingController.ClubResource))
	s.Router.Put("/api/profile/swimmers/:id/parentlink/:linkId/", s.handleAuthRequest(userController.AcceptParentLink))
	s.Router.Del("/api/profile/swimmers/:id/parentlink/:linkId/", s.handleAuthRequest(userController.DismissParentLink))

	s.Router.NotFound = http.HandlerFunc(webController.NotFoundView)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Router.ServeHTTP(w, r)
}
