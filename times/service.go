package times

import (
	"cmp"
	"fmt"
	"slices"
)

func groupRecordsByJurisdiction(records []*Record) []Record {
	grouping := make(map[any]*Record)

	for _, record := range records {
		key := fmt.Sprintf("%s-%s-%s-%s-%s-%s", record.Jurisdiction.Country, record.Jurisdiction.Province, record.Jurisdiction.Region, record.Jurisdiction.City, record.Jurisdiction.Club, record.Jurisdiction.Meet)
		groupDuplicates(grouping, record, key)
	}

	return squizeFastests(grouping)
}

func groupRecordsByDefinition(records []*Record) []Record {
	grouping := make(map[any]*Record)

	for _, record := range records {
		key := record.Definition.ID
		groupDuplicates(grouping, record, key)
	}

	return squizeFastests(grouping)
}

func groupDuplicates(grouping map[any]*Record, record *Record, key any) {
	if grouping[key] == nil {
		grouping[key] = record
		return
	}

	if record.Time > grouping[key].Time {
		faster := grouping[key]
		faster.Previous = append(faster.Previous, *record)
	}

	if record.Time < grouping[key].Time {
		record.Previous = append(record.Previous, *grouping[key])
		grouping[key] = record
	}
}

func squizeFastests(grouping map[any]*Record) []Record {
	fastestRecords := make([]Record, 0, len(grouping))
	for _, record := range grouping {
		fastestRecords = append(fastestRecords, *record)
	}
	sortByStroke(fastestRecords)
	return fastestRecords
}

func sortByStroke(records []Record) {
	slices.SortStableFunc(records, func(a, b Record) int {
		if n := cmp.Compare(a.Definition.Modality, b.Definition.Modality); n != 0 {
			return n
		}
		return cmp.Compare(a.Definition.Distance, b.Definition.Distance)
	})
}
