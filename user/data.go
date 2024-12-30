package user

import (
	"geekswimmers/storage"
	"geekswimmers/utils"
)

type signUpData struct {
	BaseTemplateData *utils.BaseTemplateData
	BirthDate        string
	Email            string
	Error            string
	ErrorAgreed      string
	ErrorBirthDate   string
	ErrorEmail       string
	ErrorFirstName   string
	ErrorGender      string
	ErrorLastName    string
	ErrorRole        string
	FirstName        string
	Gender           string
	LastName         string
	ReCaptchaSiteKey string
	Role             string
	SessionData      *storage.SessionData
}

func (sud *signUpData) errorHappened() bool {
	return len(sud.ErrorEmail) > 0 ||
		len(sud.ErrorFirstName) > 0 ||
		len(sud.ErrorLastName) > 0 ||
		len(sud.ErrorAgreed) > 0 ||
		len(sud.ErrorRole) > 0 ||
		len(sud.ErrorBirthDate) > 0 ||
		len(sud.ErrorGender) > 0
}

type passwordViewData struct {
	Confirmation     string
	Email            string
	Error            string
	BaseTemplateData *utils.BaseTemplateData
	SessionData      *storage.SessionData
}

type setNewPasswordData struct {
	Confirmation     string
	Email            string
	Error            string
	BaseTemplateData *utils.BaseTemplateData
	SessionData      *storage.SessionData
}

type resetPasswordData struct {
	Email            string
	BaseTemplateData *utils.BaseTemplateData
}

type signInData struct {
	BaseTemplateData *utils.BaseTemplateData
	Error            string
	Identifier       string
	Lock             bool
	ReCaptchaSiteKey string
	SessionData      *storage.SessionData
}
