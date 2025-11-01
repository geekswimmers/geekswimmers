package user

import (
	"database/sql"
	"fmt"
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

	ParentSwimmerApprovalPending   = "PENDING"
	ParentSwimmerApprovalAccepted  = "ACCEPTED"
	ParentSwimmerApprovalDismissed = "DISMISSED"
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
	LinkApproval  string
}

type ParentSwimmer struct {
	ID       int64
	Parent   *UserAccount
	Swimmer  *UserSwimmer
	Approval string

	// Transient
	Responsible bool
}

type SwimmerBestTime struct {
	ID       int64
	Swimmer  UserSwimmer
	Event    swimming.Event
	Course   string
	BestTime int64
	Updated  time.Time

	// Transient
	Points int
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

// GoogleIDToken represents the structure of a Google ID token
type GoogleIDToken struct {
	Iss           string `json:"iss"`
	Sub           string `json:"sub"`
	Aud           string `json:"aud"`
	Exp           int64  `json:"exp"`
	Iat           int64  `json:"iat"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
}

// Valid implements the jwt.Claims interface
func (g *GoogleIDToken) Valid() error {
	now := time.Now().Unix()

	// Check if token has expired
	if g.Exp < now {
		return fmt.Errorf("token has expired")
	}

	// Check if token was issued in the future (with some tolerance)
	if g.Iat > now+300 { // 5 minutes tolerance
		return fmt.Errorf("token issued in the future")
	}

	// Check required fields
	if g.Iss == "" || g.Sub == "" || g.Aud == "" {
		return fmt.Errorf("missing required claims")
	}

	return nil
}

// GooglePublicKey represents a Google public key
type GooglePublicKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// GoogleKeys represents Google's public keys response
type GoogleKeys struct {
	Keys []GooglePublicKey `json:"keys"`
}
