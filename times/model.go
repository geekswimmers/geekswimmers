package times

import (
	"database/sql"
	"fmt"
	"time"
)

const (
	GenderFemale = "FEMALE"
	GenderMale   = "MALE"

	CourseShort = "SHORT"
	CourseLong  = "LONG"

	StrokeFree   = "FREE"
	StrokeBack   = "BACK"
	StrokeBreast = "BREAST"
	StrokeFly    = "FLY"
	StrokeMedley = "MEDLEY"
)

type SwimSeason struct {
	ID        int64
	Name      string
	StartDate time.Time
	EndDate   time.Time
}

type Jurisdiction struct {
	ID       int64
	Country  string
	Province string
	Region   string
	City     string
	Meet     string
	Club     string
}

type TimeStandard struct {
	ID           int64
	Season       SwimSeason
	Name         string
	MinAgeTime   int64
	MaxAgeTime   int64
	Jurisdiction Jurisdiction
}

type StandardTime struct {
	TimeStandard TimeStandard
	Age          int64
	Gender       string
	Course       string
	Stroke       string
	Distance     int64
	Standard     int64

	// Transient
	Difference int64
	Percentage int64
}

type Meet struct {
	Name           string
	Course         string
	AgeDate        time.Time
	Season         SwimSeason
	TimeStandard   TimeStandard
	MinAgeEnforced bool
	MaxAgeEnforced bool

	// Transient
	Age          int64
	StandardTime StandardTime
}

type RecordDefinition struct {
	ID       int64
	MinAge   *int64
	MaxAge   *int64
	Gender   string
	Course   string
	Stroke   string
	Distance int64

	// Transient
	Age int64
}

type Record struct {
	ID           int64
	Jurisdiction Jurisdiction
	Definition   RecordDefinition
	Time         int64
	Date         sql.NullTime
	Holder       string

	// Transient
	Previous   []Record
	Title      string
	SubTitle   string
	Difference int64
	Percentage int64
}

func (record *Record) SetTitle() {
	if record.Jurisdiction.ID == 0 {
		if record.Definition.Age == 0 {
			record.Title = "World Record"
		} else {
			record.Title = "World Junior Record"
		}
		return
	}

	if record.Jurisdiction.Meet != "" {
		record.Title = record.Jurisdiction.Meet
	} else if record.Jurisdiction.Club != "" {
		record.Title = record.Jurisdiction.Club
	} else if record.Jurisdiction.City != "" {
		record.Title = record.Jurisdiction.City
	} else if record.Jurisdiction.Region != "" {
		record.Title = record.Jurisdiction.Region
	} else if record.Jurisdiction.Province != "" {
		record.Title = record.Jurisdiction.Province
	} else {
		record.Title = record.Jurisdiction.Country
	}
}

func (record *Record) SetSubTitle() {
	if record.Jurisdiction.Meet != "" {
		record.SubTitle = fmt.Sprintf("%s, %s, %s, %s - %s", record.Jurisdiction.Club, record.Jurisdiction.City, record.Jurisdiction.Region, record.Jurisdiction.Province, record.Jurisdiction.Country)
	} else if record.Jurisdiction.Club != "" {
		record.SubTitle = fmt.Sprintf("%s, %s, %s - %s", record.Jurisdiction.City, record.Jurisdiction.Region, record.Jurisdiction.Province, record.Jurisdiction.Country)
	} else if record.Jurisdiction.City != "" {
		record.SubTitle = fmt.Sprintf("%s, %s - %s", record.Jurisdiction.Region, record.Jurisdiction.Province, record.Jurisdiction.Country)
	} else if record.Jurisdiction.Region != "" {
		record.SubTitle = fmt.Sprintf("%s - %s", record.Jurisdiction.Province, record.Jurisdiction.Country)
	} else if record.Jurisdiction.Province != "" {
		record.SubTitle = record.Jurisdiction.Country
	}
}

type Swimmer struct {
	BirthDate time.Time
	Gender    string
}

func (swimmer *Swimmer) AgeAt(date time.Time) int64 {
	age := date.Year() - swimmer.BirthDate.Year()
	if date.YearDay() < swimmer.BirthDate.YearDay() {
		age--
	}
	return int64(age)
}
