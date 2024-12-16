package user

import (
	"geekswimmers/utils"
	"time"
)

const (
	FailedMatchIdentifier = "IDENTIFIER"
	FailedMatchPassword   = "PASSWORD"
	FailedMatchHumanScore = "HUMAN_SCORE"

	RoleAdmin    = "ADMIN"
	RoleAthlete  = "ATHLETE"
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

	// Transient
	Roles []*UserRole
}

type UserRole struct {
	ID          int64
	UserAccount UserAccount
	Role        string
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
	AcceptedCookies  bool
	BaseTemplateData *utils.BaseTemplateData
	Email            string
	ErrorAgreed      string
	ErrorEmail       string
	ErrorFirstName   string
	ErrorLastName    string
	FirstName        string
	LastName         string
	ReCaptchaSiteKey string
}

func (sud *signUpData) errorHappened() bool {
	return len(sud.ErrorEmail) > 0 || len(sud.ErrorFirstName) > 0 || len(sud.ErrorLastName) > 0 || len(sud.ErrorAgreed) > 0
}

type passwordViewData struct {
	Confirmation     string
	Email            string
	Error            string
	BaseTemplateData *utils.BaseTemplateData
	AcceptedCookies  bool
}

type setNewPasswordData struct {
	Confirmation     string
	Email            string
	Error            string
	BaseTemplateData *utils.BaseTemplateData
	AcceptedCookies  bool
}

type resetPasswordData struct {
	Email            string
	BaseTemplateData *utils.BaseTemplateData
}

type signInViewData struct {
	ReCaptchaSiteKey string
	BaseTemplateData *utils.BaseTemplateData
}

type signInData struct {
	Identifier       string
	Error            string
	ReCaptchaSiteKey string
	Lock             bool
	BaseTemplateData *utils.BaseTemplateData
}
