package times

import (
	"database/sql"
	"fmt"
	"geekswimmers/modules/swimming"
	"geekswimmers/utils"
	"strings"
	"time"
)

type SwimSeason struct {
	ID        int64
	Name      string
	StartDate time.Time
	EndDate   time.Time
}

type Source struct {
	Title string
	Link  string
}

type TimeStandard struct {
	ID           int64
	Name         string
	MinAgeTime   *int64
	MaxAgeTime   *int64
	Jurisdiction swimming.Jurisdiction
	Open         bool
	Source       Source
	Benchmark    bool
}

type StandardDefinition struct {
	ID       int64
	Age      sql.NullInt64
	Gender   string
	Course   string
	Style    string
	Distance int64
}

type StandardTime struct {
	ID           int64
	TimeStandard TimeStandard
	Definition   StandardDefinition
	Standard     int64
	UpdateDate   time.Time

	// Transient
	Difference int64
	Percentage int64
}

type Meet struct {
	ID             int64
	Name           string
	Course         string
	AgeDate        time.Time
	StartDate      time.Time
	EndDate        time.Time
	Website        sql.NullString
	Season         SwimSeason
	TimeStandard   TimeStandard
	MinAgeEnforced bool
	MaxAgeEnforced bool
	Organizer      swimming.Team

	// Transient
	Age          int64
	StandardTime StandardTime
}

func (meet *Meet) Duration() string {
	duration := meet.EndDate.Sub(meet.StartDate).Hours()/24 + 1

	if duration == 1 {
		return fmt.Sprintf("%v day", duration)
	}
	return fmt.Sprintf("%v days", duration)
}

type MeetEvent struct {
	ID     int64
	MeetID int64
	Event  swimming.Event
	Gender string
}

type MeetResult struct {
	ID         int64
	MeetEvent  MeetEvent
	Swimmer    swimming.Swimmer
	Team       swimming.Team
	ResultTime *int64
	DQ         bool
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

func (definition *RecordDefinition) AgeRangeValue() string {
	ageRange := definition.AgeRange()
	ageRange = strings.ReplaceAll(ageRange, "Over", "")
	ageRange = strings.ReplaceAll(ageRange, "Under", "")
	return ageRange
}

type RecordSet struct {
	ID           int64
	Title        string
	Jurisdiction swimming.Jurisdiction
	Source       Source
}

type Record struct {
	ID         int64
	RecordSet  RecordSet
	Definition RecordDefinition
	Swimmer    *swimming.Swimmer
	Time       int64
	Year       *int64
	Month      *int64
	Holder     string

	// Transient
	Previous   []Record
	Difference int64
	Percentage int64
}

func (record *Record) HolderName() string {
	if len(record.Swimmer.FirstName) > 0 || len(record.Swimmer.LastName) > 0 {
		return fmt.Sprintf("%s %s", record.Swimmer.FirstName, record.Swimmer.LastName)
	}
	return record.Holder
}

func (record *Record) AllHolders() string {
	if len(record.Previous) == 0 || record.Time != record.Previous[0].Time {
		return record.HolderName()
	}
	return fmt.Sprintf("%s, %s", record.HolderName(), record.Previous[0].HolderName())
}

func (record *Record) MonthName() string {
	return utils.MonthName(*record.Month)
}

type RecordPoster struct {
	Placeholder string
	Field       string
	Value       string
	CoordX      int64
	CoordY      int64
}
