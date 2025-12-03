package times

import (
	"cmp"
	"fmt"
	"slices"
	"strings"
)

func groupRecordsByDefinition(records []*Record) []Record {
	grouping := make(map[int64]*Record)

	for _, record := range records {
		key := record.Definition.ID
		groupByDefinition(grouping, record, key)
	}

	return squeezeFastOnes(grouping)
}

func groupByDefinition(grouping map[int64]*Record, record *Record, key int64) {
	grouped := grouping[key]

	if grouped == nil {
		grouping[key] = record
		return
	}

	if record.Time > grouped.Time {
		grouped.Previous = append(grouped.Previous, *record)
	}

	if record.Time == grouped.Time {
		if record.Year == nil || grouped.Year == nil || record.Month == nil || grouped.Month == nil {
			if record.ID > grouped.ID {
				record.Previous = append(record.Previous, *grouped)
				grouping[key] = record
			} else {
				grouped.Previous = append(grouped.Previous, *record)
				grouping[key] = grouped
			}
			return
		}

		if *record.Year > *grouped.Year {
			record.Previous = append(record.Previous, *grouped)
			grouping[key] = record
		} else if *record.Year < *grouped.Year {
			grouped.Previous = append(grouped.Previous, *record)
			grouping[key] = grouped
		} else if *record.Month > *grouped.Month {
			record.Previous = append(record.Previous, *grouped)
			grouping[key] = record
		} else if *record.Month < *grouped.Month {
			grouped.Previous = append(grouped.Previous, *record)
			grouping[key] = grouped
		} else {
			record.Previous = append(record.Previous, *grouped)
			grouping[key] = record
		}
	}

	if record.Time < grouped.Time {
		record.Previous = append(record.Previous, *grouped)
		grouping[key] = record
	}
}

func groupPosterRecordsByDefinition(records []*RecordPoster) []*RecordPoster {
	grouping := make(map[string]*RecordPoster)
	for _, record := range records {
		if grouping[record.Placeholder] == nil {
			grouping[record.Placeholder] = record
			continue
		}

		if strings.Contains(record.Placeholder, "_HOLDER") {
			grouping[record.Placeholder].Value = fmt.Sprintf(
				"%s, %s",
				grouping[record.Placeholder].Value,
				record.Value)
		} else {
			newValue := record.Value
			if newValue > grouping[record.Placeholder].Value {
				grouping[record.Placeholder].Value = record.Value
			}
		}
	}

	groupedRecords := make([]*RecordPoster, 0, len(grouping))
	for _, record := range grouping {
		groupedRecords = append(groupedRecords, record)
	}
	return groupedRecords
}

func squeezeFastOnes(grouping map[int64]*Record) []Record {
	fastestRecords := make([]Record, 0, len(grouping))
	for _, record := range grouping {
		fastestRecords = append(fastestRecords, *record)
	}
	sortByStroke(fastestRecords)
	return fastestRecords
}

func sortByStroke(records []Record) {
	slices.SortStableFunc(records, func(a, b Record) int {
		if n := cmp.Compare(a.Definition.Sequence, b.Definition.Sequence); n != 0 {
			return n
		}
		return cmp.Compare(a.Definition.Distance, b.Definition.Distance)
	})
}

func getStandardAgeInterval(age int64, timeStandard TimeStandard) (int64, int64) {
	minAge := age
	maxAge := age
	if timeStandard.Open {
		if timeStandard.MaxAgeTime != nil && age < *timeStandard.MaxAgeTime {
			maxAge = *timeStandard.MaxAgeTime
		} else if age > *timeStandard.MinAgeTime {
			minAge = *timeStandard.MinAgeTime
		}
	}
	return minAge, maxAge
}

func calculateDifferences(standardTimes []*StandardTime) []*StandardTime {
	for i, standardTime := range standardTimes {
		if i < len(standardTimes)-1 {
			standardTimes[i].Difference = standardTime.Standard - standardTimes[i+1].Standard
		}
	}
	return standardTimes
}
