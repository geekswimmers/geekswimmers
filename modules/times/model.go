package times

import (
	"fmt"
	"geekswimmers/modules/swimming"
	"geekswimmers/utils"
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
	Age      *int64
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
	Season         SwimSeason
	TimeStandard   TimeStandard
	MinAgeEnforced bool
	MaxAgeEnforced bool
	Organizer      swimming.Team

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
	Title        string
	Jurisdiction swimming.Jurisdiction
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

func (record *Record) AllHolders() string {
	if len(record.Previous) == 0 || record.Time != record.Previous[0].Time {
		return record.Holder
	}

	return fmt.Sprintf("%s, %s", record.Holder, record.Previous[0].Holder)
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
