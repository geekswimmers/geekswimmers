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
	Name string
}

type TimeStandard struct {
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

	Difference int64
	Percentage int64
}

type Meet struct {
	Name    string
	AgeDate time.Time
}

type Swimmer struct {
	BirthDate time.Time
	Gender    string
}

func (swimmer *Swimmer) AgeAt(date time.Time) int {
	age := date.Year() - swimmer.BirthDate.Year()
	if date.YearDay() < swimmer.BirthDate.YearDay() {
		age--
	}
	return age
}
