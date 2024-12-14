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
	Username        string
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

type signUpViewData struct {
	ReCaptchaSiteKey    string
	BaseTemplateContext *utils.BaseTemplateContext
}

type signUpData struct {
	Email               string
	ErrorAgreed         string
	ErrorEmail          string
	ErrorUsername       string
	UsernameSignUp      string
	ReCaptchaSiteKey    string
	BaseTemplateContext *utils.BaseTemplateContext
	AcceptedCookies     bool
}

type passwordViewData struct {
	Confirmation        string
	Email               string
	BaseTemplateContext *utils.BaseTemplateContext
	AcceptedCookies     bool
}

type setNewPasswordData struct {
	Confirmation        string
	Email               string
	Error               string
	BaseTemplateContext *utils.BaseTemplateContext
	AcceptedCookies     bool
}

type resetPasswordData struct {
	Email               string
	BaseTemplateContext *utils.BaseTemplateContext
}

type signInViewData struct {
	ReCaptchaSiteKey    string
	BaseTemplateContext *utils.BaseTemplateContext
}

type signInData struct {
	Identifier          string
	Error               string
	ReCaptchaSiteKey    string
	Lock                bool
	BaseTemplateContext *utils.BaseTemplateContext
}
