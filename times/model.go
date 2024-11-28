package times

import (
	"fmt"
	"geekswimmers/utils"
	"time"
)

const (
	GenderFemale = "FEMALE"
	GenderMale   = "MALE"

	CourseShort = "SHORT"
	CourseLong  = "LONG"

	StrokeFree   = "FREESTYLE"
	StrokeBack   = "BACKSTROKE"
	StrokeBreast = "BREASTSTROKE"
	StrokeFly    = "BUTTERFLY"
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

	// Transient
	Title    string
	SubTitle string
}

func (jurisdiction *Jurisdiction) SetTitle(age int64) {
	if jurisdiction.ID == 0 {
		if age == 0 {
			jurisdiction.Title = "World Record"
		} else {
			jurisdiction.Title = "World Junior Record"
		}
		return
	}

	if jurisdiction.Meet != "" {
		jurisdiction.Title = jurisdiction.Meet
	} else if jurisdiction.Club != "" {
		jurisdiction.Title = jurisdiction.Club
	} else if jurisdiction.City != "" {
		jurisdiction.Title = jurisdiction.City
	} else if jurisdiction.Region != "" {
		jurisdiction.Title = jurisdiction.Region
	} else if jurisdiction.Province != "" {
		jurisdiction.Title = jurisdiction.Province
	} else {
		jurisdiction.Title = jurisdiction.Country
	}
}

func (jurisdiction *Jurisdiction) SetSubTitle() {
	if jurisdiction.Meet != "" {
		jurisdiction.SubTitle = fmt.Sprintf("%s, %s, %s, %s - %s", jurisdiction.Club, jurisdiction.City, jurisdiction.Region, jurisdiction.Province, jurisdiction.Country)
	} else if jurisdiction.Club != "" {
		jurisdiction.SubTitle = fmt.Sprintf("%s, %s, %s - %s", jurisdiction.City, jurisdiction.Region, jurisdiction.Province, jurisdiction.Country)
	} else if jurisdiction.City != "" {
		jurisdiction.SubTitle = fmt.Sprintf("%s, %s - %s", jurisdiction.Region, jurisdiction.Province, jurisdiction.Country)
	} else if jurisdiction.Region != "" {
		jurisdiction.SubTitle = fmt.Sprintf("%s - %s", jurisdiction.Province, jurisdiction.Country)
	} else if jurisdiction.Province != "" {
		jurisdiction.SubTitle = jurisdiction.Country
	}
}

type Source struct {
	Title string
	Link  string
}

type TimeStandard struct {
	ID           int64
	Season       SwimSeason
	Name         string
	MinAgeTime   *int64
	MaxAgeTime   *int64
	Jurisdiction Jurisdiction
	Open         bool
	Source       Source
	Previous     *TimeStandard
	Benchmark    bool
}

type StandardTime struct {
	TimeStandard TimeStandard
	Age          int64
	Gender       string
	Course       string
	Style        string
	Distance     int64
	Standard     int64

	// Transient
	Difference int64
	Percentage int64
}

type Meet struct {
	ID             int64
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
	Style    string
	Distance int64

	// Transient
	Age int64
}

func (definition *RecordDefinition) AgeRange() string {
	if definition.MinAge != nil && definition.MaxAge != nil {
		if *definition.MinAge == *definition.MaxAge {
			return fmt.Sprintf("%d", *definition.MinAge)
		}
		return fmt.Sprintf("%d-%d", *definition.MinAge, *definition.MaxAge)
	}
	if definition.MinAge != nil {
		return fmt.Sprintf("%d-Over", *definition.MinAge)
	}
	if definition.MaxAge != nil {
		return fmt.Sprintf("%d-Under", *definition.MaxAge)
	}

	return "All"
}

type RecordSet struct {
	ID           int64
	Jurisdiction Jurisdiction
	Source       Source
}

type Record struct {
	ID         int64
	RecordSet  RecordSet
	Definition RecordDefinition
	Time       int64
	Year       *int64
	Month      *int64
	Holder     string

	// Transient
	Previous   []Record
	Difference int64
	Percentage int64
}

func (record *Record) MonthName() string {
	return utils.MonthName(*record.Month)
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
