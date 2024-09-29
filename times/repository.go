package times

import (
	"context"
	"fmt"
	"geekswimmers/storage"
)

func findChampionshipMeets(db storage.Database) ([]*Meet, error) {
	stmt := `select m.name, m.age_date, m.time_standard, m.course, ss.id, ss.name, 
	                ts.min_age_time, ts.max_age_time, m.min_age_enforced, m.max_age_enforced
			 from meet m
			    join swim_season ss on ss.id = m.season
			    join time_standard ts on ts.id = m.time_standard
			 where ss.start_date <= now() and ss.end_date >= now()
			 	and m.end_date >= now()
			 	and m.time_standard is not null 
				and m.age_date is not null
			 order by m.age_date`
	rows, err := db.Query(context.Background(), stmt)
	if err != nil {
		return nil, fmt.Errorf("findChampionshipMeets: %v", err)
	}
	defer rows.Close()

	var meets []*Meet
	for rows.Next() {
		meet := &Meet{}
		err = rows.Scan(&meet.Name, &meet.AgeDate, &meet.TimeStandard.ID, &meet.Course, &meet.Season.ID, &meet.Season.Name,
			&meet.TimeStandard.MinAgeTime, &meet.TimeStandard.MaxAgeTime, &meet.MinAgeEnforced, &meet.MaxAgeEnforced)
		if err != nil && err.Error() != storage.ErrNoRows {
			return nil, fmt.Errorf("findChampionshipMeets: %v", err)
		}
		meets = append(meets, meet)
	}

	return meets, nil
}

func findStandardChampionshipMeets(timeStandard TimeStandard, db storage.Database) ([]*Meet, error) {
	stmt := `select m.id, m.name, m.course
			 from meet m
			 where m.time_standard = $1
			 order by m.name`
	rows, err := db.Query(context.Background(), stmt, timeStandard.ID)
	if err != nil {
		return nil, fmt.Errorf("findStandardChampioshipMeets: %v", err)
	}
	defer rows.Close()

	var meets []*Meet
	for rows.Next() {
		meet := &Meet{}
		err = rows.Scan(&meet.ID, &meet.Name, &meet.Course)
		if err != nil && err.Error() != storage.ErrNoRows {
			return nil, fmt.Errorf("findStandardChampioshipMeets: %v", err)
		}
		meets = append(meets, meet)
	}

	return meets, nil
}

func findStandardTimeMeetByExample(example StandardTime, season SwimSeason, db storage.Database) (*StandardTime, error) {
	stmt := `select ts.id, ts.name, st.standard 
			 from standard_time st 
	           	join time_standard ts on ts.id = st.time_standard
	           	join swim_season ss on ss.id = ts.season 
	         where ss.id = $1
			   	and st.time_standard = $2
		       	and st.age = $3
			   	and st.gender = $4
			   	and st.course  = $5
			   	and st.style = $6
			   	and st.distance = $7`

	row := db.QueryRow(context.Background(), stmt,
		season.ID, example.TimeStandard.ID, example.Age, example.Gender, example.Course, example.Style, example.Distance)

	standardTime := &StandardTime{
		Age:      example.Age,
		Gender:   example.Gender,
		Course:   example.Course,
		Style:    example.Style,
		Distance: example.Distance,
	}
	err := row.Scan(&standardTime.TimeStandard.ID, &standardTime.TimeStandard.Name, &standardTime.Standard)
	if err != nil && err.Error() != storage.ErrNoRows {
		return nil, fmt.Errorf("findStandardTimeMeet: %v", err)
	}

	return standardTime, nil
}

func findRecordsByExample(example RecordDefinition, db storage.Database) ([]*Record, error) {
	sql := `select r.record_time, r.year, r.month, coalesce(r.holder, ''), coalesce(j.id, 0), coalesce(j.country, ''), 
	            coalesce(j.province, ''), coalesce(j.region, ''), coalesce(j.city, ''), coalesce(j.club, ''), coalesce(j.meet, ''),
				rd.min_age, rd.max_age
			from record r
                join record_definition rd on rd.id = r.definition
                left join jurisdiction j on j.id = r.jurisdiction 
            where ((rd.min_age is null and rd.max_age >= $1) or 
				(rd.min_age <= $1 and rd.max_age is null) or 
				(rd.min_age <= $1 and rd.max_age >= $1)) and
                rd.gender = $2 and
                rd.course = $3 and
                rd.style = $4 and
                rd.distance = $5
            order by r.record_time desc`
	rows, err := db.Query(context.Background(), sql, example.Age, example.Gender, example.Course, example.Style, example.Distance)
	if err != nil {
		return nil, fmt.Errorf("findRecordsByExample: %v", err)
	}
	defer rows.Close()

	var records []*Record
	for rows.Next() {
		record := &Record{
			Definition: RecordDefinition{
				Age: example.Age,
			},
		}
		err = rows.Scan(&record.Time, &record.Year, &record.Month, &record.Holder, &record.Jurisdiction.ID, &record.Jurisdiction.Country,
			&record.Jurisdiction.Province, &record.Jurisdiction.Region, &record.Jurisdiction.City,
			&record.Jurisdiction.Club, &record.Jurisdiction.Meet,
			&record.Definition.MinAge, &record.Definition.MaxAge)
		if err != nil && err.Error() != storage.ErrNoRows {
			return nil, fmt.Errorf("findRecordsByExample: %v", err)
		}
		records = append(records, record)
	}

	return records, nil
}

func findRecordsByJurisdiction(jurisdiction Jurisdiction, example RecordDefinition, db storage.Database) ([]*Record, error) {
	sql := `select r.record_time, r.year, r.month, coalesce(r.holder, ''),
				rd.id, coalesce(rd.min_age, 0), coalesce(rd.max_age, 0), rd.style, rd.distance
			from record r
                join record_definition rd on rd.id = r.definition
                left join jurisdiction j on j.id = r.jurisdiction 
            where j.id = $1 and
				((rd.min_age is null and rd.max_age >= $2) or 
				(rd.min_age <= $2 and rd.max_age is null) or 
				(rd.min_age <= $2 and rd.max_age >= $2)) and
                rd.gender = $3 and
                rd.course = $4
            order by rd.style, r.record_time asc`
	rows, err := db.Query(context.Background(), sql, jurisdiction.ID, example.Age, example.Gender, example.Course)
	if err != nil {
		return nil, fmt.Errorf("findRecordsByJurisdiction: %v", err)
	}
	defer rows.Close()

	var records []*Record
	for rows.Next() {
		record := &Record{
			Definition: RecordDefinition{
				Age:    example.Age,
				Gender: example.Gender,
				Course: example.Course,
			},
			Jurisdiction: jurisdiction,
		}
		err = rows.Scan(&record.Time, &record.Year, &record.Month, &record.Holder, &record.Definition.ID, &record.Definition.MinAge, &record.Definition.MaxAge,
			&record.Definition.Style, &record.Definition.Distance)
		if err != nil && err.Error() != storage.ErrNoRows {
			return nil, fmt.Errorf("findRecordsByJurisdiction: %v", err)
		}
		records = append(records, record)
	}

	return records, nil
}

func findRecordsAgeRanges(jurisdiction Jurisdiction, db storage.Database) ([]*RecordDefinition, error) {
	sql := `select distinct rd.min_age, rd.max_age 
			from record_definition rd 
				join record r on r.definition = rd.id 
			where r.jurisdiction = $1
			order by rd.max_age, rd.min_age`

	rows, err := db.Query(context.Background(), sql, jurisdiction.ID)
	if err != nil {
		return nil, fmt.Errorf("findRecordsAgeRanges: %v", err)
	}
	defer rows.Close()

	var definitions []*RecordDefinition
	for rows.Next() {
		definition := &RecordDefinition{}

		err = rows.Scan(&definition.MinAge, &definition.MaxAge)
		if err != nil && err.Error() != storage.ErrNoRows {
			return nil, fmt.Errorf("findRecordsAgeRanges: %v", err)
		}
		definitions = append(definitions, definition)
	}

	return definitions, nil
}

func findSwimSeasons(db storage.Database) ([]*SwimSeason, error) {
	stmt := `select ss.id, ss.name, ss.start_date, ss.end_date
	         from swim_season ss
			 order by ss.start_date desc`
	rows, err := db.Query(context.Background(), stmt)
	if err != nil {
		return nil, fmt.Errorf("findSwimSeasons: %v", err)
	}
	defer rows.Close()

	var swimSeasons []*SwimSeason
	for rows.Next() {
		swimSeason := &SwimSeason{}
		err = rows.Scan(&swimSeason.ID, &swimSeason.Name, &swimSeason.StartDate, &swimSeason.EndDate)
		if err != nil && err.Error() != storage.ErrNoRows {
			return nil, fmt.Errorf("findSwimSeasons: %v", err)
		}
		swimSeasons = append(swimSeasons, swimSeason)
	}

	return swimSeasons, nil
}

func findTimeStandards(season SwimSeason, db storage.Database) ([]*TimeStandard, error) {
	stmt := `select ts.id, ts.name, ts.min_age_time, ts.max_age_time
	         from time_standard ts
			 where ts.season = $1
			 order by ts.name`
	rows, err := db.Query(context.Background(), stmt, season.ID)
	if err != nil {
		return nil, fmt.Errorf("findTimeStandards: %v", err)
	}
	defer rows.Close()

	var timeStandards []*TimeStandard
	for rows.Next() {
		timeStandard := &TimeStandard{
			Season: season,
		}
		err = rows.Scan(&timeStandard.ID, &timeStandard.Name, &timeStandard.MinAgeTime, &timeStandard.MaxAgeTime)
		if err != nil && err.Error() != storage.ErrNoRows {
			return nil, fmt.Errorf("findTimeStandards: %v", err)
		}
		timeStandards = append(timeStandards, timeStandard)
	}

	return timeStandards, nil
}

func findTimeStandard(id int64, db storage.Database) (*TimeStandard, error) {
	stmt := `select ss.name, ts.name, ts.min_age_time, ts.max_age_time, ts.open, coalesce(ts.source_title, 'None'), coalesce(ts.source_link, '#')
			 from time_standard ts
			 	join swim_season ss on ss.id = ts.season
	         where ts.id = $1`

	row := db.QueryRow(context.Background(), stmt, id)

	timeStandard := &TimeStandard{
		ID: id,
	}
	if err := row.Scan(&timeStandard.Season.Name, &timeStandard.Name,
		&timeStandard.MinAgeTime, &timeStandard.MaxAgeTime, &timeStandard.Open,
		&timeStandard.Source.Title, &timeStandard.Source.Link); err != nil {
		return nil, fmt.Errorf("findTimeStandard: %v", err)
	}

	return timeStandard, nil
}

func findJurisdictions(db storage.Database) ([]*Jurisdiction, error) {
	stmt := `select j.id, coalesce(j.country, ''), coalesce(j.province, ''), coalesce(j.region, ''), coalesce(j.city, ''), coalesce(j.club, ''), coalesce(j.meet, '')
	         from jurisdiction j
			 order by country, province, region, city, club, meet`
	rows, err := db.Query(context.Background(), stmt)
	if err != nil {
		return nil, fmt.Errorf("findJurisdictions: %v", err)
	}
	defer rows.Close()

	var jurisdictions []*Jurisdiction
	for rows.Next() {
		jurisdiction := &Jurisdiction{}
		err = rows.Scan(&jurisdiction.ID, &jurisdiction.Country, &jurisdiction.Province, &jurisdiction.Region, &jurisdiction.City, &jurisdiction.Club, &jurisdiction.Meet)
		if err != nil && err.Error() != storage.ErrNoRows {
			return nil, fmt.Errorf("findJurisdictions: %v", err)
		}
		jurisdiction.SetTitle(0)
		jurisdiction.SetSubTitle()
		jurisdictions = append(jurisdictions, jurisdiction)
	}

	return jurisdictions, nil
}

func findJurisdiction(id int64, db storage.Database) (*Jurisdiction, error) {
	sql := `select j.id, coalesce(j.country, ''), coalesce(j.province, ''), coalesce(j.region, ''), coalesce(j.city, ''), coalesce(j.club, ''), coalesce(j.meet, '')
			from jurisdiction j
			where j.id = $1`

	row := db.QueryRow(context.Background(), sql, id)

	jurisdiction := &Jurisdiction{
		ID: id,
	}
	if err := row.Scan(&jurisdiction.ID, &jurisdiction.Country, &jurisdiction.Province, &jurisdiction.Region, &jurisdiction.City, &jurisdiction.Club, &jurisdiction.Meet); err != nil {
		return nil, fmt.Errorf("findJurisdiction: %v", err)
	}

	return jurisdiction, nil
}

func findOpenStandardTimes(example StandardTime, db storage.Database) ([]*StandardTime, error) {
	stmt := `select st.style, st.distance, st.standard
			 from standard_time st
			 where st.age between $1 and $2
			   and st.gender = $3 
			   and st.course = $4 
			   and st.time_standard = $5
			 order by st.style, st.standard asc`
	rows, err := db.Query(context.Background(), stmt,
		example.TimeStandard.MinAgeTime, example.TimeStandard.MaxAgeTime, example.Gender, example.Course,
		example.TimeStandard.ID)
	if err != nil && err.Error() != storage.ErrNoRows {
		return nil, fmt.Errorf("findStandardTimes: %v", err)
	}
	defer rows.Close()

	var times []*StandardTime
	for rows.Next() {
		time := &StandardTime{}
		err = rows.Scan(&time.Style, &time.Distance, &time.Standard)
		if err != nil {
			return nil, fmt.Errorf("findStandardTimes: %v", err)
		}
		times = append(times, time)
	}

	return times, nil
}

func findAgeGroupStandardTimes(example StandardTime, db storage.Database) ([]*StandardTime, error) {
	stmt := `select st.style, st.distance, st.standard
			 from standard_time st
			 where st.age = $1 
			   and st.gender = $2 
			   and st.course = $3 
			   and st.time_standard = $4
			 order by st.style, st.standard asc`
	rows, err := db.Query(context.Background(), stmt, example.Age, example.Gender, example.Course, example.TimeStandard.ID)
	if err != nil && err.Error() != storage.ErrNoRows {
		return nil, fmt.Errorf("findStandardTimes: %v", err)
	}
	defer rows.Close()

	var times []*StandardTime
	for rows.Next() {
		time := &StandardTime{}
		err = rows.Scan(&time.Style, &time.Distance, &time.Standard)
		if err != nil {
			return nil, fmt.Errorf("findStandardTimes: %v", err)
		}
		times = append(times, time)
	}

	return times, nil
}

func findStandardTimes(example StandardTime, db storage.Database) ([]*StandardTime, error) {
	if example.TimeStandard.Open {
		return findOpenStandardTimes(example, db)
	}

	return findAgeGroupStandardTimes(example, db)
}

func findStandardsEvent(example StandardTime, db storage.Database) ([]*StandardTime, error) {
	stmt := `select ts.id , ts.name, st.standard, ss.id, ss.name 
			 from standard_time st 
				join time_standard ts on ts.id = st.time_standard 
				join swim_season ss on ts.season = ss.id
			 where st.age = $1 and st.gender = $2 and st.course = $3 and st.distance = $4 and st.style = $5
			 order by ss.name desc, st.standard desc`

	rows, err := db.Query(context.Background(), stmt, example.Age, example.Gender, example.Course, example.Distance, example.Style)
	if err != nil && err.Error() != storage.ErrNoRows {
		return nil, fmt.Errorf("findStandardsEvent: %v", err)
	}
	defer rows.Close()

	var times []*StandardTime
	for rows.Next() {
		time := &StandardTime{}
		err = rows.Scan(
			&time.TimeStandard.ID, &time.TimeStandard.Name, &time.Standard,
			&time.TimeStandard.Season.ID, &time.TimeStandard.Season.Name,
		)
		if err != nil {
			return nil, fmt.Errorf("findStandardsEvent: %v", err)
		}
		times = append(times, time)
	}

	return times, nil
}

func findMinAndMaxStandardAges(db storage.Database) (int64, int64, error) {
	stmt := `select min(age) as min_age, max(age) as max_age from standard_time`

	row := db.QueryRow(context.Background(), stmt)

	var minAge, maxAge int64
	if err := row.Scan(&minAge, &maxAge); err != nil {
		return 0, 0, fmt.Errorf("findMinAndMaxAges: %v", err)
	}

	return minAge, maxAge, nil
}
