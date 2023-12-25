package swimmers

import "time"

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
	ID   int64
	Name string
}

type TimeStandard struct {
	ID      int64
	Season  SwimSeason
	Name    string
	Summary string
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
	Name         string
	Course       string
	AgeDate      time.Time
	Season       SwimSeason
	TimeStandard TimeStandard

	// Transient
	Age          int64
	StandardTime StandardTime
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
