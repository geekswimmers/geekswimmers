package user

import (
	"encoding/json"
	"fmt"
	"geekswimmers/config"
	"geekswimmers/storage"
	"geekswimmers/utils"
	"geekswimmers/utils/messaging"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var usernameRegex = regexp.MustCompile("^[a-zA-Z0-9]{2,30}$")

// Context used by the handler to send data to templates.
type webContext struct {
	ReCaptchaSiteKey string
	Confirmation     string
	Email            string
	Identifier       string
	Username         string
	UsernameSignUp   string
	Lock             bool

	Error         string
	ErrorUsername string
	ErrorEmail    string
	ErrorAgreed   string
}

type UserController struct {
	DB                  storage.Database
	BaseTemplateContext *utils.BaseTemplateContext
}

// SignUpView is the http handler for '/signup/'. It populates the sign up form
// with the reCaptcha site key.
// get: /signup/
func (uc *UserController) SignUpView(res http.ResponseWriter, req *http.Request) {
	reCaptchaSiteKey := config.GetConfiguration().GetString(config.RecaptchaSiteKey)

	html := utils.GetTemplate("base", "signup")
	err := html.Execute(res, &webContext{ReCaptchaSiteKey: reCaptchaSiteKey})
	if err != nil {
		log.Print(err)
	}
}

// SignUp is the http handler for '/signup/new/'. It gets data from the signup
// form, creates a new user in the database, and sends an email to the user
// with a link to confirm the email address.
// post: /signup/
func (uc *UserController) SignUp(res http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		log.Print(err)
	}

	var html *template.Template
	context := &webContext{}

	// Validates username
	context.UsernameSignUp = strings.TrimSpace(req.PostForm.Get("username"))
	if !isUsernameValid(context.UsernameSignUp) {
		log.Printf("Invalid username: %v", context.UsernameSignUp)
		context.ErrorUsername = "Invalid username. Avoid characters such as: . , / ? ! @ # $ % ^ & * ( ) + - : ; etc."
	} else {
		userAccount := FindUserAccountByUsername(context.UsernameSignUp, uc.DB)
		if userAccount != nil {
			context.ErrorUsername = "This username is already in use. Please, try another one."
		}
	}

	// Validates email
	context.Email = strings.ToLower(strings.TrimSpace(req.PostForm.Get("email")))
	if !messaging.IsEmailAddressValid(context.Email) {
		log.Printf("Invalid email address: %v", context.Email)
		context.ErrorEmail = "Invalid email address."
	} else {
		userAccount := FindUserAccountByEmail(context.Email, uc.DB)
		if userAccount != nil {
			context.ErrorEmail = "This email is already in use. Do you want to <a href='/auth/signin/'>sign in</a> instead?"
		}
	}

	// Validates terms agreement
	agreed := req.PostForm.Get("agreed")
	if agreed != "on" {
		context.ErrorAgreed = "You have to agree with our terms before creating an account."
	}

	// Back to the signup page in case of error.
	if len(context.ErrorEmail) > 0 || len(context.ErrorUsername) > 0 || len(context.ErrorAgreed) > 0 {
		html = utils.GetTemplateWithFunctions("base", "signup", template.FuncMap{"html": utils.ToHTML})
		log.Printf("Back to signup page with errors.")
		context.ReCaptchaSiteKey = config.GetConfiguration().GetString(config.RecaptchaSiteKey)
		err = html.Execute(res, context)
		if err != nil {
			log.Print(err)
		}
		return
	}

	reCaptcha := req.PostForm.Get("g-recaptcha-response")
	reCaptchaScore := getReCaptchaScore(reCaptcha)
	confirmation := uuid.New().String()

	userAccount := &UserAccount{
		Email:        context.Email,
		Username:     context.UsernameSignUp,
		HumanScore:   reCaptchaScore,
		Confirmation: &confirmation,
	}

	// Creates a new user even before checking if the reCaptchaScore is high. It helps to prevent new registrations with
	// the same email address.
	_, err = InsertUserAccount(userAccount, uc.DB)
	if err != nil {
		log.Printf("Error saving the user: %v", err)
		html = utils.GetTemplate("base", "signup")
		context.ReCaptchaSiteKey = config.GetConfiguration().GetString(config.RecaptchaSiteKey)
		err = html.Execute(res, context)
		if err != nil {
			log.Print(err)
		}
		return
	}

	// Cleans the username so it doesn't reach the template and behaves like the user is authenticated.
	context.UsernameSignUp = ""

	// Do not send email in case the interaction is more likely done by a bot. The record remains to avoid
	// reattempts and it will be purged by a job after a while.
	if userAccount.HumanScore > 0.5 {
		body := messaging.GetEmailTemplate("signup", &messaging.EmailContext{
			Username:     userAccount.Username,
			CurrentEmail: userAccount.Email,
			ServerUrl:    config.GetConfiguration().GetString(config.ServerURL),
			Confirmation: *userAccount.Confirmation,
		})

		go messaging.SendMessage(userAccount.Email, userAccount.Username, "Welcome to followinsteps.com!", body, uc.DB)
	}

	html = utils.GetTemplate("base", "signup-ok")
	err = html.Execute(res, context)
	if err != nil {
		log.Print(err)
	}
}

// PasswordView
// get: /auth/confirm/:confirmation
func (uc *UserController) PasswordView(res http.ResponseWriter, req *http.Request) {
	confirmation := req.URL.Query().Get(":confirmation")
	userAccount := FindUserAccountByConfirmation(confirmation, "", uc.DB)

	if userAccount != nil {
		html := utils.GetTemplate("base", "password")

		err := html.Execute(res, &webContext{Email: userAccount.Email, Confirmation: confirmation})
		if err != nil {
			log.Print(err)
		}
	} else {
		http.Redirect(res, req, "/", http.StatusSeeOther)
	}
}

// SetNewPassword
// post: /auth/password/
func (uc *UserController) SetNewPassword(res http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		log.Print(err)
	}

	email := strings.ToLower(req.PostForm.Get("email"))
	confirmation := req.PostForm.Get("confirmation")

	userAccount := FindUserAccountByConfirmation(confirmation, email, uc.DB)

	if userAccount == nil {
		html := utils.GetTemplate("base", "password")

		err = html.Execute(res, &webContext{Email: email, Confirmation: confirmation, Error: "Error setting a new password. User doesn't match."})
		if err != nil {
			log.Print(err)
		}
		return
	}

	password := req.PostForm.Get("password")
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	userAccount.Password = hashedPassword
	userAccount.Confirmation = &confirmation

	err = setUserAccountNewPassword(userAccount, uc.DB)
	if err != nil {
		log.Print(err)
	}

	body := messaging.GetEmailTemplate("reset-password-ok", nil)
	go messaging.SendMessage(userAccount.Email, userAccount.Username, "Your new password on followinsteps.com has been set!", body, uc.DB)

	http.Redirect(res, req, "/auth/signin/", http.StatusSeeOther)
}

// ResetPasswordView
// get: /auth/password/reset/
func (uc *UserController) ResetPasswordView(res http.ResponseWriter, req *http.Request) {
	html := utils.GetTemplate("base", "password-reset")

	err := html.Execute(res, nil)
	if err != nil {
		log.Print(err)
	}
}

// ResetPassword
// post: /auth/password/reset/
func (uc *UserController) ResetPassword(res http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		log.Print(err)
	}

	email := strings.ToLower(req.PostForm.Get("email"))
	userAccount := FindUserAccountByEmail(email, uc.DB)

	if userAccount != nil {
		confirmation := uuid.New().String()
		userAccount.Confirmation = &confirmation
		err = UpdateUserAccount(userAccount, uc.DB)
		if err == nil {
			body := messaging.GetEmailTemplate("reset-password", &messaging.EmailContext{
				ServerUrl:    config.GetConfiguration().GetString(config.ServerURL),
				Confirmation: *userAccount.Confirmation,
			})

			go messaging.SendMessage(userAccount.Email, userAccount.Username, "Reset your password on followinsteps.com", body, uc.DB)
		}
	}

	html := utils.GetTemplate("base", "password-reset-ok")

	err = html.Execute(res, &webContext{Email: email})
	if err != nil {
		log.Print(err)
	}
}

// SignInView
// get: /auth/signin/
func (uc *UserController) SignInView(res http.ResponseWriter, req *http.Request) {
	reCaptchaSiteKey := config.GetConfiguration().GetString(config.RecaptchaSiteKey)

	html := utils.GetTemplate("base", "signin")

	err := html.Execute(res, &webContext{ReCaptchaSiteKey: reCaptchaSiteKey})
	if err != nil {
		log.Print(err)
	}
}

// SignIn
// post: /auth/signin/
func (uc *UserController) SignIn(res http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		log.Print(err)
	}

	identifier := strings.TrimSpace(req.PostForm.Get("identifier"))
	password := req.PostForm.Get("password")

	reCaptcha := req.PostForm.Get("g-recaptcha-response")
	humanScore := getReCaptchaScore(reCaptcha)

	userAccount, signInAttempt := uc.Authenticate(identifier, password, utils.GetIP(req), humanScore)

	if signInAttempt == nil {
		html := utils.GetTemplateWithFunctions("base", "signin", template.FuncMap{"html": utils.ToHTML})
		err = html.Execute(res, &webContext{
			Identifier:       identifier,
			Error:            "Too many attempts to sign in. Please, try again after an hour from now.",
			ReCaptchaSiteKey: config.GetConfiguration().GetString(config.RecaptchaSiteKey),
			Lock:             true,
		})
		if err != nil {
			log.Print(err)
		}
		return
	}

	if userAccount != nil {
		if signInAttempt.Status == StatusSucceed {
			userAccount.Roles = FindUserRoles(userAccount, uc.DB)

			if err = uc.addUserToSession(userAccount.Username, userAccount.Roles, res, req); err != nil {
				http.Error(res, err.Error(), http.StatusInternalServerError)
				return
			}

			// Looks for pending invitation and redirect if any
			confirmationLink := storage.GetSessionEntryValue(req, "profile", "confirmation")
			if len(confirmationLink) > 0 {
				http.Redirect(res, req, confirmationLink, http.StatusSeeOther)
				storage.RemoveSessionEntry(res, req, "profile", "confirmation")
				return
			}

			redirect := storage.GetSessionEntryValue(req, "profile", "redirect")
			if len(redirect) > 0 {
				http.Redirect(res, req, redirect, http.StatusSeeOther)
				storage.RemoveSessionEntry(res, req, "profile", "redirect")
				return
			}

			http.Redirect(res, req, "/to/"+userAccount.Username+"/", http.StatusSeeOther)
		} else {
			html := utils.GetTemplateWithFunctions("base", "signin", template.FuncMap{"html": utils.ToHTML})
			err = html.Execute(res, &webContext{
				Identifier:       identifier,
				Error:            "Your credentials don't match. Did you <a href='/auth/password/reset/'>forget your password</a>?",
				ReCaptchaSiteKey: config.GetConfiguration().GetString(config.RecaptchaSiteKey),
				Lock:             false,
			})
			if err != nil {
				log.Print(err)
			}
		}
	} else {
		html := utils.GetTemplate("base", "signin")

		log.Printf("Fail to login: %v", identifier)
		err = html.Execute(res, &webContext{
			Identifier:       identifier,
			Error:            "Credentials don't match.",
			ReCaptchaSiteKey: config.GetConfiguration().GetString(config.RecaptchaSiteKey),
			Lock:             false,
		})
		if err != nil {
			log.Print(err)
		}
	}
}

func (uc *UserController) Authenticate(identifier, password, ipAddress string, humanScore float32) (*UserAccount, *SignInAttempt) {
	if TooManySignInAttempts(ipAddress, uc.DB) {
		log.Printf("Too many sign in attempts made by: %v", identifier)
		return nil, nil
	}

	var userAccount *UserAccount
	signInAttempt := SignInAttempt{
		Identifier: identifier,
		HumanScore: humanScore,
		IPAddress:  ipAddress,
	}

	if messaging.IsEmailAddress(identifier) {
		userAccount = FindUserAccountByEmail(strings.ToLower(identifier), uc.DB)
	} else if isUsernameValid(identifier) {
		userAccount = FindUserAccountByUsername(identifier, uc.DB)
	}

	if userAccount != nil {
		if err := bcrypt.CompareHashAndPassword(userAccount.Password, []byte(password)); err != nil {
			log.Printf("Fail to login: %v", identifier)
			signInAttempt.Status = StatusFailed
			signInAttempt.FailedMatch = FailedMatchPassword
		} else {
			if userAccount.SignOff != nil {
				if err = ResetUserAccountSignOffPeriod(userAccount, uc.DB); err != nil {
					log.Printf("Error reseting sign-off period: %v", err)
				} else {
					log.Printf("User %v reset sign off", identifier)
				}
			}

			signInAttempt.Status = StatusSucceed
			log.Printf("User %v authenticated", identifier)
		}
	} else {
		signInAttempt.Status = StatusFailed
		signInAttempt.FailedMatch = FailedMatchIdentifier
		log.Printf("User with identifier %v not found", identifier)
	}

	if humanScore < 0.5 {
		signInAttempt.Status = StatusFailed
		signInAttempt.FailedMatch = FailedMatchHumanScore
		log.Printf("Human score %v is too low to authenticate", humanScore)
	}

	if err := InsertSignInAttempt(signInAttempt, uc.DB); err != nil {
		log.Printf("auth.Authenticate: %v", err)
	}

	return userAccount, &signInAttempt
}

// SignOut
// get: /auth/signout/
func (uc *UserController) SignOut(res http.ResponseWriter, req *http.Request) {
	username := storage.GetSessionEntryValue(req, "profile", "username")
	role := storage.GetSessionEntryValue(req, "profile", "role")

	if err := storage.RemoveSessionEntry(res, req, "profile", "username"); err != nil {
		log.Printf("Error signing out the user %v: %v", username, err)
		return
	}

	if err := storage.RemoveSessionEntry(res, req, "profile", "role"); err != nil {
		log.Printf("Error signing out the user %v with role %v: %v", username, role, err)
		return
	}

	log.Printf("User %v signed out.", username)

	http.Redirect(res, req, "/", http.StatusSeeOther)
}

/* Implements the reCaptcha verification service. */
func getReCaptchaScore(reCaptchaResponse string) float32 {
	reCaptchaSecretKey := config.GetConfiguration().GetString(config.RecaptchaSecretKey)
	reqBody := url.Values{
		"secret":   {reCaptchaSecretKey},
		"response": {reCaptchaResponse},
	}
	reCaptchaRes, err := http.PostForm("https://www.google.com/recaptcha/api/siteverify", reqBody)
	if err != nil {
		log.Printf("Error calling reCaptcha API: %v", err)
	}
	defer reCaptchaRes.Body.Close()
	var reCaptchaResBody map[string]interface{}
	json.NewDecoder(reCaptchaRes.Body).Decode(&reCaptchaResBody)
	reCaptchaScore, _ := strconv.ParseFloat(fmt.Sprintf("%v", reCaptchaResBody["score"]), 32)

	log.Printf("Success: %v, Score: %v", reCaptchaResBody["success"], reCaptchaResBody["score"])
	return float32(reCaptchaScore)
}

func (uc *UserController) addUserToSession(username string, roles []*UserRole, res http.ResponseWriter, req *http.Request) error {
	if err := storage.AddSessionEntry(res, req, "profile", "username", username); err != nil {
		return err
	}

	for _, role := range roles {
		if err := storage.AddSessionEntry(res, req, "profile", "role"+role.Role, role.Role); err != nil {
			return err
		}
	}

	return nil
}

func isUsernameValid(username string) bool {
	return len(username) >= 2 && usernameRegex.MatchString(username)
}

// SaveEmailSettings
// put: /to/:username/email
func (uc *UserController) SaveEmailSettings(res http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		log.Print(err)
	}

	username := storage.GetSessionEntryValue(req, "profile", "username")
	user := FindUserAccountByUsername(username, uc.DB)

	user.PromotionalMsg, err = strconv.ParseBool(req.PostForm.Get("notification_promo"))
	if err != nil {
		log.Printf("Error reading form value: %v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	UpdateUserAccount(user, uc.DB)
}
