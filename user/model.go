package user

import (
	"database/sql"
	"strings"
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
	Gender          sql.NullString
	BirthDate       sql.NullTime
	HumanScore      float32
	Confirmation    *string
	Created         time.Time
	Modified        time.Time
	SignOff         *time.Time
	SignOffFeedback *string
	PromotionalMsg  bool
	Role            string
}

func (ua *UserAccount) CleanEmail() string {
	email := ua.Email
	email = strings.ToLower(email)
	email = strings.TrimSpace(email)
	return email
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
