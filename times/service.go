package times

import (
	"fmt"
	"geekswimmers/storage"
	"log"
)

func findCurrentAndPreviousRecords(example StandardTime, db storage.Database) ([]*Record, error) {
	records, err := findRecordsByExample(example, db)
	if err != nil {
		log.Printf("findCurrentAndPreviousRecords.%v", err)
	}

	checkDuplicates := make(map[string]*Record)

	for _, record := range records {
		key := fmt.Sprintf("%s-%s-%s-%s-%s-%s", record.Jurisdiction.Country, record.Jurisdiction.Province, record.Jurisdiction.Region, record.Jurisdiction.City, record.Jurisdiction.Club, record.Jurisdiction.Meet)
		if checkDuplicates[key] == nil {
			checkDuplicates[key] = record
			continue
		}

		if record.Time > checkDuplicates[key].Time {
			faster := checkDuplicates[key]
			faster.Previous = append(faster.Previous, *record)
		}

		if record.Time < checkDuplicates[key].Time {
			record.Previous = append(record.Previous, *checkDuplicates[key])
			checkDuplicates[key] = record
		}
	}

	fastestRecords := make([]*Record, 0, len(checkDuplicates))
	for _, record := range checkDuplicates {
		fastestRecords = append(fastestRecords, record)
	}
	return fastestRecords, nil
}
