package server

import (
	"geekswimmers/auth"
	"geekswimmers/storage"
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
	authController := &auth.AuthController{
		DB: s.DB,
	}

	webController := &web.WebController{
		DB: s.DB,
		AC: authController,
	}

	// The order here must be absolutely respected.
	s.Router = pat.New()
	s.Router.Get("/", s.handleRequest(webController.HomeView))
	s.Router.Get("/robots.txt", http.HandlerFunc(webController.CrawlerView))
	s.Router.Get("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("./web/static"))))

	s.Router.Get("/signup/", http.HandlerFunc(authController.SignUpView))
	s.Router.Post("/signup/", s.handleRequest(authController.SignUp))

	s.Router.Get("/auth/confirm/:confirmation", s.handleRequest(authController.PasswordView))
	s.Router.Get("/auth/password/reset/", http.HandlerFunc(authController.ResetPasswordView))
	s.Router.Post("/auth/password/reset/", s.handleRequest(authController.ResetPassword))
	s.Router.Post("/auth/password/", s.handleRequest(authController.SetNewPassword))

	s.Router.Get("/auth/signout/", http.HandlerFunc(authController.SignOut))
	s.Router.Get("/auth/signin/", http.HandlerFunc(authController.SignInView))
	s.Router.Post("/auth/signin/", s.handleRequest(authController.SignIn))

	s.Router.Get("/to/:username/settings", s.handleRequest(webController.SettingsView))
	s.Router.Post("/to/:username/profile", s.handleRequest(webController.SaveProfile))
	s.Router.Get("/to/:username/email", s.handleRequest(webController.EmailView))
	s.Router.Put("/to/:username/email", s.handleRequest(authController.SaveEmailSettings))
	s.Router.Post("/to/:username/email", s.handleRequest(webController.ConfirmEmailChange))
	s.Router.Get("/to/:username/email/:confirmation", s.handleRequest(webController.ChangeEmail))
	s.Router.Get("/to/:username/account", s.handleRequest(webController.AccountView))
	s.Router.Post("/to/:username/account", s.handleRequest(webController.ContinueSignOff))
	s.Router.Post("/to/:username/account/signoff", s.handleRequest(webController.SignOff))
	s.Router.Get("/to/:username/", s.handleRequest(webController.ProfileView))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Router.ServeHTTP(w, r)
}
