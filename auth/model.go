package auth

import (
	"fmt"
	"strings"
	"time"
)

const (
	FailedMatchIdentifier = "IDENTIFIER"
	FailedMatchPassword   = "PASSWORD"
	FailedMatchHumanScore = "HUMAN_SCORE"

	RoleAdmin    = "ADMIN"
	RoleUser     = "USER"
	RoleSwimmer  = "SWIMMER"
	RoleParent   = "PARENT"
	RoleCoach    = "COACH"
	RoleOfficial = "OFFICIAL"

	StatusSucceed = "SUCCEED"
	StatusFailed  = "FAILED"
)

type UserAccount struct {
	ID                int64      `json:"id"`
	Email             string     `json:"-"`
	Username          string     `json:"username"`
	Password          []byte     `json:"-"`
	FirstName         *string    `json:"-"`
	LastName          *string    `json:"-"`
	HumanScore        float32    `json:"-"`
	Confirmation      *string    `json:"-"`
	Created           time.Time  `json:"-"`
	Modified          time.Time  `json:"-"`
	SignOff           *time.Time `json:"-"`
	SignOffFeedback   *string    `json:"-"`
	NotificationPromo bool       `json:"-"` // User agrees to receive promotional emails in addition to notification emails.
	Role              string     `json:"-"`
}

func (userAccount *UserAccount) Name() string {
	var name string

	if userAccount.FirstName != nil && len(*userAccount.FirstName) > 0 {
		name = *userAccount.FirstName
	}

	if userAccount.LastName != nil && len(*userAccount.LastName) > 0 {
		name = fmt.Sprintf("%s %s", name, *userAccount.LastName)
	}

	name = strings.TrimSpace(name)

	if len(name) > 0 {
		return name
	}

	if len(userAccount.Username) > 0 {
		return fmt.Sprintf("@%s", userAccount.Username)
	}

	return ""
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
