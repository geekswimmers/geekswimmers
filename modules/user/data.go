package user

import (
	"database/sql"
	"geekswimmers/modules/swimming"
	"geekswimmers/modules/times"
	"geekswimmers/storage"
	"geekswimmers/utils"
	"geekswimmers/utils/messaging"
	"log"
	"time"
)

type signUpData struct {
	Agreed              string
	BaseTemplateData    *utils.BaseTemplateData
	BirthDate           string
	Club                int64
	Email               string
	Error               string
	ErrorAgreed         string
	ErrorBirthDate      string
	ErrorClub           string
	ErrorEmail          string
	ErrorFirstName      string
	ErrorGender         string
	ErrorLastName       string
	ErrorRole           string
	FirstName           string
	Gender              string
	Jurisdiction        int64
	Jurisdictions       []*swimming.Jurisdiction
	LastName            string
	ReCaptchaSiteKey    string
	Role                string
	SessionData         *storage.SessionData
	ExistingUserAccount *UserAccount
	UserAccountExists   bool
}

func (sud *signUpData) createUserAccount() *UserAccount {
	userAccount := &UserAccount{
		FirstName: sud.FirstName,
		LastName:  sud.LastName,
		Email:     sud.Email,
		Role:      sud.Role,
	}

	return userAccount
}

func (sud *signUpData) createSwimmer(userAccount *UserAccount) *UserSwimmer {
	swimmer := &UserSwimmer{
		Swimmer: &swimming.Swimmer{
			FirstName: sud.FirstName,
			LastName:  sud.LastName,
			Club: &swimming.Club{
				ID: sql.NullInt64{Int64: sud.Club, Valid: true},
			},
		},
		UserAccount: userAccount,
	}

	birthDate, err := time.Parse("2006-01-02", sud.BirthDate)
	if err != nil {
		log.Printf("Invalid birth date: %v", sud.BirthDate)
		return nil
	}

	swimmer.Swimmer.BirthDate = sql.NullTime{
		Time: birthDate,
	}

	swimmer.Swimmer.Gender = sql.NullString{
		String: sud.Gender,
	}

	return swimmer
}

func (sud *signUpData) valid() bool {
	valid := true

	// Validates firstName
	if sud.FirstName == "" {
		log.Printf("Invalid first name: %v", sud.FirstName)
		sud.ErrorFirstName = "First Name is empty."
		valid = false
	}

	// Validates lastName
	if sud.LastName == "" {
		log.Printf("Invalid last name: %v", sud.LastName)
		sud.ErrorLastName = "Last Name is empty."
		valid = false
	}

	// Validates email
	if !messaging.IsEmailAddressValid(sud.Email) {
		log.Printf("Invalid email address: %v", sud.Email)
		sud.ErrorEmail = "Invalid email address."
		valid = false
	}

	if sud.ExistingUserAccount != nil {
		sud.ErrorEmail = "This email is already in use. Do you want to <a href='/auth/signin/'>sign in</a> instead?"
		valid = false
	}

	// Validates role
	if sud.Role == "" || (sud.Role != RoleParent && sud.Role != RoleSwimmer) {
		log.Printf("Invalid role: %v", sud.Role)
		sud.ErrorRole = "Select a role."
		valid = false
	}

	if sud.Role == RoleSwimmer {
		// Validates birthDate
		if sud.BirthDate == "" {
			log.Printf("Birth date is required.")
			sud.ErrorBirthDate = "Birth date is required."
			valid = false
		}
		if sud.BirthDate != "" {
			birthDate, err := time.Parse("2006-01-02", sud.BirthDate)
			if err != nil {
				log.Printf("Invalid birth date: %v", sud.BirthDate)
				sud.ErrorBirthDate = "Invalid birth date."
				valid = false
			}

			// Validates age
			swimmer := swimming.Swimmer{
				BirthDate: sql.NullTime{
					Time: birthDate,
				},
			}

			age := swimmer.AgeAt(time.Now())
			if age < 13 {
				log.Printf("Invalid age: %v", age)
				sud.ErrorBirthDate = "You must be at least 13 years old to use Geek Swimmers."
				valid = false
			}
		}

		// Validates gender
		if (sud.Gender == "" || (sud.Gender != swimming.GenderFemale && sud.Gender != swimming.GenderMale)) && sud.Role == RoleSwimmer {
			log.Printf("Invalid Gender: %v", sud.Gender)
			sud.ErrorGender = "Select your gender."
			valid = false
		}

		// Validates club
		if sud.Club == 0 {
			log.Printf("Invalid club: %v", sud.Club)
			sud.ErrorClub = "Select your club."
			valid = false
		}
	}

	// Validates terms agreement
	if sud.Agreed != "on" {
		sud.ErrorAgreed = "You have to agree with our terms before creating an account."
		valid = false
	}

	return valid
}

type passwordViewData struct {
	Confirmation     string
	Email            string
	Error            string
	BaseTemplateData *utils.BaseTemplateData
	SessionData      *storage.SessionData
}

type profileData struct {
	BaseTemplateData *utils.BaseTemplateData
	BirthDate        string
	Club             int64
	CurrentUser      *UserAccount
	CurrentSwimmer   *UserSwimmer
	Email            string
	Error            string
	ErrorBirthDate   string
	ErrorClub        string
	ErrorEmail       string
	ErrorFirstName   string
	ErrorGender      string
	ErrorLastName    string
	ExistingUser     *UserAccount
	FirstName        string
	Gender           string
	Jurisdiction     int64
	Jurisdictions    []*swimming.Jurisdiction
	LastName         string
	Role             string
	SessionData      *storage.SessionData
	Swimmers         []*UserSwimmer
}

func (sud *profileData) valid() bool {
	valid := true

	// Validates firstName
	if sud.FirstName == "" {
		log.Printf("Invalid first name: %v", sud.FirstName)
		sud.ErrorFirstName = "First Name is empty."
		valid = false
	}

	// Validates lastName
	if sud.LastName == "" {
		log.Printf("Invalid last name: %v", sud.LastName)
		sud.ErrorLastName = "Last Name is empty."
		valid = false
	}

	// Validates email
	if !messaging.IsEmailAddressValid(sud.Email) {
		log.Printf("Invalid email address: %v", sud.Email)
		sud.ErrorEmail = "Invalid email address."
		valid = false
	} else {
		if sud.CurrentUser.Email != sud.Email {
			if sud.ExistingUser != nil {
				sud.ErrorEmail = "This email is already in use. Do you want to <a href='/auth/signin/'>sign in</a> instead?"
				valid = false
			}
		}
	}

	if sud.CurrentUser.Role == RoleSwimmer && sud.CurrentSwimmer != nil {
		// Validates birthDate
		if sud.BirthDate == "" {
			log.Printf("Birth date is required.")
			sud.ErrorBirthDate = "Birth date is required."
			valid = false
		} else {
			_, err := time.Parse("2006-01-02", sud.BirthDate)
			if err != nil {
				log.Printf("Invalid birth date: %v", sud.BirthDate)
				sud.ErrorBirthDate = "Invalid birth date."
				valid = false
			}

			// Validates age
			age := sud.CurrentSwimmer.Swimmer.AgeAt(time.Now())
			if age < 13 {
				log.Printf("Invalid age: %v", age)
				sud.ErrorBirthDate = "You must be at least 13 years old to use Geek Swimmers."
				valid = false
			}
		}

		// Validates gender
		if sud.Gender == "" || (sud.Gender != swimming.GenderFemale && sud.Gender != swimming.GenderMale) {
			log.Printf("Invalid Gender: %v", sud.Gender)
			sud.ErrorGender = "Select your gender."
			valid = false
		}

		// Validates club
		if sud.Club == 0 {
			log.Printf("Invalid club: %v", sud.Club)
			sud.ErrorClub = "Select your club."
			valid = false
		}
	}

	return valid
}

type swimmerData struct {
	Swimmer          *UserSwimmer
	BaseTemplateData *utils.BaseTemplateData
	BirthDate        string
	Club             int64
	Email            string
	Error            string
	ErrorBirthDate   string
	ErrorClub        string
	ErrorEmail       string
	ErrorFirstName   string
	ErrorGender      string
	ErrorLastName    string
	FirstName        string
	Jurisdiction     int64
	Jurisdictions    []*swimming.Jurisdiction
	Events           []*swimming.Event
	FoundSwimmers    []*UserSwimmer
	BestTimes        []*SwimmerBestTime
	LinkRequests     []*ParentSwimmer
	ParentSwimmer    *ParentSwimmer
	Gender           string
	LastName         string
	SessionData      *storage.SessionData
}

func (sd *swimmerData) valid() bool {
	valid := true

	// Validates firstName
	if sd.FirstName == "" {
		log.Printf("Invalid first name: %v", sd.FirstName)
		sd.ErrorFirstName = "First Name is empty."
		valid = false
	}

	// Validates lastName
	if sd.LastName == "" {
		log.Printf("Invalid last name: %v", sd.LastName)
		sd.ErrorLastName = "Last Name is empty."
		valid = false
	}

	// Validates birthDate
	if sd.BirthDate == "" {
		log.Printf("Birth date is required.")
		sd.ErrorBirthDate = "Birth date is required."
		valid = false
	} else {
		_, err := time.Parse("2006-01-02", sd.BirthDate)
		if err != nil {
			log.Printf("Invalid birth date: %v", sd.BirthDate)
			sd.ErrorBirthDate = "Invalid birth date."
			valid = false
		}
	}

	// Validates gender
	if sd.Gender == "" || (sd.Gender != swimming.GenderFemale && sd.Gender != swimming.GenderMale) {
		log.Printf("Invalid Gender: %v", sd.Gender)
		sd.ErrorGender = "Select your gender."
		valid = false
	}

	// Validates club
	if sd.Club == 0 {
		log.Printf("Invalid club: %v", sd.Club)
		sd.ErrorClub = "Select your club."
		valid = false
	}

	return valid
}

func (sd *swimmerData) createSwimmer() *UserSwimmer {
	swimmer := &UserSwimmer{
		Swimmer: &swimming.Swimmer{
			FirstName: sd.FirstName,
			LastName:  sd.LastName,
			Club: &swimming.Club{
				ID: sql.NullInt64{Int64: sd.Club, Valid: true},
			},
		},
	}

	birthDate, err := time.Parse("2006-01-02", sd.BirthDate)
	if err != nil {
		log.Printf("Invalid birth date: %v", sd.BirthDate)
		return nil
	}

	swimmer.Swimmer.BirthDate = sql.NullTime{
		Time: birthDate,
	}

	swimmer.Swimmer.Gender = sql.NullString{
		String: sd.Gender,
	}

	return swimmer
}

type swimmerBestTimeData struct {
	BaseTemplateData *utils.BaseTemplateData
	SwimmerBestTime  *SwimmerBestTime
	Course           string
	Error            string
	ErrorCourse      string
	ErrorEvent       string
	ErrorMinute      string
	ErrorSecond      string
	ErrorMillisecond string
	Event            int64
	Minute           int
	Second           int
	Millisecond      int
	Swimmer          *UserSwimmer
	Events           []*swimming.Event
	Meets            []*times.Meet
	Records          []times.Record
	SessionData      *storage.SessionData
}

func (sbt *swimmerBestTimeData) valid() bool {
	if len(sbt.Course) == 0 {
		sbt.ErrorCourse = "Course is required"
		return false
	}
	if sbt.Event == 0 {
		sbt.ErrorEvent = "Event is required"
		return false
	}

	if sbt.Minute < 0 {
		sbt.ErrorMinute = "Minute is required"
		return false
	}

	if sbt.Second < 0 {
		sbt.ErrorSecond = "Second is required"
		return false
	}

	if sbt.Millisecond < 0 {
		sbt.ErrorMillisecond = "Millisecond is required"
		return false
	}

	return true
}

type swimmerBestTimeBenchmarkData struct {
	BaseTemplateData *utils.BaseTemplateData
	SessionData      *storage.SessionData
	TimeBenchmarks   map[string]*timeBenchmarkData
	Swimmer          *UserSwimmer
	BestTimes        []*SwimmerBestTime
	TimeStandard     times.TimeStandard
	TimeStandards    []*times.TimeStandard
	Meet             *times.Meet
	Meets            []*times.Meet
}

type timeBenchmarkData struct {
	StandardTime int64
	Difference   int64
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
