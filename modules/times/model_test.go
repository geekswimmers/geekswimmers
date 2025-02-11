package times

import (
	"database/sql"
	"geekswimmers/modules/swimming"
	"geekswimmers/modules/user"
	"testing"
	"time"
)

func TestAgeAt(t *testing.T) {

	// Test normal case
	swimmer := user.UserSwimmer{
		Swimmer: &swimming.Swimmer{
			BirthDate: sql.NullTime{
				Time: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
	}
	date := time.Date(2010, 6, 15, 0, 0, 0, 0, time.UTC)
	expected := int64(20)

	age := swimmer.Swimmer.AgeAt(date)

	if age != expected {
		t.Errorf("Expected %d, got %d", expected, age)
	}

	// Test edge case - birthday later in year
	swimmer = user.UserSwimmer{
		Swimmer: &swimming.Swimmer{
			BirthDate: sql.NullTime{
				Time: time.Date(1990, 6, 15, 0, 0, 0, 0, time.UTC),
			},
		},
	}
	date = time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC)
	expected = 19

	age = swimmer.Swimmer.AgeAt(date)

	if age != expected {
		t.Errorf("Expected %d, got %d", expected, age)
	}
}
