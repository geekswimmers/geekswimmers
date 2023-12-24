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
	ID        int64
	Name      string
	StartDate time.Time
	EndDate   time.Time
}

type TimeStandard struct {
	ID      int64
	Season  SwimSeason
	Name    string
	Summary string
}

type StandardTime struct {
	ID           int64
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
