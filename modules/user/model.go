package user

import (
	"database/sql"
	"geekswimmers/modules/swimming"
	"strings"
	"time"
)

const (
	FailedMatchIdentifier       = "IDENTIFIER"
	FailedMatchPassword         = "PASSWORD"
	FailedMatchHumanScore       = "HUMAN_SCORE"
	FailedMatchAttemptsExceeded = "ATTEMPTS_EXCEEDED"

	RoleAdmin   = "ADMIN"
	RoleSwimmer = "SWIMMER"
	RoleParent  = "PARENT"

	StatusSucceed = "SUCCEED"
	StatusFailed  = "FAILED"
)

type UserAccount struct {
	ID              int64
	Email           string
	Password        []byte
	FirstName       string
	LastName        string
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

type UserSwimmer struct {
	ID            int64
	Swimmer       *swimming.Swimmer
	UserAccount   *UserAccount
	UserAccountID sql.NullInt64

	// Transient
	LinkApproved bool
}

type SwimmerBestTime struct {
	ID       int64
	Swimmer  UserSwimmer
	Event    swimming.Event
	Course   string
	BestTime int64
	Updated  time.Time
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
