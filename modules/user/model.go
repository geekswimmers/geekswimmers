package user

import (
	"database/sql"
	"geekswimmers/modules/swimming"
	"strings"
	"time"
)

const (
	GenderFemale = "FEMALE"
	GenderMale   = "MALE"

	FailedMatchIdentifier       = "IDENTIFIER"
	FailedMatchPassword         = "PASSWORD"
	FailedMatchHumanScore       = "HUMAN_SCORE"
	FailedMatchAttemptsExceeded = "ATTEMPTS_EXCEEDED"

	RoleAdmin    = "ADMIN"
	RoleSwimmer  = "SWIMMER"
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

type Swimmer struct {
	ID            int64
	FirstName     string
	LastName      string
	Gender        sql.NullString
	BirthDate     sql.NullTime
	UserAccount   *UserAccount
	UserAccountID sql.NullInt64

	// Transient
	LinkApproved bool
}

func (swimmer *Swimmer) AgeAt(date time.Time) int64 {
	age := date.Year() - swimmer.BirthDate.Time.Year()
	if date.YearDay() < swimmer.BirthDate.Time.YearDay() {
		age--
	}
	return int64(age)
}

type SwimmerBestTime struct {
	ID       int64
	Swimmer  *Swimmer
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
