package admin

import (
	"database/sql"
	"fmt"
	"geekswimmers/modules/swimming"
	"geekswimmers/modules/times"
	"geekswimmers/storage"
	"geekswimmers/utils"
	"log"
	"strconv"
	"time"
)

func findSwimmer(raw hy3RawResult, db storage.Database) (*swimming.Swimmer, error) {
	if raw.NationalNumber != "" {
		swimmer, err := swimming.FindSwimmerByNationalNumber(raw.NationalNumber, db)
		if err != nil {
			return nil, err
		}
		if swimmer != nil {
			if !swimmer.NationalNumber.Valid {
				if err = swimming.UpdateSwimmerNationalNumber(swimmer.ID, raw.NationalNumber, db); err != nil {
					log.Printf("findSwimmer: %v", err)
				}
			}
			return swimmer, nil
		}
	}

	swimmer, err := swimming.FindSwimmerByNameAndBirthDate(raw.FirstName, raw.LastName, raw.BirthDate, db)
	if err != nil {
		return nil, err
	}
	if swimmer != nil && raw.NationalNumber != "" && !swimmer.NationalNumber.Valid {
		if err = swimming.UpdateSwimmerNationalNumber(swimmer.ID, raw.NationalNumber, db); err != nil {
			log.Printf("findSwimmer: %v", err)
		}
	}
	return swimmer, nil
}

func findOrCreateMeet(parsed hy3ParsedFile, db storage.Database) (*times.Meet, error) {
	meet, err := times.FindMeetByNameAndDate(parsed.MeetName, parsed.StartDate, db)
	if err != nil {
		return nil, err
	}
	if meet != nil {
		return meet, nil
	}

	season, err := times.FindSeasonByDate(parsed.StartDate, db)
	if err != nil {
		return nil, err
	}

	meet = &times.Meet{
		Name:      parsed.MeetName,
		Course:    parsed.Course,
		StartDate: parsed.StartDate,
		EndDate:   parsed.EndDate,
		Location:  sql.NullString{String: parsed.Location, Valid: parsed.Location != ""},
		Facility:  sql.NullString{String: parsed.Facility, Valid: parsed.Facility != ""},
	}
	if season != nil {
		meet.Season = *season
	}

	if err = times.InsertMeet(meet, db); err != nil {
		return nil, err
	}
	return meet, nil
}

func processMeetResults(parsed hy3ParsedFile, db storage.Database) (meetID int64, imported, skipped, failed int) {
	meet, err := findOrCreateMeet(parsed, db)
	if err != nil {
		log.Printf("processMeetResults: %v", err)
		return 0, 0, 0, 0
	}
	meetID = meet.ID

	for _, raw := range parsed.Results {
		team, err := swimming.FindTeamByAcronym(raw.TeamAcronym, db)
		if err != nil {
			log.Printf("processMeetResults: %v", err)
			failed++
			continue
		}
		if team == nil {
			skipped++
			continue
		}

		swimmer, err := findSwimmer(raw, db)
		if err != nil {
			log.Printf("processMeetResults: %v", err)
			failed++
			continue
		}
		if swimmer == nil {
			skipped++
			continue
		}

		event, err := swimming.FindSwimEventByStrokeAndDistance(raw.Stroke, raw.Distance, db)
		if err != nil {
			log.Printf("processMeetResults: %v", err)
			failed++
			continue
		}
		if event == nil {
			skipped++
			continue
		}

		meetEvent, err := findMeetEvent(meetID, event.ID, raw.Gender, db)
		if err != nil {
			log.Printf("processMeetResults: %v", err)
			failed++
			continue
		}
		if meetEvent == nil {
			meetEvent = &times.MeetEvent{MeetID: meetID, Event: *event, Gender: raw.Gender}
			if err = insertMeetEvent(meetEvent, db); err != nil {
				log.Printf("processMeetResults: %v", err)
				failed++
				continue
			}
		}

		var resultTime *int64
		if !raw.DQ && raw.Time != "" {
			t, err := utils.MillisecondsFromText(raw.Time)
			if err != nil {
				log.Printf("processMeetResults: invalid time %q: %v", raw.Time, err)
				failed++
				continue
			}
			resultTime = &t
		}

		result := &times.MeetResult{
			MeetEvent:  *meetEvent,
			Swimmer:    *swimmer,
			Team:       *team,
			ResultTime: resultTime,
			DQ:         raw.DQ,
		}
		if err = insertMeetResult(result, db); err != nil {
			log.Printf("processMeetResults: %v", err)
			failed++
			continue
		}
		imported++
	}
	return
}

func processStandardRecords(records [][]string, timeStandard *times.TimeStandard, publicationDate time.Time, db storage.Database) ([]times.StandardTime, []times.StandardTime, []string) {
	headers := records[0]
	dataRows := records[1:]

	var importedRecords []times.StandardTime
	var notImportedRecords []times.StandardTime
	var failedRecords []string

	// Process each data row
	for i, row := range dataRows {
		if len(row) != len(headers) {
			failedRecords = append(failedRecords, fmt.Sprintf("Row %d has %d columns, expected %d: %v", i+2, len(row), len(headers), row))
			continue
		}

		gender := processGender(row[0])
		if gender == "" {
			failedRecords = append(failedRecords, fmt.Sprintf("Row %d has invalid gender %v: %v", i+2, row[0], row))
			continue
		}

		age := processAge(row[1], row[2])
		if age == 0 {
			failedRecords = append(failedRecords, fmt.Sprintf("Row %d has invalid age '%v'/'%v': %v", i+2, row[1], row[2], row))
			continue
		}

		course := processCourse(row[3])
		if course == "" {
			failedRecords = append(failedRecords, fmt.Sprintf("Row %d has invalid course %v: %v", i+2, row[3], row))
			continue
		}

		distance := processDistance(row[4])
		if distance == 0 {
			failedRecords = append(failedRecords, fmt.Sprintf("Row %d has invalid distance %v: %v", i+2, row[4], row))
			continue
		}

		style := processStyle(row[5])
		if style == "" {
			failedRecords = append(failedRecords, fmt.Sprintf("Row %d has invalid style %v: %v", i+2, row[5], row))
			continue
		}

		standard, err := utils.MillisecondsFromText(row[6])
		if err != nil {
			failedRecords = append(failedRecords, fmt.Sprintf("Row %d has invalid time %v: %v", i+2, row[6], err))
			continue
		}

		definition := times.StandardDefinition{
			Age:      sql.NullInt64{Int64: age, Valid: true},
			Gender:   gender,
			Course:   course,
			Style:    style,
			Distance: distance,
		}

		standardTime := times.StandardTime{
			TimeStandard: *timeStandard,
			Definition:   definition,
			Standard:     standard,
			UpdateDate:   publicationDate,
		}

		standardDefinition, err := times.FindStandardDefinition(&definition, db)
		if err != nil {
			log.Printf("Error finding standard definition: %v", err)
			notImportedRecords = append(notImportedRecords, standardTime)
			continue
		}
		standardTime.Definition = *standardDefinition

		existingStandardTime, err := times.GetStandardTimeByDefinition(standardDefinition, timeStandard, db)
		if existingStandardTime == nil {
			err = times.InsertStandardTime(&standardTime, db)
			if err != nil {
				failedRecords = append(failedRecords, fmt.Sprintf("Row %d is invalid: %v", i+2, err))
				continue
			}
			importedRecords = append(importedRecords, standardTime)
			continue
		}

		if existingStandardTime.Standard == standard || existingStandardTime.UpdateDate.Compare(publicationDate) >= 0 {
			notImportedRecords = append(notImportedRecords, *existingStandardTime)
			continue
		}

		err = times.InsertStandardTime(&standardTime, db)
		if err != nil {
			failedRecords = append(failedRecords, fmt.Sprintf("Row %d is invalid: %v", i+2, err))
		}
	}

	return importedRecords, notImportedRecords, failedRecords
}

func processGender(value string) string {
	switch value {
	case "M":
		return swimming.GenderMale
	case "F":
		return swimming.GenderFemale
	default:
		return ""
	}
}

func processAge(valueFrom string, valueTo string) int64 {
	ageFrom, err := strconv.ParseInt(valueFrom, 10, 64)
	if err != nil {
		ageTo, err := strconv.ParseInt(valueTo, 10, 64)
		if err != nil {
			return 0
		}
		return ageTo
	}
	return ageFrom
}

func processCourse(value string) string {
	switch value {
	case "LCM":
		return swimming.CourseLong
	case "SCM":
		return swimming.CourseShort
	default:
		return ""
	}
}

func processDistance(value string) int64 {
	distance, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0
	}
	return distance
}

func processStyle(value string) string {
	switch value {
	case "FREE":
		return swimming.StyleFreestyle
	case "BACK":
		return swimming.StyleBackstroke
	case "BREAST":
		return swimming.StyleBreaststroke
	case "FLY":
		return swimming.StyleButterfly
	case "MEDLEY":
		return swimming.StyleMedley
	default:
		return ""
	}
}
