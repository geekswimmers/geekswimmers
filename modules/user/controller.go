package user

import (
	"crypto/rsa"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"geekswimmers/config"
	"geekswimmers/modules/content"
	"geekswimmers/modules/swimming"
	"geekswimmers/modules/times"
	"geekswimmers/storage"
	"geekswimmers/utils"
	"geekswimmers/utils/messaging"
	"html/template"
	"io"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Controller struct {
	DB               storage.Database
	BaseTemplateData *utils.BaseTemplateData
}

func (uc *Controller) SignUpView(res http.ResponseWriter, req *http.Request) {
	reCaptchaSiteKey := config.GetConfiguration().GetString(config.RecaptchaSiteKey)
	googleClientID := config.GetConfiguration().GetString(config.GoogleClientID)
	sessionData := storage.NewSessionData(req)

	jurisdictions, err := swimming.FindJurisdictionsByLevel(swimming.JurisdictionLevelRegion, uc.DB)
	if err != nil {
		log.Printf("Error loading jurisdictions: %v", err)
	}

	data := &signUpData{
		SessionData:      sessionData,
		BaseTemplateData: uc.BaseTemplateData,
		Jurisdictions:    jurisdictions,
		ReCaptchaSiteKey: reCaptchaSiteKey,
		GoogleClientID:   googleClientID,
	}

	data.PrivacyPolicy, err = content.LoadMarkdownContent(fmt.Sprintf("web/content/%s", "privacy-policy.md"))
	if err != nil {
		log.Printf("Error loading the privacy policy: %v", err)
	}

	data.TermsAndConditions, err = content.LoadMarkdownContent(fmt.Sprintf("web/content/%s", "terms-and-conditions.md"))
	if err != nil {
		log.Printf("Error loading the terms and conditions: %v", err)
	}

	if !userAccountExists(uc.DB) {
		data.Error = "You'll become the first Geek Swimmers' user. You will be automatically assigned to an admin role."
		data.FirstUserAccount = true
	}

	html := utils.GetTemplateWithFunctions("base", "signup", template.FuncMap{"markdown": utils.MarkdownToHTML})
	err = html.Execute(res, data)
	if err != nil {
		log.Print(err)
	}
}

func (uc *Controller) SignUp(res http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		log.Print(err)
	}

	sessionData := storage.NewSessionData(req)
	var html *template.Template

	jurisdiction, err := strconv.ParseInt(req.PostForm.Get("jurisdiction"), 10, 64)
	if err != nil {
		jurisdiction = 0
	}

	team, err := strconv.ParseInt(req.PostForm.Get("team"), 10, 64)
	if err != nil {
		team = 0
	}

	data := &signUpData{
		SessionData:       sessionData,
		BaseTemplateData:  uc.BaseTemplateData,
		Email:             strings.ToLower(strings.TrimSpace(req.PostForm.Get("email"))),
		FirstName:         strings.TrimSpace(req.PostForm.Get("firstName")),
		LastName:          strings.TrimSpace(req.PostForm.Get("lastName")),
		Role:              req.PostForm.Get("role"),
		BirthDate:         req.PostForm.Get("birthDate"),
		Gender:            req.PostForm.Get("gender"),
		Jurisdiction:      jurisdiction,
		Team:              team,
		UserAccountExists: userAccountExists(uc.DB),
	}

	if !data.UserAccountExists {
		data.Role = RoleAdmin
	}

	data.ExistingUserAccount = FindUserAccountByEmail(data.Email, uc.DB)

	// Back to the signup page in case of error.
	if !data.valid() {
		jurisdictions, err := swimming.FindJurisdictionsByLevel(swimming.JurisdictionLevelRegion, uc.DB)
		if err != nil {
			log.Printf("Error loading jurisdictions: %v", err)
		}
		data.Jurisdictions = jurisdictions

		html = utils.GetTemplateWithFunctions("base", "signup", template.FuncMap{"html": utils.ToHTML})
		log.Printf("Back to signup page with errors.")
		data.ReCaptchaSiteKey = config.GetConfiguration().GetString(config.RecaptchaSiteKey)
		err = html.Execute(res, data)
		if err != nil {
			log.Print(err)
		}
		return
	}

	userAccount := data.createUserAccount()

	reCaptcha := req.PostForm.Get("g-recaptcha-response")
	var reCaptchaScore float32
	if reCaptcha != "" {
		reCaptchaScore = getReCaptchaScore(reCaptcha)
	} else {
		reCaptchaScore = -1 // ReCaptcha not used
	}
	confirmation := uuid.New().String()
	userAccount.Confirmation = &confirmation
	userAccount.HumanScore = reCaptchaScore

	// Creates a new user even before checking if the reCaptchaScore is high.
	// It helps to prevent new registrations with the same email address.
	userAccount.ID, err = InsertUserAccount(userAccount, uc.DB)
	if err != nil {
		log.Printf("Error saving the user: %v", err)
		html = utils.GetTemplate("base", "signup")
		data.Error = `Due to an internal error, it was not possible to create
			your account at this moment. Please, trying again later.
			Thank you for your undestanding.`
		data.ReCaptchaSiteKey = config.GetConfiguration().GetString(config.RecaptchaSiteKey)
		err = html.Execute(res, data)
		if err != nil {
			log.Print(err)
		}
		return
	}

	if userAccount.Role == RoleSwimmer {
		swimmer := data.createSwimmer(userAccount)

		_, err = saveSwimmer(swimmer, uc.DB)
		if err != nil {
			log.Printf("Error saving the swimmer during signup: %v", err)
			html = utils.GetTemplate("base", "signup")
			data.Error = `Due to an internal error, it was not possible to create
				your account at this moment. Please, trying again later.
				Thank you for your undestanding.`
			data.ReCaptchaSiteKey = config.GetConfiguration().GetString(config.RecaptchaSiteKey)
			err = html.Execute(res, data)
			if err != nil {
				log.Print(err)
			}
			return
		}
	}

	if config.GetConfiguration().GetString(config.EmailServer) != "" {
		// Do not send email in case the interaction is more likely done by a bot. The record remains to avoid
		// reattempts and it will be purged by a job after a while.
		if userAccount.HumanScore > 0.5 {
			body := messaging.GetEmailTemplate("signup", &messaging.EmailData{
				CurrentEmail: userAccount.Email,
				ServerUrl:    config.GetConfiguration().GetString(config.ServerURL),
				Confirmation: *userAccount.Confirmation,
			})

			go messaging.SendMessage(userAccount.Email, "Welcome to Geek Swimmers!", body, uc.DB)
		}

		html = utils.GetTemplate("base", "signup-ok")
		err = html.Execute(res, data)
		if err != nil {
			log.Print(err)
		}
	} else {
		http.Redirect(res, req, "/auth/confirm/"+confirmation, http.StatusSeeOther)
	}
}

func (uc *Controller) ChangePasswordView(res http.ResponseWriter, req *http.Request) {
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

func (uc *Controller) SetNewPassword(res http.ResponseWriter, req *http.Request) {
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

func (uc *Controller) ResetPasswordView(res http.ResponseWriter, _ *http.Request) {
	html := utils.GetTemplate("base", "password-reset")

	err := html.Execute(res, nil)
	if err != nil {
		log.Print(err)
	}
}

func (uc *Controller) ResetPassword(res http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		log.Print(err)
	}

	email := strings.ToLower(req.PostForm.Get("email"))
	userAccount := FindUserAccountByEmail(email, uc.DB)

	if userAccount != nil {
		confirmation := uuid.New().String()
		userAccount.Confirmation = &confirmation
		err = updateUserAccount(userAccount, uc.DB)
		if err == nil {
			body := messaging.GetEmailTemplate("reset-password", &messaging.EmailData{
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

func (uc *Controller) SignInView(res http.ResponseWriter, req *http.Request) {
	reCaptchaSiteKey := config.GetConfiguration().GetString(config.RecaptchaSiteKey)
	googleClientID := config.GetConfiguration().GetString(config.GoogleClientID)

	if !userAccountExists(uc.DB) {
		http.Redirect(res, req, "/signup/", http.StatusSeeOther)
		return
	}

	html := utils.GetTemplate("base", "signin")

	err := html.Execute(res, &signInData{
		BaseTemplateData: uc.BaseTemplateData,
		ReCaptchaSiteKey: reCaptchaSiteKey,
		GoogleClientID:   googleClientID,
		SessionData:      storage.NewSessionData(req),
	})
	if err != nil {
		log.Print(err)
	}
}

func (uc *Controller) SignIn(res http.ResponseWriter, req *http.Request) {
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

	if userAccount.Role == RoleSwimmer {
		swimmer := FindSwimmerByUserAccount(userAccount, uc.DB)
		if err = uc.addSwimmerToSession(swimmer.Swimmer, res, req); err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
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

	if userAccount.Role == RoleSwimmer {
		http.Redirect(res, req, "/profile/swimmers", http.StatusSeeOther)
	} else {
		http.Redirect(res, req, "/profile/", http.StatusSeeOther)
	}
}

func (uc *Controller) ProfileView(res http.ResponseWriter, req *http.Request, sessionData *storage.SessionData) {
	user := FindUserAccountByEmail(sessionData.Email, uc.DB)
	swimmers, err := findSwimmersParent(user, uc.DB)
	if err != nil {
		log.Printf("Error finding swimmers: %v", err)
	}

	data := &profileData{
		BaseTemplateData: uc.BaseTemplateData,
		SessionData:      sessionData,
		FirstName:        user.FirstName,
		LastName:         user.LastName,
		Email:            user.Email,
		Role:             user.Role,
		Swimmers:         swimmers,
	}

	html := utils.GetTemplateWithFunctions("base", "profile", template.FuncMap{"Title": utils.Title})
	if err := html.Execute(res, data); err != nil {
		log.Printf("Error loading the user's profile: %v", err)
	}
}

func (uc *Controller) ProfileFormView(res http.ResponseWriter, req *http.Request, sessionData *storage.SessionData) {
	user := FindUserAccountByEmail(sessionData.Email, uc.DB)
	swimmer := FindSwimmerByUserAccount(user, uc.DB)
	jurisdictions, err := swimming.FindJurisdictionsByLevel(swimming.JurisdictionLevelRegion, uc.DB)
	if err != nil {
		log.Printf("Error loading jurisdictions: %v", err)
	}

	var teamId int64
	var jurisdictionId int64
	if swimmer != nil {
		teamId = swimmer.Swimmer.Team.ID.Int64
		jurisdictionId = swimmer.Swimmer.Team.Jurisdiction.ID.Int64
	} else {
		teamId = 0
		jurisdictionId = 0
	}

	data := &profileData{
		BaseTemplateData: uc.BaseTemplateData,
		SessionData:      sessionData,
		FirstName:        user.FirstName,
		Jurisdiction:     jurisdictionId,
		Jurisdictions:    jurisdictions,
		LastName:         user.LastName,
		Email:            user.Email,
		Role:             user.Role,
		Team:             teamId,
	}

	if swimmer != nil {
		birthDate := &swimmer.Swimmer.BirthDate.Time
		gender := swimmer.Swimmer.Gender.String
		data.BirthDate = birthDate.Format("2006-01-02")
		data.Gender = gender
	}

	html := utils.GetTemplateWithFunctions("base", "profile-form", template.FuncMap{"Title": utils.Title})
	if err := html.Execute(res, data); err != nil {
		log.Printf("Error loading the user's profile: %v", err)
	}
}

func (uc *Controller) ProfileForm(res http.ResponseWriter, req *http.Request, sessionData *storage.SessionData) {
	err := req.ParseForm()
	if err != nil {
		log.Print(err)
	}

	jurisdiction, err := strconv.ParseInt(req.PostForm.Get("jurisdiction"), 10, 64)
	if err != nil {
		jurisdiction = 0
	}

	team, err := strconv.ParseInt(req.PostForm.Get("team"), 10, 64)
	if err != nil {
		team = 0
	}

	data := &profileData{
		BaseTemplateData: uc.BaseTemplateData,
		SessionData:      sessionData,
		Email:            strings.ToLower(strings.TrimSpace(req.PostForm.Get("email"))),
		FirstName:        strings.TrimSpace(req.PostForm.Get("firstName")),
		LastName:         strings.TrimSpace(req.PostForm.Get("lastName")),
		BirthDate:        req.PostForm.Get("birthDate"),
		Gender:           req.PostForm.Get("gender"),
		Jurisdiction:     jurisdiction,
		Team:             team,
	}

	data.CurrentUser = FindUserAccountByEmail(sessionData.Email, uc.DB)
	data.ExistingUser = FindUserAccountByEmail(data.Email, uc.DB)
	data.CurrentSwimmer = FindSwimmerByUserAccount(data.CurrentUser, uc.DB)
	data.Role = data.CurrentUser.Role

	// Back to the profile form in case of error.
	if !data.valid() {
		jurisdictions, err := swimming.FindJurisdictionsByLevel(swimming.JurisdictionLevelRegion, uc.DB)
		if err != nil {
			log.Printf("Error loading jurisdictions: %v", err)
		}
		data.Jurisdictions = jurisdictions

		html := utils.GetTemplateWithFunctions("base", "profile-form", template.FuncMap{"Title": utils.Title})
		log.Printf("Back to profile page with errors.")
		err = html.Execute(res, data)
		if err != nil {
			log.Print(err)
		}
		return
	}

	data.CurrentUser.FirstName = data.FirstName
	data.CurrentUser.LastName = data.LastName
	data.CurrentUser.Email = data.Email

	if data.CurrentUser.Role == RoleSwimmer && data.CurrentSwimmer != nil {
		data.CurrentSwimmer.Swimmer.Team.ID.Int64 = team
		data.CurrentSwimmer.Swimmer.FirstName = data.CurrentUser.FirstName
		data.CurrentSwimmer.Swimmer.LastName = data.CurrentUser.LastName

		data.CurrentSwimmer.Swimmer.Gender = sql.NullString{
			String: data.Gender,
		}

		birthDate, _ := time.Parse("2006-01-02", data.BirthDate)
		data.CurrentSwimmer.Swimmer.BirthDate = sql.NullTime{
			Time:  birthDate,
			Valid: true,
		}

		if err := updateSwimmerProfile(data.CurrentUser, data.CurrentSwimmer, uc.DB); err != nil {
			log.Printf("Error saving the swimmers' profile: %v", err)
			html := utils.GetTemplateWithFunctions("base", "profile-form", template.FuncMap{"Title": utils.Title})
			data.Error = `Due to an internal error, it was not possible to save
			your account at this moment. Please, trying again later.
			Thank you for your understanding.`
			err = html.Execute(res, data)
			if err != nil {
				log.Print(err)
			}
			return
		}

		if err := uc.addSwimmerToSession(data.CurrentSwimmer.Swimmer, res, req); err != nil {
			log.Printf("Error adding swimmer to session: %v", err)
		}
	} else {
		if err := updateProfile(data.CurrentUser, uc.DB); err != nil {
			log.Printf("Error saving the profile: %v", err)
			html := utils.GetTemplateWithFunctions("base", "profile-form", template.FuncMap{"Title": utils.Title})
			data.Error = `Due to an internal error, it was not possible to save
			your profile at this moment. Please, trying again later.
			Thank you for your understanding.`
			err = html.Execute(res, data)
			if err != nil {
				log.Print(err)
			}
			return
		}
	}

	if err := uc.addUserToSession(data.CurrentUser, res, req); err != nil {
		log.Printf("Error adding user to session: %v", err)
	}

	if data.CurrentUser.Role == RoleSwimmer {
		http.Redirect(res, req, "/profile/swimmers", http.StatusSeeOther)
	} else {
		http.Redirect(res, req, "/profile/", http.StatusSeeOther)
	}
}

func (uc *Controller) ProfileSwimmerView(res http.ResponseWriter, req *http.Request, sessionData *storage.SessionData) {
	var swimmer *UserSwimmer
	parentSwimmer := &ParentSwimmer{
		Responsible: false,
	}

	if sessionData.Role == RoleSwimmer {
		// The swimmer is the user itself
		swimmer = FindSwimmerByEmail(sessionData.Email, uc.DB)
	} else {
		// The swimmer is a child of the user
		id := req.URL.Query().Get(":id")
		swimmerId, _ := strconv.ParseInt(id, 10, 64)
		swimmer = FindSwimmerByID(swimmerId, uc.DB)

		parent := FindUserAccountByEmail(sessionData.Email, uc.DB)
		parentSwimmer = findParentSwimmer(parent, swimmer, uc.DB)

		if parentSwimmer == nil || parentSwimmer.Approval != ParentSwimmerApprovalAccepted {
			http.Error(res, "The swimmer is not related to the parent.", http.StatusForbidden)
			return
		}
	}

	course := req.URL.Query().Get("course")
	if course == "" {
		course = times.DefaultCourse
	}

	linkRequests, err := findLinkRequests(swimmer, uc.DB)
	if err != nil {
		log.Printf("ProfileSwimmerView: %v", err)
	}

	bestTimes, err := findSwimmerBestTimes(swimmer, course, uc.DB)
	if err != nil {
		log.Printf("ProfileSwimmerView: %v", err)
	}
	bestTimes = addSwimmingPointsToRecords(swimmer, course, bestTimes, uc.DB)

	data := &swimmerData{
		BaseTemplateData: uc.BaseTemplateData,
		Course:           course,
		SessionData:      sessionData,
		Swimmer:          swimmer,
		BestTimes:        bestTimes,
		LinkRequests:     linkRequests,
		ParentSwimmer:    parentSwimmer,
	}

	html := utils.GetTemplateWithFunctions("base", "profile-swimmer",
		template.FuncMap{
			"Title":              utils.Title,
			"FormatMilliseconds": utils.FormatMilliseconds,
		})
	if err := html.Execute(res, data); err != nil {
		log.Printf("Error loading the swimmer's profile: %v", err)
	}
}

func (uc *Controller) ProfileSwimmerFormView(res http.ResponseWriter, req *http.Request, sessionData *storage.SessionData) {
	jurisdictions, err := swimming.FindJurisdictionsByLevel(swimming.JurisdictionLevelRegion, uc.DB)
	if err != nil {
		log.Printf("Error loading jurisdictions: %v", err)
	}

	id := req.URL.Query().Get(":id")
	swimmerId, _ := strconv.ParseInt(id, 10, 64)
	swimmer := FindSwimmerByID(swimmerId, uc.DB)

	data := &swimmerData{
		BaseTemplateData: uc.BaseTemplateData,
		SessionData:      sessionData,
		Swimmer:          swimmer,
		FirstName:        swimmer.Swimmer.FirstName,
		LastName:         swimmer.Swimmer.LastName,
		BirthDate:        swimmer.Swimmer.BirthDate.Time.Format("2006-01-02"),
		Gender:           swimmer.Swimmer.Gender.String,
		Team:             swimmer.Swimmer.Team.ID.Int64,
		Jurisdiction:     swimmer.Swimmer.Team.Jurisdiction.ID.Int64,
		Jurisdictions:    jurisdictions,
		SwimRanking:      swimmer.Swimmer.SwimRanking.String,
		SwimCloud:        swimmer.Swimmer.SwimCloud.String,
	}

	html := utils.GetTemplateWithFunctions("base", "profile-swimmer-form",
		template.FuncMap{
			"Title": utils.Title,
		})
	if err := html.Execute(res, data); err != nil {
		log.Printf("Error loading the swimmer's profile form: %v", err)
	}
}

func (uc *Controller) ProfileSwimmerForm(res http.ResponseWriter, req *http.Request, sessionData *storage.SessionData) {
	err := req.ParseForm()
	if err != nil {
		log.Print(err)
	}

	id := req.URL.Query().Get(":id")
	swimmerId, _ := strconv.ParseInt(id, 10, 64)
	swimmer := FindSwimmerByID(swimmerId, uc.DB)

	jurisdiction, err := strconv.ParseInt(req.PostForm.Get("jurisdiction"), 10, 64)
	if err != nil {
		jurisdiction = 0
	}

	team, err := strconv.ParseInt(req.PostForm.Get("team"), 10, 64)
	if err != nil {
		team = 0
	}

	data := &swimmerData{
		BaseTemplateData: uc.BaseTemplateData,
		SessionData:      sessionData,
		Team:             team,
		Jurisdiction:     jurisdiction,
		Swimmer:          swimmer,
		Consent:          "on",
		FirstName:        req.PostForm.Get("firstName"),
		LastName:         req.PostForm.Get("lastName"),
		Gender:           req.PostForm.Get("gender"),
		BirthDate:        req.PostForm.Get("birthDate"),
		SwimRanking:      req.PostForm.Get("swimRanking"),
		SwimCloud:        req.PostForm.Get("swimCloud"),
	}

	if !data.valid() {
		log.Print("Swimmer data invalid")
		html := utils.GetTemplateWithFunctions("base", "profile-swimmer-form", template.FuncMap{
			"Title": utils.Title,
		})
		data.Jurisdictions, err = swimming.FindJurisdictionsByLevel(swimming.JurisdictionLevelRegion, uc.DB)
		if err != nil {
			log.Printf("Error loading jurisdictions: %v", err)
			http.Error(res, err.Error(), http.StatusInternalServerError)
		}
		err = html.Execute(res, data)
		if err != nil {
			log.Print(err)
		}
		return
	}

	swimmer.Swimmer.FirstName = data.FirstName
	swimmer.Swimmer.LastName = data.LastName
	swimmer.Swimmer.BirthDate.Time, _ = time.Parse("2006-01-02", data.BirthDate)
	swimmer.Swimmer.Gender.String = data.Gender
	swimmer.Swimmer.Team.ID.Int64 = data.Team
	swimmer.Swimmer.SwimRanking.String = data.SwimRanking
	swimmer.Swimmer.SwimCloud.String = data.SwimCloud

	_, err = saveSwimmer(swimmer, uc.DB)
	if err != nil {
		log.Printf("Error saving the swimmer profile: %v", err)
		html := utils.GetTemplateWithFunctions("base", "profile-swimmer-form", template.FuncMap{
			"Title": utils.Title,
		})
		data.Error = `Due to an internal error, it was not possible to save
			the swimmer at this moment. Please, trying again later.
			Thank you for your understanding.`
		data.Jurisdictions, err = swimming.FindJurisdictionsByLevel(swimming.JurisdictionLevelRegion, uc.DB)
		if err != nil {
			log.Printf("Error loading jurisdictions: %v", err)
			http.Error(res, err.Error(), http.StatusInternalServerError)
		}
		err = html.Execute(res, data)
		if err != nil {
			log.Print(err)
		}
		return
	}

	http.Redirect(res, req, fmt.Sprintf("/profile/swimmers/%d/", swimmer.ID), http.StatusSeeOther)
}

func (uc *Controller) SwimmerFormView(res http.ResponseWriter, req *http.Request, sessionData *storage.SessionData) {
	jurisdiction, err := strconv.ParseInt(req.PostForm.Get("jurisdiction"), 10, 64)
	if err != nil {
		jurisdiction = 0
	}

	team, err := strconv.ParseInt(req.PostForm.Get("team"), 10, 64)
	if err != nil {
		team = 0
	}

	data := &swimmerData{
		BaseTemplateData: uc.BaseTemplateData,
		SessionData:      sessionData,
		Jurisdiction:     jurisdiction,
		Team:             team,
	}

	data.Jurisdictions, err = swimming.FindJurisdictionsByLevel(swimming.JurisdictionLevelRegion, uc.DB)
	if err != nil {
		log.Printf("Error loading jurisdictions: %v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	html := utils.GetTemplate("base", "swimmer-form")
	if err := html.Execute(res, data); err != nil {
		log.Printf("Error loading the swimmer form: %v", err)
	}
}

func (uc *Controller) SwimmerForm(res http.ResponseWriter, req *http.Request, sessionData *storage.SessionData) {
	err := req.ParseForm()
	if err != nil {
		log.Print(err)
	}

	jurisdiction, err := strconv.ParseInt(req.PostForm.Get("jurisdiction"), 10, 64)
	if err != nil {
		jurisdiction = 0
	}

	team, err := strconv.ParseInt(req.PostForm.Get("team"), 10, 64)
	if err != nil {
		team = 0
	}

	var html *template.Template
	data := &swimmerData{
		SessionData:      sessionData,
		BaseTemplateData: uc.BaseTemplateData,
		Consent:          req.PostForm.Get("consent"),
		FirstName:        strings.TrimSpace(req.PostForm.Get("firstName")),
		LastName:         strings.TrimSpace(req.PostForm.Get("lastName")),
		BirthDate:        req.PostForm.Get("birthDate"),
		Gender:           req.PostForm.Get("gender"),
		Jurisdiction:     jurisdiction,
		Team:             team,
	}

	// Back to the signup page in case of error.
	if !data.valid() {
		data.Jurisdictions, err = swimming.FindJurisdictionsByLevel(swimming.JurisdictionLevelRegion, uc.DB)
		if err != nil {
			log.Printf("Error loading jurisdictions: %v", err)
			http.Error(res, err.Error(), http.StatusInternalServerError)
		}

		html = utils.GetTemplate("base", "swimmer-form")
		log.Printf("Back to swimmer form with errors.")
		err = html.Execute(res, data)
		if err != nil {
			log.Print(err)
		}
		return
	}

	swimmer := data.createSwimmer()

	swimmer.ID, err = saveSwimmer(swimmer, uc.DB)
	if err != nil {
		log.Printf("Error saving the swimmer: %v", err)

		data.Jurisdictions, err = swimming.FindJurisdictionsByLevel(swimming.JurisdictionLevelRegion, uc.DB)
		if err != nil {
			log.Printf("Error loading jurisdictions: %v", err)
			http.Error(res, err.Error(), http.StatusInternalServerError)
		}

		html = utils.GetTemplate("base", "swimmer-form")
		data.Error = `Due to an internal error, it was not possible to create
			your account at this moment. Please, trying again later.
			Thank you for your undestanding.`
		err = html.Execute(res, data)
		if err != nil {
			log.Print(err)
		}
		return
	}

	parent := FindUserAccountByEmail(sessionData.Email, uc.DB)
	swimmers := []*UserSwimmer{swimmer}
	if err := linkSwimmersToParent(parent, swimmers, uc.DB); err != nil {
		log.Printf("Error linking the swimmer to the parent: %v", err)
		data.Jurisdictions, err = swimming.FindJurisdictionsByLevel(swimming.JurisdictionLevelRegion, uc.DB)
		if err != nil {
			log.Printf("Error loading jurisdictions: %v", err)
			http.Error(res, err.Error(), http.StatusInternalServerError)
		}

		html = utils.GetTemplate("base", "swimmer-form")
		data.Error = `Due to an internal error, it was not possible to create
			your account at this moment. Please, trying again later.
			Thank you for your undestanding.`
		err = html.Execute(res, data)
		if err != nil {
			log.Print(err)
		}
		return
	}

	http.Redirect(res, req, "/profile/", http.StatusSeeOther)
}

func (uc *Controller) SwimmerBestTimeView(res http.ResponseWriter, req *http.Request, sessionData *storage.SessionData) {
	id := req.URL.Query().Get(":id")
	swimmerId, _ := strconv.ParseInt(id, 10, 64)
	swimmer := FindSwimmerByID(swimmerId, uc.DB)

	id = req.URL.Query().Get(":bestId")
	bestId, _ := strconv.ParseInt(id, 10, 64)
	bestTime := findSwimmerBestTime(swimmer, bestId, uc.DB)

	meets, err := times.FindMeetsWithTimeStandardByJurisdiction(swimmer.Swimmer.Team.Jurisdiction.ID.Int64, uc.DB)
	if err != nil {
		log.Printf("user.%v", err)
	}

	var foundMeets []*times.Meet
	for _, meet := range meets {
		meet.Age = swimmer.Swimmer.AgeAt(meet.AgeDate)
		searchAge := meet.Age

		if !meet.MinAgeEnforced && meet.TimeStandard.MinAgeTime != nil && meet.Age < *meet.TimeStandard.MinAgeTime {
			searchAge = *meet.TimeStandard.MinAgeTime
		} else if meet.MinAgeEnforced && meet.Age < *meet.TimeStandard.MinAgeTime {
			continue
		}

		if !meet.MaxAgeEnforced && meet.TimeStandard.MaxAgeTime != nil && meet.Age > *meet.TimeStandard.MaxAgeTime {
			searchAge = *meet.TimeStandard.MaxAgeTime
		} else if meet.MaxAgeEnforced && meet.Age > *meet.TimeStandard.MaxAgeTime {
			continue
		}

		standardDefinition := times.StandardDefinition{
			Age:      &searchAge,
			Gender:   swimmer.Swimmer.Gender.String,
			Course:   bestTime.Course,
			Style:    bestTime.Event.Style.Stroke,
			Distance: bestTime.Event.Distance,
		}

		standardTimeExample := times.StandardTime{
			TimeStandard: meet.TimeStandard,
			Definition:   standardDefinition,
		}
		standardTime, err := times.GetStandardTimeMeetByExample(standardTimeExample, meet.Season, uc.DB)
		if err != nil {
			log.Printf("times.%v", err)
		}

		if standardTime.Standard > 0 {
			standardTime.Difference = bestTime.BestTime - standardTime.Standard

			if bestTime.BestTime <= standardTime.Standard {
				standardTime.Percentage = 100
			} else {
				standardTime.Percentage = (standardTime.Standard * 100) / bestTime.BestTime
			}
			meet.StandardTime = *standardTime
			foundMeets = append(foundMeets, meet)
		}
	}

	recordExample := times.RecordDefinition{
		Age:      swimmer.Swimmer.AgeAt(time.Now()),
		Gender:   swimmer.Swimmer.Gender.String,
		Course:   bestTime.Course,
		Style:    bestTime.Event.Style.Stroke,
		Distance: bestTime.Event.Distance,
	}
	records, err := times.FindRecordsByExample(recordExample, uc.DB)
	if err != nil {
		log.Printf("times.%v", err)
	}
	for _, record := range records {
		record.Difference = bestTime.BestTime - record.Time

		if bestTime.BestTime <= record.Time {
			record.Percentage = 100
		} else {
			record.Percentage = (record.Time * 100) / bestTime.BestTime
		}
	}

	sort.SliceStable(foundMeets, func(i, j int) bool {
		return foundMeets[i].StandardTime.Difference < foundMeets[j].StandardTime.Difference
	})

	data := &swimmerBestTimeData{
		BaseTemplateData: uc.BaseTemplateData,
		Meets:            foundMeets,
		Records:          records,
		SwimmerBestTime:  bestTime,
		SessionData:      sessionData,
	}

	html := utils.GetTemplateWithFunctions("base", "swimmer-besttime", template.FuncMap{
		"Title":              utils.Title,
		"FormatMilliseconds": utils.FormatMilliseconds,
		"Lowercase":          utils.Lowercase,
		"Abs":                utils.Abs,
	})
	if err := html.Execute(res, data); err != nil {
		log.Printf("Error loading the swimmer's best time: %v", err)
	}
}

func (uc *Controller) ProfileSwimmersBestTimeBenchmarkView(res http.ResponseWriter, req *http.Request, sessionData *storage.SessionData) {
	id := req.URL.Query().Get(":id")
	swimmerId, _ := strconv.ParseInt(id, 10, 64)
	swimmer := FindSwimmerByID(swimmerId, uc.DB)

	course := req.URL.Query().Get("course")
	if course == "" {
		course = swimming.DefaultCourse
	}

	timeStandardID, _ := strconv.ParseInt(req.URL.Query().Get("standard"), 10, 64)
	timeStandard, err := times.FindTimeStandard(timeStandardID, uc.DB)
	if err != nil || timeStandard == nil {
		timeStandard = &times.TimeStandard{
			ID:   timeStandardID,
			Open: false,
		}
	}

	var meet *times.Meet
	meetID, _ := strconv.ParseInt(req.URL.Query().Get("meet"), 10, 64)
	meet = times.GetMeet(meetID, uc.DB)
	if meet == nil {
		meet = &times.Meet{
			ID: 0,
		}
	}

	currentSwimSeason, err := times.GetLatestSwimSeason(uc.DB)
	if err != nil {
		log.Printf("ProfileSwimmersBestTimeBenchmarkView: %v", err)
	}

	timeStandards, err := times.FindTimeStandards(*currentSwimSeason, swimmer.Swimmer.Team.Jurisdiction, uc.DB)
	if err != nil {
		log.Printf("ProfileSwimmersBestTimeBenchmarkView: %v", err)
	}

	meets, err := times.FindMeetsByTimeStandard(*timeStandard, uc.DB)
	if err != nil {
		log.Printf("ProfileSwimmersBestTimeBenchmarkView: %v", err)
	}

	bestTimes, err := findSwimmerBestTimes(swimmer, course, uc.DB)
	if err != nil {
		log.Printf("ProfileSwimmerView: %v", err)
	}

	// Age is the swimmer's age by default, but it can be changed by the user
	// to see the benchmark for a different age.
	minimum, maximum := times.FindMinAndMaxStandardAges(timeStandard, uc.DB)

	swimmerAge := swimmer.Swimmer.AgeAt(time.Now())
	min := swimmerAge
	age, err := strconv.ParseInt(req.URL.Query().Get("age"), 10, 64)
	if err != nil || age < swimmerAge {
		age = swimmerAge
	}

	if meet.ID != 0 {
		meetAge := swimmer.Swimmer.AgeAt(meet.AgeDate)
		if age < meetAge {
			age = meetAge
			min = meetAge
		}
	}

	var ages []int64
	if !timeStandard.Open {
		for i := min; i <= maximum; i++ {
			ages = append(ages, i)
		}
	} else {
		age = swimmerAge
	}

	queryAge := age
	if age < minimum {
		queryAge = minimum
	} else if age > maximum {
		queryAge = maximum
	}

	// If the meet is not in the list, reset the meet to the default.
	resetMeet := true
	for _, m := range meets {

		if m.ID == meet.ID {
			resetMeet = false
			break
		}
	}
	if resetMeet {
		meet = &times.Meet{
			ID: 0,
		}
		queryAge = swimmerAge
		age = swimmerAge
	}

	standardTimes, err := times.FindStandardTimesBySwimmer(swimmer.Swimmer, course, queryAge, *timeStandard, uc.DB)
	if err != nil {
		log.Printf("ProfileSwimmersBestTimeBenchmarkView: %v", err)
	}

	timeBenchmarks := make(map[string]*timeBenchmarkData)
	for _, standardTime := range standardTimes {
		course := standardTime.Definition.Course
		stroke := standardTime.Definition.Style
		distance := standardTime.Definition.Distance
		key := fmt.Sprintf("%s-%s-%d", course, stroke, distance)

		timeBenchmarks[key] = &timeBenchmarkData{
			StandardTime: standardTime.Standard,
		}
	}

	for _, bestTime := range bestTimes {
		course := bestTime.Course
		stroke := bestTime.Event.Style.Stroke
		distance := bestTime.Event.Distance
		key := fmt.Sprintf("%s-%s-%d", course, stroke, distance)

		if timeBenchmark, ok := timeBenchmarks[key]; ok {
			timeBenchmark.BestTime = bestTime.BestTime
			timeBenchmark.Difference = utils.Abs(bestTime.BestTime - timeBenchmark.StandardTime)
			timeBenchmark.Qualified = bestTime.BestTime <= timeBenchmark.StandardTime
			timeBenchmark.Considered = bestTime.BestTime <= int64(float64(timeBenchmark.StandardTime)*float64(1.05))
		}
	}

	data := &swimmerBestTimeBenchmarkData{
		Age:              age,
		Ages:             ages,
		BaseTemplateData: uc.BaseTemplateData,
		BestTimes:        bestTimes,
		Course:           course,
		Meet:             meet,
		Meets:            meets,
		SessionData:      sessionData,
		TimeBenchmarks:   timeBenchmarks,
		Swimmer:          swimmer,
		TimeStandard:     *timeStandard,
		TimeStandards:    timeStandards,
	}

	html := utils.GetTemplateWithFunctions("base", "profile-swimmer-benchmark", template.FuncMap{
		"Title":              utils.Title,
		"FormatMilliseconds": utils.FormatMilliseconds,
	})
	if err := html.Execute(res, data); err != nil {
		log.Printf("Error loading the swimmer's benchmark: %v", err)
	}
}

func (uc *Controller) SwimmerBestTimeFormView(res http.ResponseWriter, req *http.Request, sessionData *storage.SessionData) {
	id := req.URL.Query().Get(":id")
	swimmerId, _ := strconv.ParseInt(id, 10, 64)
	swimmer := FindSwimmerByID(swimmerId, uc.DB)

	id = req.URL.Query().Get(":bestId")
	bestId, _ := strconv.ParseInt(id, 10, 64)
	bestTime := findSwimmerBestTime(swimmer, bestId, uc.DB)

	events, err := findSwimmerMissingBestTimes(swimmer, swimming.DefaultCourse, uc.DB)
	if err != nil {
		log.Printf("home.events.%v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	data := &swimmerBestTimeData{
		BaseTemplateData: uc.BaseTemplateData,
		SwimmerBestTime:  bestTime,
		SessionData:      sessionData,
		Events:           events,
		Swimmer:          swimmer,
		Course:           swimming.DefaultCourse,
	}

	if bestTime != nil {
		data.Course = bestTime.Course
		data.Minute, data.Second, data.Millisecond = utils.FromMiliseconds(bestTime.BestTime)
	}

	html := utils.GetTemplateWithFunctions("base", "swimmer-besttime-form", template.FuncMap{
		"Title": utils.Title,
	})
	if err := html.Execute(res, data); err != nil {
		log.Printf("Error loading the swimmer best time form: %v", err)
	}
}

func (uc *Controller) SwimmerBestTimeForm(res http.ResponseWriter, req *http.Request, sessionData *storage.SessionData) {
	err := req.ParseForm()
	if err != nil {
		log.Print(err)
	}

	id := req.URL.Query().Get(":id")
	swimmerId, _ := strconv.ParseInt(id, 10, 64)
	swimmer := FindSwimmerByID(swimmerId, uc.DB)

	id = req.URL.Query().Get(":bestId")
	bestId, _ := strconv.ParseInt(id, 10, 64)
	swimmerBestTime := findSwimmerBestTime(swimmer, bestId, uc.DB)

	minute, _ := strconv.Atoi(req.PostForm.Get("minute"))
	second, _ := strconv.Atoi(req.PostForm.Get("second"))
	millisecond, _ := strconv.Atoi(req.PostForm.Get("millisecond"))
	bestTime := utils.ToMiliseconds(minute, second, millisecond)

	var eventId int64
	var course string
	if swimmerBestTime == nil {
		eventId, _ = strconv.ParseInt(req.PostForm.Get("event"), 10, 64)
		course = req.PostForm.Get("course")

		event := swimming.Event{
			ID: eventId,
		}
		swimmerBestTime = &SwimmerBestTime{
			Swimmer:  *swimmer,
			Event:    event,
			Course:   course,
			BestTime: bestTime,
		}
	} else {
		eventId = swimmerBestTime.Event.ID
		course = swimmerBestTime.Course
		swimmerBestTime.BestTime = bestTime
	}

	data := &swimmerBestTimeData{
		BaseTemplateData: uc.BaseTemplateData,
		SessionData:      sessionData,
		Event:            eventId,
		Course:           course,
		Minute:           minute,
		Second:           second,
		Millisecond:      millisecond,
		Swimmer:          swimmer,
		SwimmerBestTime:  swimmerBestTime,
	}

	if !data.valid() {
		log.Printf("Error saving the best time: %v", err)
		html := utils.GetTemplateWithFunctions("base", "swimmer-besttime-form", template.FuncMap{
			"Title": utils.Title,
		})
		data.Events, err = findSwimmerMissingBestTimes(swimmer, swimming.DefaultCourse, uc.DB)
		if err != nil {
			log.Printf("home.events.%v", err)
			http.Error(res, err.Error(), http.StatusInternalServerError)
		}
		err = html.Execute(res, data)
		if err != nil {
			log.Print(err)
		}
		return
	}

	_, err = saveSwimmerBestTime(swimmerBestTime, uc.DB)
	if err != nil {
		log.Printf("Error saving the best time: %v", err)
		html := utils.GetTemplateWithFunctions("base", "swimmer-besttime-form", template.FuncMap{
			"Title": utils.Title,
		})
		data.Error = `Due to an internal error, it was not possible to save
			your best time at this moment. Please, trying again later.
			Thank you for your understanding.`
		data.Events, err = findSwimmerMissingBestTimes(swimmer, swimming.DefaultCourse, uc.DB)
		if err != nil {
			log.Printf("home.events.%v", err)
			http.Error(res, err.Error(), http.StatusInternalServerError)
		}
		err = html.Execute(res, data)
		if err != nil {
			log.Print(err)
		}
		return
	}

	if sessionData.Role == "SWIMMER" {
		http.Redirect(res, req, fmt.Sprintf("/profile/swimmers/?course=%s", course), http.StatusSeeOther)
	} else {
		http.Redirect(res, req, fmt.Sprintf("/profile/swimmers/%d/?course=%s", swimmer.ID, course), http.StatusSeeOther)
	}
}

func (uc *Controller) SwimmerFormSearch(res http.ResponseWriter, req *http.Request, sessionData *storage.SessionData) {
	err := req.ParseForm()
	if err != nil {
		log.Print(err)
	}

	data := &swimmerData{
		SessionData:      sessionData,
		BaseTemplateData: uc.BaseTemplateData,
		Email:            strings.ToLower(strings.TrimSpace(req.PostForm.Get("email"))),
	}

	data.Jurisdictions, err = swimming.FindJurisdictionsByLevel(swimming.JurisdictionLevelRegion, uc.DB)
	if err != nil {
		log.Printf("Error loading jurisdictions: %v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	// Validates email
	if !messaging.IsEmailAddressValid(data.Email) {
		log.Printf("Invalid email address: %v", data.Email)
		data.ErrorEmail = "Invalid email address."
	}

	parent := FindUserAccountByEmail(sessionData.Email, uc.DB)
	data.FoundSwimmers, err = findLinkableSwimmerByEmail(data.Email, parent, uc.DB)
	if err != nil {
		log.Printf("Error finding swimmers: %v", err)
	}

	html := utils.GetTemplate("base", "swimmer-form")
	if err := html.Execute(res, data); err != nil {
		log.Printf("Error loading the swimmer form: %v", err)
	}
}

func (uc *Controller) SwimmerLinkForm(res http.ResponseWriter, req *http.Request, sessionData *storage.SessionData) {
	err := req.ParseForm()
	if err != nil {
		log.Print(err)
	}

	parent := FindUserAccountByEmail(sessionData.Email, uc.DB)

	selectedSwimmers := req.Form["swimmers"]
	var swimmers []*UserSwimmer
	for _, swimmerIDStr := range selectedSwimmers {
		swimmerID, err := strconv.ParseInt(swimmerIDStr, 10, 64)
		if err != nil {
			log.Printf("Invalid swimmer ID: %v", swimmerID)
		}
		swimmer := FindSwimmerByID(swimmerID, uc.DB)
		swimmers = append(swimmers, swimmer)
	}

	context := &swimmerData{
		FoundSwimmers:    swimmers,
		SessionData:      sessionData,
		BaseTemplateData: uc.BaseTemplateData,
	}

	if err := linkSwimmersToParent(parent, context.FoundSwimmers, uc.DB); err != nil {
		log.Printf("Error linking swimmers to parent: %v", err)
		html := utils.GetTemplate("base", "swimmer-form")
		context.Error = `Due to an internal error, it was not possible to link
			the swimmer wity your account at this moment. Please, trying again later.`
		err = html.Execute(res, context)
		if err != nil {
			log.Print(err)
		}
		return
	}

	http.Redirect(res, req, "/profile/", http.StatusSeeOther)
}

func (uc *Controller) authenticate(email, password, ipAddress string, humanScore float32) (*UserAccount, SignInAttempt) {
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

func (uc *Controller) SignOut(res http.ResponseWriter, req *http.Request) {
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
		reCaptchaScore, err := strconv.ParseFloat(fmt.Sprintf("%v", reCaptchaResBody["score"]), 32)
		if err != nil {
			log.Printf("Error parsing reCaptcha score: %v", err)
		}

		log.Printf("Success: %v, Score: %v", reCaptchaResBody["success"], reCaptchaResBody["score"])
		return float32(reCaptchaScore)
	}
	return 0
}

func (uc *Controller) addUserToSession(userAccount *UserAccount, res http.ResponseWriter, req *http.Request) error {
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

func (uc *Controller) addSwimmerToSession(swimmer *swimming.Swimmer, res http.ResponseWriter, req *http.Request) error {
	if err := storage.AddSessionEntry(res, req, "profile", "gender", swimmer.Gender.String); err != nil {
		return err
	}

	if err := storage.AddSessionEntry(res, req, "profile", "birthDate", swimmer.BirthDate.Time.Format("2006-01-02")); err != nil {
		return err
	}

	return nil
}

func (uc *Controller) SaveEmailSettings(res http.ResponseWriter, req *http.Request) {
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

	if err := updateUserAccount(user, uc.DB); err != nil {
		log.Printf("Error updating user account: %v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
}

// GoogleSignIn handles Google One Tap authentication
func (uc *Controller) GoogleSignIn(res http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		http.Error(res, "Invalid form data", http.StatusBadRequest)
		return
	}

	credential := req.PostForm.Get("credential")
	if credential == "" {
		http.Error(res, "No credential provided", http.StatusBadRequest)
		return
	}

	// Verify and parse the JWT token
	googleUser, err := uc.verifyGoogleToken(credential)
	if err != nil {
		log.Printf("Error verifying Google token: %v", err)
		http.Error(res, "Invalid Google credential", http.StatusUnauthorized)
		return
	}

	// Check if user exists or create new user
	userAccount := FindUserAccountByEmail(strings.ToLower(googleUser.Email), uc.DB)

	if userAccount == nil {
		// Create new user account from Google profile
		userAccount = &UserAccount{
			Email:     strings.ToLower(googleUser.Email),
			FirstName: googleUser.GivenName,
			LastName:  googleUser.FamilyName,
			Role:      RoleParent, // Default role
			Password:  []byte{},   // No password for Google users
		}

		userAccount.ID, err = InsertUserAccount(userAccount, uc.DB)
		if err != nil {
			log.Printf("Error creating user from Google sign-in: %v", err)
			http.Error(res, "Error creating user account", http.StatusInternalServerError)
			return
		}
		log.Printf("Created new user account for Google user: %v", userAccount.Email)
	}

	// Sign in the user
	if err = uc.addUserToSession(userAccount, res, req); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Google user %v signed in successfully", userAccount.Email)

	// Redirect based on user role
	if userAccount.Role == RoleSwimmer {
		http.Redirect(res, req, "/profile/swimmers", http.StatusSeeOther)
	} else {
		http.Redirect(res, req, "/profile/", http.StatusSeeOther)
	}
}

// verifyGoogleToken validates the Google ID token
func (uc *Controller) verifyGoogleToken(tokenString string) (*GoogleIDToken, error) {
	// Parse and verify the token using MapClaims
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing algorithm
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Get the kid from the header
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("no kid in token header")
		}

		// Get Google's public key for verification
		return uc.getGooglePublicKey(kid)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %v", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Extract claims as map
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims format")
	}

	// Convert map claims to our GoogleIDToken struct
	googleToken := &GoogleIDToken{}

	if iss, ok := claims["iss"].(string); ok {
		googleToken.Iss = iss
	} else {
		return nil, fmt.Errorf("missing or invalid iss claim")
	}

	if sub, ok := claims["sub"].(string); ok {
		googleToken.Sub = sub
	} else {
		return nil, fmt.Errorf("missing or invalid sub claim")
	}

	if aud, ok := claims["aud"].(string); ok {
		googleToken.Aud = aud
	} else {
		return nil, fmt.Errorf("missing or invalid aud claim")
	}

	if exp, ok := claims["exp"].(float64); ok {
		googleToken.Exp = int64(exp)
	} else {
		return nil, fmt.Errorf("missing or invalid exp claim")
	}

	if iat, ok := claims["iat"].(float64); ok {
		googleToken.Iat = int64(iat)
	} else {
		return nil, fmt.Errorf("missing or invalid iat claim")
	}

	if email, ok := claims["email"].(string); ok {
		googleToken.Email = email
	} else {
		return nil, fmt.Errorf("missing or invalid email claim")
	}

	if emailVerified, ok := claims["email_verified"].(bool); ok {
		googleToken.EmailVerified = emailVerified
	} else {
		return nil, fmt.Errorf("missing or invalid email_verified claim")
	}

	// Optional fields
	if name, ok := claims["name"].(string); ok {
		googleToken.Name = name
	}

	if picture, ok := claims["picture"].(string); ok {
		googleToken.Picture = picture
	}

	if givenName, ok := claims["given_name"].(string); ok {
		googleToken.GivenName = givenName
	}

	if familyName, ok := claims["family_name"].(string); ok {
		googleToken.FamilyName = familyName
	}

	// Validate our custom claims using the Valid() method we defined
	if err := googleToken.Valid(); err != nil {
		return nil, fmt.Errorf("token validation failed: %v", err)
	}

	// Verify the audience matches your client ID
	expectedClientID := config.GetConfiguration().GetString(config.GoogleClientID)
	if googleToken.Aud != expectedClientID {
		return nil, fmt.Errorf("invalid audience: expected %s, got %s", expectedClientID, googleToken.Aud)
	}

	// Verify the issuer
	if googleToken.Iss != "https://accounts.google.com" && googleToken.Iss != "accounts.google.com" {
		return nil, fmt.Errorf("invalid issuer: %s", googleToken.Iss)
	}

	// Verify email is verified by Google
	if !googleToken.EmailVerified {
		return nil, fmt.Errorf("email not verified by Google")
	}

	return googleToken, nil
}

// getGooglePublicKey fetches and converts Google's public key for JWT verification
func (uc *Controller) getGooglePublicKey(kid string) (*rsa.PublicKey, error) {
	// Fetch Google's public keys
	resp, err := http.Get("https://www.googleapis.com/oauth2/v3/certs")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var keys GoogleKeys
	if err := json.Unmarshal(body, &keys); err != nil {
		return nil, err
	}

	// Find the key with matching kid
	for _, key := range keys.Keys {
		if key.Kid == kid {
			return uc.convertToRSAPublicKey(key)
		}
	}

	return nil, fmt.Errorf("key with kid %s not found", kid)
}

// convertToRSAPublicKey converts Google's JWK to RSA public key
func (uc *Controller) convertToRSAPublicKey(key GooglePublicKey) (*rsa.PublicKey, error) {
	// Decode the modulus
	nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
	if err != nil {
		return nil, err
	}

	// Decode the exponent
	eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
	if err != nil {
		return nil, err
	}

	// Convert to big integers
	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)

	return &rsa.PublicKey{
		N: n,
		E: int(e.Int64()),
	}, nil
}
