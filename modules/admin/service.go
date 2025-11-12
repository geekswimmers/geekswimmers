package admin

import (
	"fmt"
	"geekswimmers/modules/swimming"
	"geekswimmers/modules/times"
	"geekswimmers/storage"
	"geekswimmers/utils"
	"log"
	"strconv"
	"time"
)

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
			Age:      &age,
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
