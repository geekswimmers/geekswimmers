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
			   	and st.stroke = $6
			   	and st.distance = $7`

	row := db.QueryRow(context.Background(), stmt,
		season.ID, example.TimeStandard.ID, example.Age, example.Gender, example.Course, example.Stroke, example.Distance)

	standardTime := &StandardTime{
		Age:      example.Age,
		Gender:   example.Gender,
		Course:   example.Course,
		Stroke:   example.Stroke,
		Distance: example.Distance,
	}
	err := row.Scan(&standardTime.TimeStandard.ID, &standardTime.TimeStandard.Name, &standardTime.Standard)
	if err != nil && err.Error() != storage.ErrNoRows {
		return nil, fmt.Errorf("findStandardTimeMeet: %v", err)
	}

	return standardTime, nil
}

func findRecordsByExample(example StandardTime, db storage.Database) ([]*Record, error) {
	sql := `select r.record_time, r.record_date, coalesce(r.holder, ''), coalesce(j.id, 0), coalesce(j.country, ''), 
	            coalesce(j.province, ''), coalesce(j.region, ''), coalesce(j.city, ''), coalesce(j.club, ''), coalesce(j.meet, '')
			from record r
                join record_definition rd on rd.id = r.definition
                left join jurisdiction j on j.id = r.jurisdiction 
            where ((rd.min_age is null and rd.max_age >= $1) or 
				(rd.min_age <= $1 and rd.max_age is null) or 
				(rd.min_age <= $1 and rd.max_age >= $1)) and
                rd.gender = $2 and
                rd.course = $3 and
                rd.stroke = $4 and
                rd.distance = $5
            order by r.record_time desc`
	rows, err := db.Query(context.Background(), sql, example.Age, example.Gender, example.Course, example.Stroke, example.Distance)
	if err != nil {
		return nil, fmt.Errorf("findRecords: %v", err)
	}
	defer rows.Close()

	var records []*Record
	for rows.Next() {
		record := &Record{
			Definition: RecordDefinition{
				Age: example.Age,
			},
		}
		err = rows.Scan(&record.Time, &record.Date, &record.Holder, &record.Jurisdiction.ID, &record.Jurisdiction.Country,
			&record.Jurisdiction.Province, &record.Jurisdiction.Region, &record.Jurisdiction.City,
			&record.Jurisdiction.Club, &record.Jurisdiction.Meet)
		if err != nil && err.Error() != storage.ErrNoRows {
			return nil, fmt.Errorf("findRecords: %v", err)
		}
		records = append(records, record)
	}

	return records, nil
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
			 where ts.season = $1`
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
	stmt := `select ss.name, ts.name, ts.min_age_time, ts.max_age_time
			 from time_standard ts
			 	join swim_season ss on ss.id = ts.season
	         where ts.id = $1`

	row := db.QueryRow(context.Background(), stmt, id)

	timeStandard := &TimeStandard{
		ID: id,
	}
	if err := row.Scan(&timeStandard.Season.Name, &timeStandard.Name, &timeStandard.MinAgeTime, &timeStandard.MaxAgeTime); err != nil {
		return nil, fmt.Errorf("findTimeStandard: %v", err)
	}

	return timeStandard, nil
}

func findStandardTimes(example StandardTime, db storage.Database) ([]*StandardTime, error) {
	stmt := `select st.stroke, st.distance, st.standard
			 from standard_time st
			 where st.age = $1 
			   and st.gender = $2 
			   and st.course = $3 
			   and st.time_standard = $4
			 order by st.stroke, st.standard asc`
	rows, err := db.Query(context.Background(), stmt, example.Age, example.Gender, example.Course, example.TimeStandard.ID)
	if err != nil && err.Error() != storage.ErrNoRows {
		return nil, fmt.Errorf("findStandardTimes: %v", err)
	}
	defer rows.Close()

	var times []*StandardTime
	for rows.Next() {
		time := &StandardTime{}
		err = rows.Scan(&time.Stroke, &time.Distance, &time.Standard)
		if err != nil {
			return nil, fmt.Errorf("findStandardTimes: %v", err)
		}
		times = append(times, time)
	}

	return times, nil
}

func findStandardsEvent(example StandardTime, db storage.Database) ([]*StandardTime, error) {
	stmt := `select ts.id , ts.name, st.standard 
			 from standard_time st 
				join time_standard ts on ts.id = st.time_standard 
			 where st.age = $1 and st.gender = $2 and st.course = $3 and st.distance = $4 and st.stroke = $5
			 order by st.standard desc`
	rows, err := db.Query(context.Background(), stmt, example.Age, example.Gender, example.Course, example.Distance, example.Stroke)
	if err != nil && err.Error() != storage.ErrNoRows {
		return nil, fmt.Errorf("findStandardsEvent: %v", err)
	}
	defer rows.Close()

	var times []*StandardTime
	for rows.Next() {
		time := &StandardTime{}
		err = rows.Scan(&time.TimeStandard.ID, &time.TimeStandard.Name, &time.Standard)
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
