package swimming

import (
	"database/sql"
	"fmt"
	"time"
)

const (
	DefaultCourse = "SHORT"

	GenderFemale = "FEMALE"
	GenderMale   = "MALE"

	JurisdictionLevelCountry  = "COUNTRY"
	JurisdictionLevelProvince = "PROVINCE"
	JurisdictionLevelRegion   = "REGION"
	JurisdictionLevelCity     = "CITY"
	JurisdictionLevelClub     = "CLUB"
	JurisdictionLevelMeet     = "MEET"
)

type Style struct {
	ID          int64
	Stroke      string
	Description string
	Sequence    int64
}

type Instruction struct {
	ID          int64
	Style       *Style
	Instruction string
	Sequence    int64
}

type Event struct {
	ID       int64
	Distance int64
	Style    Style
	Course   string
}

type Jurisdiction struct {
	ID       int64
	Country  string
	Province *string
	Region   *string
	City     *string
	Meet     *string
	Club     *string

	// Transient
	Title    string
	SubTitle string
}

func (jurisdiction *Jurisdiction) SetTitle() {
	if jurisdiction.Meet != nil {
		jurisdiction.Title = *jurisdiction.Meet
	} else if jurisdiction.Club != nil {
		jurisdiction.Title = *jurisdiction.Club
	} else if jurisdiction.City != nil {
		jurisdiction.Title = *jurisdiction.City
	} else if jurisdiction.Region != nil {
		jurisdiction.Title = *jurisdiction.Region
	} else if jurisdiction.Province != nil {
		jurisdiction.Title = *jurisdiction.Province
	} else {
		jurisdiction.Title = jurisdiction.Country
	}
}

func (jurisdiction *Jurisdiction) SetSubTitle() {
	if jurisdiction.Meet != nil {
		jurisdiction.SubTitle = fmt.Sprintf("%v, %v, %v, %v - %v", *jurisdiction.Club, *jurisdiction.City, *jurisdiction.Region, *jurisdiction.Province, jurisdiction.Country)
	} else if jurisdiction.Club != nil {
		jurisdiction.SubTitle = fmt.Sprintf("%v, %v, %v - %v", *jurisdiction.City, *jurisdiction.Region, *jurisdiction.Province, jurisdiction.Country)
	} else if jurisdiction.City != nil {
		jurisdiction.SubTitle = fmt.Sprintf("%v, %v - %v", *jurisdiction.Region, *jurisdiction.Province, jurisdiction.Country)
	} else if jurisdiction.Region != nil {
		jurisdiction.SubTitle = fmt.Sprintf("%v - %v", *jurisdiction.Province, jurisdiction.Country)
	} else if jurisdiction.Province != nil {
		jurisdiction.SubTitle = jurisdiction.Country
	}
}

type Club struct {
	ID           sql.NullInt64
	FullName     string
	Acronym      string
	WebSite      string
	Jurisdiction Jurisdiction
}

type Swimmer struct {
	FirstName string
	LastName  string
	Gender    sql.NullString
	BirthDate sql.NullTime
	Club      *Club
}

func (swimmer *Swimmer) AgeAt(date time.Time) int64 {
	age := date.Year() - swimmer.BirthDate.Time.Year()
	if date.YearDay() < swimmer.BirthDate.Time.YearDay() {
		age--
	}
	return int64(age)
}
