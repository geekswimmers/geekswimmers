package user

import (
	"geekswimmers/modules/swimming"
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

type profileData struct {
	BaseTemplateData *utils.BaseTemplateData
	BirthDate        string
	Email            string
	Error            string
	ErrorBirthDate   string
	ErrorEmail       string
	ErrorFirstName   string
	ErrorGender      string
	ErrorLastName    string
	FirstName        string
	Gender           string
	LastName         string
	Role             string
	SessionData      *storage.SessionData
	Swimmers         []*Swimmer
}

func (sud *profileData) errorHappened() bool {
	return len(sud.ErrorEmail) > 0 ||
		len(sud.ErrorFirstName) > 0 ||
		len(sud.ErrorLastName) > 0 ||
		len(sud.Error) > 0 ||
		len(sud.ErrorBirthDate) > 0 ||
		len(sud.ErrorGender) > 0
}

type swimmerData struct {
	Swimmer          *Swimmer
	BaseTemplateData *utils.BaseTemplateData
	BirthDate        string
	Email            string
	Error            string
	ErrorBirthDate   string
	ErrorEmail       string
	ErrorFirstName   string
	ErrorGender      string
	ErrorLastName    string
	FirstName        string
	Events           []*swimming.Event
	FoundSwimmers    []*Swimmer
	BestTimes        []*SwimmerBestTime
	Gender           string
	LastName         string
	SessionData      *storage.SessionData
}

func (ad *swimmerData) errorHappened() bool {
	return len(ad.ErrorFirstName) > 0 ||
		len(ad.ErrorLastName) > 0 ||
		len(ad.ErrorBirthDate) > 0 ||
		len(ad.ErrorGender) > 0
}

type swimmerBestTimeData struct {
	BaseTemplateData *utils.BaseTemplateData
	Course           string
	Error            string
	ErrorCourse      string
	ErrorEvent       string
	ErrorMinute      string
	ErrorSecond      string
	ErrorMillisecond string
	Event            string
	Minute           string
	Second           string
	Millisecond      string
	Swimmer          *Swimmer
	Events           []*swimming.Event
	SessionData      *storage.SessionData
}

func (sbt *swimmerBestTimeData) valid() bool {
	if len(sbt.Course) == 0 {
		sbt.ErrorCourse = "Course is required"
		return false
	}
	if len(sbt.Event) == 0 {
		sbt.ErrorEvent = "Event is required"
		return false
	}

	if len(sbt.Minute) == 0 {
		sbt.ErrorMinute = "Minute is required"
		return false
	}

	if len(sbt.Second) == 0 {
		sbt.ErrorSecond = "Second is required"
		return false
	}

	if len(sbt.Millisecond) == 0 {
		sbt.ErrorMillisecond = "Millisecond is required"
		return false
	}

	return true
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
