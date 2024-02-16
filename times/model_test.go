package times

import (
	"testing"
	"time"
)

func TestAgeAt(t *testing.T) {

	// Test normal case
	swimmer := Swimmer{
		BirthDate: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	date := time.Date(2010, 6, 15, 0, 0, 0, 0, time.UTC)
	expected := int64(20)

	age := swimmer.AgeAt(date)

	if age != expected {
		t.Errorf("Expected %d, got %d", expected, age)
	}

	// Test edge case - birthday later in year
	swimmer = Swimmer{
		BirthDate: time.Date(1990, 6, 15, 0, 0, 0, 0, time.UTC),
	}
	date = time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC)
	expected = 19

	age = swimmer.AgeAt(date)

	if age != expected {
		t.Errorf("Expected %d, got %d", expected, age)
	}
}
