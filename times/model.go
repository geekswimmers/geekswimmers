package times

import (
	"fmt"
	"geekswimmers/storage"
	"geekswimmers/utils"
	"time"
)

const (
	GenderFemale = "FEMALE"
	GenderMale   = "MALE"

	DefaultCourse = "SHORT"

	JurisdictionLevelCountry  = "COUNTRY"
	JurisdictionLevelProvince = "PROVINCE"
	JurisdictionLevelRegion   = "REGION"
	JurisdictionLevelCity     = "CITY"
	JurisdictionLevelClub     = "CLUB"
	JurisdictionLevelMeet     = "MEET"

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
	Age      int64
	Sequence int64
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

type RecordPoster struct {
	Placeholder string
	Field       string
	Holder      string
	Time        int64
	CoordX      int64
	CoordY      int64
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

type benchmaskTimeViewData struct {
	Distance         int64
	Course           string
	Style            string
	Meets            []*Meet
	FormatedTime     string
	Records          []Record
	BaseTemplateData *utils.BaseTemplateData
	SessionData      *storage.SessionData
}

type timeStandardsViewData struct {
	SwimSeason       *SwimSeason
	SwimSeasons      []*SwimSeason
	TimeStandards    []*TimeStandard
	BaseTemplateData *utils.BaseTemplateData
	SessionData      *storage.SessionData
}

type timeStandardViewData struct {
	Age                int64
	Gender             string
	Course             string
	TimeStandard       *TimeStandard
	StandardTimes      []*StandardTime
	Ages               []int64
	LatestTimeStandard *TimeStandard
	Meets              []*Meet
	BaseTemplateData   *utils.BaseTemplateData
	SessionData        *storage.SessionData
}

type recordsListViewData struct {
	RecordSets       []*RecordSet
	BaseTemplateData *utils.BaseTemplateData
	SessionData      *storage.SessionData
}

type recordsViewData struct {
	Age              int64
	AgeRange         string
	AgeRanges        []*RecordDefinition
	Gender           string
	Course           string
	RecordSet        *RecordSet
	RecordDefinition RecordDefinition
	Records          []Record
	BaseTemplateData *utils.BaseTemplateData
	SessionData      *storage.SessionData
}

type recordHistoryViewData struct {
	RecordDefinition *RecordDefinition
	RecordSet        RecordSet
	Records          []*Record
	Jurisdiction     Jurisdiction
	BaseTemplateData *utils.BaseTemplateData
	SessionData      *storage.SessionData
}

type standardsEventViewData struct {
	Age              int64
	Ages             []int64
	Course           string
	Distance         int64
	Event            string
	Gender           string
	Style            string
	StandardTimes    []*StandardTime
	BaseTemplateData *utils.BaseTemplateData
	SessionData      *storage.SessionData
}

type clubRecordsReportData struct {
	Records    []*RecordPoster
	LastUpdate time.Time
}
