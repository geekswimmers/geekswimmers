package swimming

import (
	"database/sql"
	"time"
)

const (
	DefaultCourse = "SHORT"

	GenderFemale = "FEMALE"
	GenderMale   = "MALE"
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

type Swimmer struct {
	FirstName string
	LastName  string
	Gender    sql.NullString
	BirthDate sql.NullTime
}

func (swimmer *Swimmer) AgeAt(date time.Time) int64 {
	age := date.Year() - swimmer.BirthDate.Time.Year()
	if date.YearDay() < swimmer.BirthDate.Time.YearDay() {
		age--
	}
	return int64(age)
}
