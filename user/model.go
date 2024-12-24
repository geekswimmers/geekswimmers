package user

import (
	"geekswimmers/storage"
	"geekswimmers/utils"
	"time"
)

const (
	FailedMatchIdentifier       = "IDENTIFIER"
	FailedMatchPassword         = "PASSWORD"
	FailedMatchHumanScore       = "HUMAN_SCORE"
	FailedMatchAttemptsExceeded = "ATTEMPTS_EXCEEDED"

	RoleAdmin    = "ADMIN"
	RoleAthlete  = "ATHLETE"
	RoleClub     = "CLUB"
	RoleCoach    = "COACH"
	RoleOfficial = "OFFICIAL"
	RoleParent   = "PARENT"

	StatusSucceed = "SUCCEED"
	StatusFailed  = "FAILED"
)

type UserAccount struct {
	ID              int64
	Email           string
	Password        []byte
	FirstName       string
	LastName        string
	Gender          string
	BirthDate       time.Time
	HumanScore      float32
	Confirmation    *string
	Created         time.Time
	Modified        time.Time
	SignOff         *time.Time
	SignOffFeedback *string
	PromotionalMsg  bool
	Role            string
}

type Family struct {
	ID     int64
	Member int64
	Main   bool
	Name   *string
}

type EmailMessage struct {
	ID        int64
	Recipient string
	Subject   string
	Body      string
	Username  string
	Sent      time.Time
}

type SignInAttempt struct {
	ID          int64
	Identifier  string
	HumanScore  float32
	Created     time.Time
	Status      string
	IPAddress   string
	FailedMatch string
}

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
