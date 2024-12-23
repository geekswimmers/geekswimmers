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
	"strconv"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserController struct {
	DB               storage.Database
	BaseTemplateData *utils.BaseTemplateData
}

func (uc *UserController) SignUpView(res http.ResponseWriter, req *http.Request) {
	reCaptchaSiteKey := config.GetConfiguration().GetString(config.RecaptchaSiteKey)
	sessionData := storage.NewSessionData(req)

	html := utils.GetTemplate("base", "signup")
	err := html.Execute(res, &signUpData{
		SessionData:      sessionData,
		BaseTemplateData: uc.BaseTemplateData,
		ReCaptchaSiteKey: reCaptchaSiteKey,
	})
	if err != nil {
		log.Print(err)
	}
}

func (uc *UserController) SignUp(res http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		log.Print(err)
	}

	sessionData := storage.NewSessionData(req)
	var html *template.Template
	context := &signUpData{
		SessionData:      sessionData,
		BaseTemplateData: uc.BaseTemplateData,
		Email:            strings.ToLower(strings.TrimSpace(req.PostForm.Get("email"))),
		FirstName:        strings.TrimSpace(req.PostForm.Get("firstName")),
		LastName:         strings.TrimSpace(req.PostForm.Get("lastName")),
	}

	// Validates firstName
	if context.FirstName == "" {
		log.Printf("Invalid first name: %v", context.FirstName)
		context.ErrorFirstName = "First Name is empty."
	}

	// Validates lastName
	if context.LastName == "" {
		log.Printf("Invalid last name: %v", context.LastName)
		context.ErrorLastName = "Last Name is empty."
	}

	// Validates email
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
	if context.errorHappened() {
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
	var reCaptchaScore float32
	if reCaptcha != "" {
		reCaptchaScore = getReCaptchaScore(reCaptcha)
	} else {
		reCaptchaScore = -1 // ReCaptcha not used
	}
	confirmation := uuid.New().String()

	userAccount := &UserAccount{
		Email:        context.Email,
		FirstName:    context.FirstName,
		LastName:     context.LastName,
		HumanScore:   reCaptchaScore,
		Confirmation: &confirmation,
	}

	// Creates a new user even before checking if the reCaptchaScore is high.
	// It helps to prevent new registrations with the same email address.
	_, err = InsertUserAccount(userAccount, uc.DB)
	if err != nil {
		log.Printf("Error saving the user: %v", err)
		html = utils.GetTemplate("base", "signup")
		context.Error = `Due to an internal error, it was not possible to create
			your account at this moment. Please, trying again later. 
			Thank you for your undestanding.`
		context.ReCaptchaSiteKey = config.GetConfiguration().GetString(config.RecaptchaSiteKey)
		err = html.Execute(res, context)
		if err != nil {
			log.Print(err)
		}
		return
	}

	if config.GetConfiguration().GetString(config.EmailServer) != "" {
		// Do not send email in case the interaction is more likely done by a bot. The record remains to avoid
		// reattempts and it will be purged by a job after a while.
		if userAccount.HumanScore > 0.5 {
			body := messaging.GetEmailTemplate("signup", &messaging.EmailContext{
				CurrentEmail: userAccount.Email,
				ServerUrl:    config.GetConfiguration().GetString(config.ServerURL),
				Confirmation: *userAccount.Confirmation,
			})

			go messaging.SendMessage(userAccount.Email, "Welcome to Geek Swimmers!", body, uc.DB)
		}

		html = utils.GetTemplate("base", "signup-ok")
		err = html.Execute(res, context)
		if err != nil {
			log.Print(err)
		}
	} else {
		http.Redirect(res, req, "/auth/confirm/"+confirmation, http.StatusSeeOther)
	}
}

func (uc *UserController) PasswordView(res http.ResponseWriter, req *http.Request) {
	confirmation := req.URL.Query().Get(":confirmation")
	userAccount := FindUserAccountByConfirmation(confirmation, "", uc.DB)

	if userAccount != nil {
		sessionData := storage.NewSessionData(req)
		html := utils.GetTemplate("base", "password")
		err := html.Execute(res, &passwordViewData{
			SessionData:      sessionData,
			BaseTemplateData: uc.BaseTemplateData,
			Email:            userAccount.Email,
			Confirmation:     confirmation,
		})
		if err != nil {
			log.Print(err)
		}
	} else {
		http.Redirect(res, req, "/", http.StatusSeeOther)
	}
}

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

		sessionData := storage.NewSessionData(req)
		err = html.Execute(res, &setNewPasswordData{
			SessionData:      sessionData,
			BaseTemplateData: uc.BaseTemplateData,
			Email:            email,
			Confirmation:     confirmation,
			Error:            "Error setting a new password. User confirmation not found.",
		})
		if err != nil {
			log.Print(err)
		}
		return
	}

	password := req.PostForm.Get("password")
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		log.Print(err)
	}
	userAccount.Password = hashedPassword
	userAccount.Confirmation = &confirmation

	err = setUserAccountNewPassword(userAccount, uc.DB)
	if err != nil {
		log.Print(err)
	}

	if config.GetConfiguration().GetString(config.EmailServer) != "" {
		body := messaging.GetEmailTemplate("reset-password-ok", nil)
		go messaging.SendMessage(userAccount.Email, "Your new password on Geek Swimmers has been set!", body, uc.DB)
	}

	http.Redirect(res, req, "/auth/signin/", http.StatusSeeOther)
}

func (uc *UserController) ResetPasswordView(res http.ResponseWriter, req *http.Request) {
	html := utils.GetTemplate("base", "password-reset")

	err := html.Execute(res, nil)
	if err != nil {
		log.Print(err)
	}
}

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

			go messaging.SendMessage(userAccount.Email, "Reset your password on geekswimmers.com", body, uc.DB)
		}
	}

	html := utils.GetTemplate("base", "password-reset-ok")

	err = html.Execute(res, &resetPasswordData{
		Email:            email,
		BaseTemplateData: uc.BaseTemplateData,
	})
	if err != nil {
		log.Print(err)
	}
}

func (uc *UserController) SignInView(res http.ResponseWriter, req *http.Request) {
	reCaptchaSiteKey := config.GetConfiguration().GetString(config.RecaptchaSiteKey)

	html := utils.GetTemplate("base", "signin")

	err := html.Execute(res, &signInData{
		BaseTemplateData: uc.BaseTemplateData,
		ReCaptchaSiteKey: reCaptchaSiteKey,
		SessionData:      storage.NewSessionData(req),
	})
	if err != nil {
		log.Print(err)
	}
}

func (uc *UserController) SignIn(res http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		log.Print(err)
	}

	email := strings.TrimSpace(req.PostForm.Get("identifier"))
	password := req.PostForm.Get("password")

	reCaptcha := req.PostForm.Get("g-recaptcha-response")
	var reCaptchaScore float32
	if reCaptcha != "" {
		reCaptchaScore = getReCaptchaScore(reCaptcha)
	} else {
		reCaptchaScore = -1 // ReCaptcha not used
	}

	userAccount, signInAttempt := uc.authenticate(email, password, utils.GetIP(req), reCaptchaScore)
	if err := InsertSignInAttempt(signInAttempt, uc.DB); err != nil {
		log.Printf("auth.Authenticate: %v", err)
	}
	sessionData := storage.NewSessionData(req)

	if signInAttempt.FailedMatch == FailedMatchAttemptsExceeded {
		html := utils.GetTemplateWithFunctions("base", "signin", template.FuncMap{"html": utils.ToHTML})
		err = html.Execute(res, &signInData{
			BaseTemplateData: uc.BaseTemplateData,
			Error:            "Too many attempts to sign in. Please, try again after an hour from now.",
			Identifier:       email,
			Lock:             true,
			ReCaptchaSiteKey: config.GetConfiguration().GetString(config.RecaptchaSiteKey),
			SessionData:      sessionData,
		})
		if err != nil {
			log.Print(err)
		}
		return
	}

	if userAccount == nil {
		html := utils.GetTemplate("base", "signin")

		log.Printf("Fail to login: %v", email)
		err = html.Execute(res, &signInData{
			BaseTemplateData: uc.BaseTemplateData,
			Error:            "Credentials don't match.",
			Identifier:       email,
			Lock:             false,
			ReCaptchaSiteKey: config.GetConfiguration().GetString(config.RecaptchaSiteKey),
			SessionData:      sessionData,
		})
		if err != nil {
			log.Print(err)
		}
		return
	}

	if signInAttempt.Status != StatusSucceed {
		html := utils.GetTemplateWithFunctions("base", "signin", template.FuncMap{"html": utils.ToHTML})
		err = html.Execute(res, &signInData{
			BaseTemplateData: uc.BaseTemplateData,
			Error:            "Your credentials don't match.", //Did you <a href='/auth/password/reset/'>forget your password</a>?",
			Identifier:       email,
			Lock:             false,
			ReCaptchaSiteKey: config.GetConfiguration().GetString(config.RecaptchaSiteKey),
			SessionData:      sessionData,
		})
		if err != nil {
			log.Print(err)
		}
		return
	}

	if err = uc.addUserToSession(userAccount, res, req); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	// Looks for pending invitation and redirect if any
	confirmationLink := storage.GetSessionEntryValue(req, "profile", "confirmation")
	if len(confirmationLink) > 0 {
		http.Redirect(res, req, confirmationLink, http.StatusSeeOther)
		if err := storage.RemoveSessionEntry(res, req, "profile", "confirmation"); err != nil {
			log.Printf("Error removing session entry: %v", err)
		}
		return
	}

	redirect := storage.GetSessionEntryValue(req, "profile", "redirect")
	if len(redirect) > 0 {
		http.Redirect(res, req, redirect, http.StatusSeeOther)
		if err := storage.RemoveSessionEntry(res, req, "profile", "redirect"); err != nil {
			log.Printf("Error removing session entry: %v", err)
		}
		return
	}

	http.Redirect(res, req, "/", http.StatusSeeOther)
}

func (uc *UserController) authenticate(email, password, ipAddress string, humanScore float32) (*UserAccount, SignInAttempt) {
	signInAttempt := SignInAttempt{
		Identifier: email,
		HumanScore: humanScore,
		IPAddress:  ipAddress,
	}

	if TooManySignInAttempts(ipAddress, uc.DB) {
		signInAttempt.Status = StatusFailed
		signInAttempt.FailedMatch = FailedMatchAttemptsExceeded
		log.Printf("Too many sign in attempts made by: %v", email)
		return nil, signInAttempt
	}

	userAccount := FindUserAccountByEmail(strings.ToLower(email), uc.DB)

	if userAccount == nil {
		signInAttempt.Status = StatusFailed
		signInAttempt.FailedMatch = FailedMatchIdentifier
		log.Printf("User with identifier %v not found", email)
		return userAccount, signInAttempt
	}

	if err := bcrypt.CompareHashAndPassword(userAccount.Password, []byte(password)); err != nil {
		log.Printf("Fail to login: %v", email)
		signInAttempt.Status = StatusFailed
		signInAttempt.FailedMatch = FailedMatchPassword
		return userAccount, signInAttempt
	}

	if humanScore >= 0 && humanScore < 0.5 {
		signInAttempt.Status = StatusFailed
		signInAttempt.FailedMatch = FailedMatchHumanScore
		log.Printf("Human score %v is too low to authenticate", humanScore)
		return userAccount, signInAttempt
	} else if humanScore < 0 {
		log.Printf("ReCaptcha not used")
	}

	if userAccount.SignOff != nil {
		if err := ResetUserAccountSignOffPeriod(userAccount, uc.DB); err != nil {
			log.Printf("Error reseting sign-off period: %v", err)
		} else {
			log.Printf("User %v reset sign off", email)
		}
	}

	signInAttempt.Status = StatusSucceed
	log.Printf("User %v authenticated", email)

	return userAccount, signInAttempt
}

func (uc *UserController) SignOut(res http.ResponseWriter, req *http.Request) {
	email := storage.GetSessionEntryValue(req, "profile", "email")
	role := storage.GetSessionEntryValue(req, "profile", "role")
	firstName := storage.GetSessionEntryValue(req, "profile", "firstName")
	lastName := storage.GetSessionEntryValue(req, "profile", "lastName")

	if err := storage.RemoveSessionEntry(res, req, "profile", "email"); err != nil {
		log.Printf("Error signing out the user %v: %v", email, err)
		return
	}

	if err := storage.RemoveSessionEntry(res, req, "profile", "role"); err != nil {
		log.Printf("Error signing out the user %v with role %v: %v", email, role, err)
		return
	}

	if err := storage.RemoveSessionEntry(res, req, "profile", "firstName"); err != nil {
		log.Printf("Error signing out the user %v %v: %v", firstName, lastName, err)
		return
	}

	if err := storage.RemoveSessionEntry(res, req, "profile", "lastName"); err != nil {
		log.Printf("Error signing out the user %v %v: %v", firstName, lastName, err)
		return
	}

	log.Printf("User %v signed out.", email)

	http.Redirect(res, req, "/", http.StatusSeeOther)
}

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
	if reCaptchaRes != nil {
		defer reCaptchaRes.Body.Close()

		var reCaptchaResBody map[string]interface{}
		decoder := json.NewDecoder(reCaptchaRes.Body)
		if err := decoder.Decode(&reCaptchaResBody); err != nil {
			log.Printf("Error decoding reCaptcha response: %v", err)
		}
		reCaptchaScore, _ := strconv.ParseFloat(fmt.Sprintf("%v", reCaptchaResBody["score"]), 32)

		log.Printf("Success: %v, Score: %v", reCaptchaResBody["success"], reCaptchaResBody["score"])
		return float32(reCaptchaScore)
	}
	return 0
}

func (uc *UserController) addUserToSession(userAccount *UserAccount, res http.ResponseWriter, req *http.Request) error {
	if err := storage.AddSessionEntry(res, req, "profile", "email", userAccount.Email); err != nil {
		return err
	}

	if err := storage.AddSessionEntry(res, req, "profile", "firstName", userAccount.FirstName); err != nil {
		return err
	}

	if err := storage.AddSessionEntry(res, req, "profile", "lastName", userAccount.LastName); err != nil {
		return err
	}

	if err := storage.AddSessionEntry(res, req, "profile", "role", userAccount.Role); err != nil {
		return err
	}

	return nil
}

func (uc *UserController) SaveEmailSettings(res http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		log.Print(err)
	}

	email := storage.GetSessionEntryValue(req, "profile", "email")
	user := FindUserAccountByEmail(email, uc.DB)

	user.PromotionalMsg, err = strconv.ParseBool(req.PostForm.Get("notification_promo"))
	if err != nil {
		log.Printf("Error reading form value: %v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	UpdateUserAccount(user, uc.DB)
}
