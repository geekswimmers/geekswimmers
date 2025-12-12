package swimming

import (
	"database/sql"
	"fmt"
	"time"
)

const (
	CourseShort   = "SHORT"
	CourseLong    = "LONG"
	DefaultCourse = CourseShort

	GenderFemale = "FEMALE"
	GenderMale   = "MALE"

	JurisdictionLevelCountry  = "COUNTRY"
	JurisdictionLevelProvince = "PROVINCE"
	JurisdictionLevelRegion   = "REGION"
	JurisdictionLevelCity     = "CITY"
	JurisdictionLevelTeam     = "TEAM"
	JurisdictionLevelMeet     = "MEET"

	StyleFreestyle    = "FREESTYLE"
	StyleBackstroke   = "BACKSTROKE"
	StyleBreaststroke = "BREASTSTROKE"
	StyleButterfly    = "BUTTERFLY"
	StyleMedley       = "MEDLEY"
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
	ID       sql.NullInt64
	World    sql.NullString
	Country  sql.NullString
	Province sql.NullString
	Region   sql.NullString
	City     sql.NullString
	Meet     sql.NullString
	Team     sql.NullString
}

func (jurisdiction *Jurisdiction) Title() string {
	if jurisdiction.Meet.Valid {
		return jurisdiction.Meet.String
	} else if jurisdiction.Team.Valid {
		return jurisdiction.Team.String
	} else if jurisdiction.City.Valid {
		return jurisdiction.City.String
	} else if jurisdiction.Region.Valid {
		return jurisdiction.Region.String
	} else if jurisdiction.Province.Valid {
		return jurisdiction.Province.String
	} else if jurisdiction.Country.Valid {
		return jurisdiction.Country.String
	} else if jurisdiction.World.Valid {
		return jurisdiction.World.String
	}

	return "None"
}

func (jurisdiction *Jurisdiction) SubTitle() string {
	if jurisdiction.Meet.Valid {
		return fmt.Sprintf("%v, %v, %v, %v - %v",
			jurisdiction.Team.String,
			jurisdiction.City.String,
			jurisdiction.Region.String,
			jurisdiction.Province.String,
			jurisdiction.Country.String)
	} else if jurisdiction.Team.Valid {
		return fmt.Sprintf("%v, %v, %v - %v",
			jurisdiction.City.String,
			jurisdiction.Region.String,
			jurisdiction.Province.String,
			jurisdiction.Country.String)
	} else if jurisdiction.City.Valid {
		return fmt.Sprintf("%v, %v - %v",
			jurisdiction.Region.String,
			jurisdiction.Province.String,
			jurisdiction.Country.String)
	} else if jurisdiction.Region.Valid {
		return fmt.Sprintf("%v - %v",
			jurisdiction.Province.String,
			jurisdiction.Country.String)
	} else if jurisdiction.Province.Valid || jurisdiction.Country.Valid {
		return jurisdiction.Country.String
	}

	return ""
}

type Team struct {
	ID           sql.NullInt64
	FullName     string
	Acronym      string
	WebSite      string
	Jurisdiction Jurisdiction
}

type Swimmer struct {
	ID          int64
	FirstName   string
	LastName    string
	Gender      sql.NullString
	BirthDate   sql.NullTime
	Team        *Team
	SwimRanking sql.NullString
	SwimCloud   sql.NullString
}

func (swimmer *Swimmer) AgeAt(date time.Time) int64 {
	age := date.Year() - swimmer.BirthDate.Time.Year()
	if date.YearDay() < swimmer.BirthDate.Time.YearDay() {
		age--
	}
	return int64(age)
}
