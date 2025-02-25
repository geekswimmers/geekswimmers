package times

import (
	"database/sql"
	"geekswimmers/modules/swimming"
	"testing"
	"time"
)

func TestAgeAt(t *testing.T) {

	// Test normal case
	swimmer := &swimming.Swimmer{
		BirthDate: sql.NullTime{
			Time: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	date := time.Date(2010, 6, 15, 0, 0, 0, 0, time.UTC)
	expected := int64(20)

	age := swimmer.AgeAt(date)

	if age != expected {
		t.Errorf("Expected %d, got %d", expected, age)
	}

	// Check age today
	age = swimmer.AgeAt(time.Now())
	if age != 35 {
		t.Errorf("Expected 25, got %d", age)
	}

	// Test edge case - birthday later in year
	swimmer = &swimming.Swimmer{
		BirthDate: sql.NullTime{
			Time: time.Date(1990, 6, 15, 0, 0, 0, 0, time.UTC),
		},
	}
	date = time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC)
	expected = 19

	age = swimmer.AgeAt(date)

	if age != expected {
		t.Errorf("Expected %d, got %d", expected, age)
	}
}
