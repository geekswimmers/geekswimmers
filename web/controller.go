package web

import (
	"fmt"
	"geekswimmers/auth"
	"geekswimmers/config"
	"geekswimmers/messaging"
	"geekswimmers/storage"
	"geekswimmers/utils"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type context struct {
	User                *auth.UserAccount
	Username            string
	CurrentEmail        string
	NewEmail            string
	Error               string
	ErrorEmail          string
	ErrorAuthentication string
}

type WebController struct {
	DB storage.Database
	AC *auth.AuthController
}

// HomeView
// get: /
func (wc *WebController) HomeView(res http.ResponseWriter, req *http.Request) {
	ctx := &context{}
	ctx.Username = storage.GetSessionValue(req, "profile", "username")

	html := utils.GetTemplateWithFunctions("base", "home", template.FuncMap{"markdown": utils.ToMarkdown})
	err := html.Execute(res, ctx)
	utils.Check(err)
}

// ProfileView
// get: /to/:username/
func (wc *WebController) ProfileView(res http.ResponseWriter, req *http.Request) {
	ctx := &context{}

	user := auth.FindUserAccountByUsername(req.URL.Query().Get(":username"), wc.DB)
	ctx.User = user

	username := storage.GetSessionValue(req, "profile", "username")
	ctx.Username = username

	html := utils.GetTemplateWithFunctions("base", "profile", template.FuncMap{"markdown": utils.ToMarkdown})
	err := html.Execute(res, ctx)
	utils.Check(err)
}

// SettingsView
// get: /to/:username/settings
func (wc *WebController) SettingsView(res http.ResponseWriter, req *http.Request) {
	user := auth.FindUserAccountByUsername(req.URL.Query().Get(":username"), wc.DB)
	username := storage.GetSessionValue(req, "profile", "username")

	if user.Username != username {
		http.Redirect(res, req, fmt.Sprintf("/to/%s", user.Username), http.StatusSeeOther)
	}

	html := utils.GetTemplate("base", "settings-profile")
	err := html.Execute(res, &context{User: user, Username: username})
	utils.Check(err)
}

// SaveProfile
// post: /to/:username/profile
func (wc *WebController) SaveProfile(res http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	utils.Check(err)

	firstName := req.PostForm.Get("firstName")
	lastName := req.PostForm.Get("lastName")

	username := storage.GetSessionValue(req, "profile", "username")

	userAccount := auth.FindUserAccountByUsername(username, wc.DB)
	userAccount.FirstName = &firstName
	userAccount.LastName = &lastName
	err = auth.UpdateUserAccount(userAccount, wc.DB)
	utils.Check(err)

	html := utils.GetTemplate("base", "settings-profile")
	err = html.Execute(res, &context{User: userAccount, Username: username})
	utils.Check(err)
}

// EmailView
// get: /to/:username/email
func (wc *WebController) EmailView(res http.ResponseWriter, req *http.Request) {
	user := auth.FindUserAccountByUsername(req.URL.Query().Get(":username"), wc.DB)

	username := storage.GetSessionValue(req, "profile", "username")

	if user.Username != username {
		http.Redirect(res, req, fmt.Sprintf("/to/%s", user.Username), http.StatusSeeOther)
	}

	html := utils.GetTemplate("base", "settings-email")
	err := html.Execute(res, &context{User: user, Username: username})
	utils.Check(err)
}

// ConfirmEmailChange
// post: /to/:username/email
func (wc *WebController) ConfirmEmailChange(res http.ResponseWriter, req *http.Request) {
	user := auth.FindUserAccountByUsername(req.URL.Query().Get(":username"), wc.DB)

	err := req.ParseForm()
	utils.Check(err)
	context := &context{}
	context.User = user
	currentEmail := req.PostForm.Get("currentEmail")
	context.CurrentEmail = currentEmail
	password := req.PostForm.Get("password")

	username := storage.GetSessionValue(req, "profile", "username")

	// Check if the user requesting the email change is the signed one.
	if user.Username != username {
		http.Redirect(res, req, fmt.Sprintf("/to/%s", user.Username), http.StatusSeeOther)
	}
	context.Username = username

	// Check if the new email already exists.
	context.NewEmail = strings.ToLower(strings.TrimSpace(req.PostForm.Get("newEmail")))
	if !messaging.IsEmailAddressValid(context.NewEmail) {
		log.Printf("Invalid email address: %v", context.NewEmail)
		context.ErrorEmail = "Invalid email address."
	} else {
		userAccount := auth.FindUserAccountByEmail(context.NewEmail, wc.DB)
		if userAccount != nil {
			context.ErrorEmail = "This email is already in use. Do you have another account?"
		}
	}

	if len(context.ErrorEmail) > 0 {
		html := utils.GetTemplate("base", "settings-email")
		err = html.Execute(res, context)
		utils.Check(err)
		return
	}

	// Authenticate user
	if err = bcrypt.CompareHashAndPassword(user.Password, []byte(password)); err != nil {
		html := utils.GetTemplateWithFunctions("base", "settings-email", template.FuncMap{"html": utils.ToHTML})
		log.Printf("Error failing to login: %v", currentEmail)
		context.ErrorAuthentication = "Your password doesn't match."
		err = html.Execute(res, context)
		utils.Check(err)
		return
	}

	// Send confirmation email to the new email
	confirmation := fmt.Sprintf("%v-%v", context.NewEmail, uuid.New().String())
	user.Confirmation = &confirmation

	if err = auth.UpdateUserAccount(user, wc.DB); err != nil {
		context.Error = "We had a problem in the process of changing your email."
		html := utils.GetTemplateWithFunctions("base", "settings-email", template.FuncMap{"html": utils.ToHTML})
		err = html.Execute(res, context)
		utils.Check(err)
		return
	}

	body := messaging.GetEmailTemplate("change-email", &messaging.EmailContext{
		ServerUrl:    config.GetConfiguration().GetString(config.ServerURL),
		Username:     user.Username,
		NewEmail:     context.NewEmail,
		Confirmation: *user.Confirmation,
	})
	go messaging.SendMessage(context.NewEmail, user.Username, "Changing your email on geekswimmers.com", body, wc.DB)

	html := utils.GetTemplate("base", "settings-email-ok")
	err = html.Execute(res, context)
	utils.Check(err)
}

// ChangeEmail
// get: /to/:username/email/:confirmation
func (wc *WebController) ChangeEmail(res http.ResponseWriter, req *http.Request) {
	user := auth.FindUserAccountByUsername(req.URL.Query().Get(":username"), wc.DB)
	userWithConfirmation := auth.FindUserAccountByConfirmation(req.URL.Query().Get(":confirmation"), user.Email, wc.DB)

	if userWithConfirmation == nil {
		log.Printf("Confirmation doesn't match: %v", req.URL.Query().Get(":confirmation"))
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}

	username := storage.GetSessionValue(req, "profile", "username")

	confirmation := []rune(*userWithConfirmation.Confirmation)
	newEmail := string(confirmation[0 : len(confirmation)-37])

	if messaging.IsEmailAddress(newEmail) {
		newEmail = strings.ToLower(newEmail)
		if err := auth.SetUserAccountNewEmail(userWithConfirmation, newEmail, wc.DB); err != nil {
			log.Printf("Error setting the new email %s: %s", newEmail, err)
			http.Redirect(res, req, "/", http.StatusSeeOther)
			return
		}

		bodyOld := messaging.GetEmailTemplate("change-email-notification-old", &messaging.EmailContext{
			CurrentEmail: user.Email,
			NewEmail:     newEmail,
		})
		go messaging.SendMessage(userWithConfirmation.Email, userWithConfirmation.Username, "Your email on geekswimmers.com has been changed!", bodyOld, wc.DB)

		bodyNew := messaging.GetEmailTemplate("change-email-notification-new", &messaging.EmailContext{
			NewEmail: newEmail,
		})

		go messaging.SendMessage(newEmail, userWithConfirmation.Username, "Your email on geekswimmers.com has been changed!", bodyNew, wc.DB)

		html := utils.GetTemplate("base", "settings-email-changed")
		err := html.Execute(res, &context{User: user, CurrentEmail: userWithConfirmation.Email, NewEmail: newEmail, Username: username})
		utils.Check(err)
	} else {
		log.Printf("New email address is invalid: %s", newEmail)
		http.Redirect(res, req, "/", http.StatusSeeOther)
	}
}

// AccountView
// get: /to/:username/account
func (wc *WebController) AccountView(res http.ResponseWriter, req *http.Request) {
	user := auth.FindUserAccountByUsername(req.URL.Query().Get(":username"), wc.DB)

	username := storage.GetSessionValue(req, "profile", "username")

	if user.Username != username {
		http.Redirect(res, req, fmt.Sprintf("/to/%s", user.Username), http.StatusSeeOther)
	}

	html := utils.GetTemplate("base", "settings-account")
	err := html.Execute(res, &context{User: user, Username: username})
	utils.Check(err)
}

func (wc *WebController) ContinueSignOff(res http.ResponseWriter, req *http.Request) {
	user := auth.FindUserAccountByUsername(req.URL.Query().Get(":username"), wc.DB)

	// Authenticate user
	err := req.ParseForm()
	utils.Check(err)
	identifier := req.PostForm.Get("email")
	password := req.PostForm.Get("password")

	userAccount, signInAttempt := wc.AC.Authenticate(identifier, password, utils.GetIP(req), 1.0)

	if signInAttempt != nil && signInAttempt.Status == auth.StatusSucceed && (user.Username == userAccount.Username) {
		html := utils.GetTemplate("base", "settings-account-signoff")
		err := html.Execute(res, &context{User: userAccount, Username: userAccount.Username})
		utils.Check(err)
	} else {
		html := utils.GetTemplate("base", "settings-account")
		err := html.Execute(res, &context{User: user, Username: user.Username,
			ErrorAuthentication: "Your credentials don't match."})
		utils.Check(err)
	}
}

func (wc *WebController) SignOff(res http.ResponseWriter, req *http.Request) {
	user := auth.FindUserAccountByUsername(req.URL.Query().Get(":username"), wc.DB)
	username := storage.GetSessionValue(req, "profile", "username")

	if user.Username != username {
		html := utils.GetTemplate("base", "settings-profile")
		err := html.Execute(res, &context{User: user, Username: username})
		utils.Check(err)
	}

	err := req.ParseForm()
	utils.Check(err)
	feedback := req.PostForm.Get("feedback")

	if err = auth.StartUserAccountSignOffPeriod(user, feedback, wc.DB); err != nil {
		html := utils.GetTemplate("base", "settings-account-signoff")
		err := html.Execute(res, &context{User: user, Username: username,
			Error: "We couldn't sign you off at the moment due to an internal error. Please give us some time to fix it and try again later."})
		utils.Check(err)
		return
	}

	body := messaging.GetEmailTemplate("signoff", &messaging.EmailContext{})
	go messaging.SendMessage(user.Email, username, "Sign-Off from geekswimmers.com", body, wc.DB)

	log.Printf("User %v signed off", username)
	wc.AC.SignOut(res, req)
}

// CrawlerView
// get: /robots.txt
func (wc *WebController) CrawlerView(res http.ResponseWriter, req *http.Request) {
	txt, err := template.ParseFiles("web/templates/robots.txt")
	utils.Check(err)

	err = txt.Execute(res, nil)
	utils.Check(err)
}
