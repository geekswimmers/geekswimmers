package admin

import (
	"bufio"
	"geekswimmers/modules/swimming"
	"io"
	"strconv"
	"strings"
	"time"
)

// Field positions in HY3 records (0-indexed).
// These follow the HY-TEK Meet Manager HY3 file specification.
const (
	// B1 - Meet record
	hy3B1NameStart     = 2
	hy3B1NameEnd       = 30
	hy3B1CourseStart   = 30
	hy3B1CourseEnd     = 32
	hy3B1StartDateStart = 32
	hy3B1StartDateEnd   = 38
	hy3B1EndDateStart   = 38
	hy3B1EndDateEnd     = 44
	hy3B1LocationStart  = 44
	hy3B1LocationEnd    = 74
	hy3B1FacilityStart  = 74
	hy3B1FacilityEnd    = 104

	// C1 - Team record
	hy3C1AcronymStart = 2
	hy3C1AcronymEnd   = 7

	// D0 - Athlete record
	hy3D0LastNameStart        = 2
	hy3D0LastNameEnd          = 32
	hy3D0FirstNameStart       = 32
	hy3D0FirstNameEnd         = 47
	hy3D0GenderPos            = 47
	hy3D0BirthDateStart       = 48
	hy3D0BirthDateEnd         = 54
	hy3D0NationalNumberStart  = 54
	hy3D0NationalNumberEnd    = 74

	// E0 - Individual event
	hy3E0DistanceStart = 9
	hy3E0DistanceEnd   = 13
	hy3E0StrokePos     = 13

	// E1 - Individual result
	hy3E1FinalTimeStart = 10
	hy3E1FinalTimeEnd   = 18
	hy3E1TimeCodePos    = 18
)

type hy3ParsedFile struct {
	MeetName  string
	Course    string
	StartDate time.Time
	EndDate   time.Time
	Location  string
	Facility  string
	Results   []hy3RawResult
}

type hy3RawResult struct {
	TeamAcronym    string
	FirstName      string
	LastName       string
	BirthDate      time.Time
	Gender         string
	NationalNumber string
	Stroke         string
	Distance       int64
	Time           string
	DQ             bool
}

func parseHY3(r io.Reader) hy3ParsedFile {
	var parsed hy3ParsedFile

	var currentTeam string
	var currentFirstName, currentLastName string
	var currentBirthDate time.Time
	var currentGender, currentNationalNumber string
	var currentStroke string
	var currentDistance int64

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 2 {
			continue
		}

		switch line[0:2] {
		case "B1":
			parsed.MeetName = strings.TrimSpace(safeSlice(line, hy3B1NameStart, hy3B1NameEnd))
			parsed.Course = hy3ParseCourse(safeSlice(line, hy3B1CourseStart, hy3B1CourseEnd))
			if sd, err := time.Parse("010206", safeSlice(line, hy3B1StartDateStart, hy3B1StartDateEnd)); err == nil {
				parsed.StartDate = sd
			}
			if ed, err := time.Parse("010206", safeSlice(line, hy3B1EndDateStart, hy3B1EndDateEnd)); err == nil {
				parsed.EndDate = ed
			}
			parsed.Location = strings.TrimSpace(safeSlice(line, hy3B1LocationStart, hy3B1LocationEnd))
			parsed.Facility = strings.TrimSpace(safeSlice(line, hy3B1FacilityStart, hy3B1FacilityEnd))

		case "C1":
			currentTeam = strings.TrimSpace(safeSlice(line, hy3C1AcronymStart, hy3C1AcronymEnd))
			currentFirstName, currentLastName = "", ""
			currentBirthDate = time.Time{}
			currentGender, currentNationalNumber = "", ""

		case "D0":
			currentLastName = strings.TrimSpace(safeSlice(line, hy3D0LastNameStart, hy3D0LastNameEnd))
			currentFirstName = strings.TrimSpace(safeSlice(line, hy3D0FirstNameStart, hy3D0FirstNameEnd))
			currentGender = hy3ParseGender(safeChar(line, hy3D0GenderPos))
			if bd, err := time.Parse("010206", safeSlice(line, hy3D0BirthDateStart, hy3D0BirthDateEnd)); err == nil {
				currentBirthDate = bd
			} else {
				currentBirthDate = time.Time{}
			}
			currentNationalNumber = strings.TrimSpace(safeSlice(line, hy3D0NationalNumberStart, hy3D0NationalNumberEnd))
			currentStroke = ""
			currentDistance = 0

		case "E0":
			currentStroke = hy3ParseStroke(safeChar(line, hy3E0StrokePos))
			if d, err := strconv.ParseInt(strings.TrimSpace(safeSlice(line, hy3E0DistanceStart, hy3E0DistanceEnd)), 10, 64); err == nil {
				currentDistance = d
			} else {
				currentDistance = 0
			}

		case "E1":
			if currentTeam == "" || currentFirstName == "" || currentLastName == "" || currentStroke == "" || currentDistance == 0 {
				continue
			}

			timeCode := safeChar(line, hy3E1TimeCodePos)
			if timeCode == 'S' || timeCode == 'N' {
				continue
			}

			parsed.Results = append(parsed.Results, hy3RawResult{
				TeamAcronym:    currentTeam,
				FirstName:      currentFirstName,
				LastName:       currentLastName,
				BirthDate:      currentBirthDate,
				Gender:         currentGender,
				NationalNumber: currentNationalNumber,
				Stroke:         currentStroke,
				Distance:       currentDistance,
				Time:           strings.TrimSpace(safeSlice(line, hy3E1FinalTimeStart, hy3E1FinalTimeEnd)),
				DQ:             timeCode == 'Q',
			})
		}
	}

	return parsed
}

func hy3ParseCourse(code string) string {
	switch strings.TrimSpace(strings.ToUpper(code)) {
	case "SC", "S":
		return swimming.CourseShort
	case "LC", "L":
		return swimming.CourseLong
	default:
		return ""
	}
}

func hy3ParseGender(code byte) string {
	switch code {
	case 'M':
		return swimming.GenderMale
	case 'F':
		return swimming.GenderFemale
	default:
		return ""
	}
}

func hy3ParseStroke(code byte) string {
	switch code {
	case 'A':
		return swimming.StyleFreestyle
	case 'B':
		return swimming.StyleBackstroke
	case 'C':
		return swimming.StyleBreaststroke
	case 'D':
		return swimming.StyleButterfly
	case 'E':
		return swimming.StyleMedley
	default:
		return ""
	}
}

func safeSlice(s string, start, end int) string {
	if start >= len(s) {
		return ""
	}
	if end > len(s) {
		end = len(s)
	}
	return s[start:end]
}

func safeChar(s string, pos int) byte {
	if pos >= len(s) {
		return ' '
	}
	return s[pos]
}
