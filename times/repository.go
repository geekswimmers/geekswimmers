package times

import (
	"context"
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
		return nil, err
	}
	defer rows.Close()

	var meets []*Meet
	for rows.Next() {
		meet := &Meet{}
		err = rows.Scan(&meet.Name, &meet.AgeDate, &meet.TimeStandard.ID, &meet.Course, &meet.Season.ID, &meet.Season.Name,
			&meet.TimeStandard.MinAgeTime, &meet.TimeStandard.MaxAgeTime, &meet.MinAgeEnforced, &meet.MaxAgeEnforced)
		if err != nil && err.Error() != storage.ErrNoRows {
			return nil, err
		}
		meets = append(meets, meet)
	}

	return meets, nil
}

func findStandardTimeMeet(example StandardTime, season SwimSeason, db storage.Database) (*StandardTime, error) {
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
		return nil, err
	}

	return standardTime, nil
}

func findSwimSeasons(db storage.Database) ([]*SwimSeason, error) {
	stmt := `select ss.id, ss.name, ss.start_date, ss.end_date
	         from swim_season ss
			 order by ss.start_date desc`
	rows, err := db.Query(context.Background(), stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var swimSeasons []*SwimSeason
	for rows.Next() {
		swimSeason := &SwimSeason{}
		err = rows.Scan(&swimSeason.ID, &swimSeason.Name, &swimSeason.StartDate, &swimSeason.EndDate)
		if err != nil && err.Error() != storage.ErrNoRows {
			return nil, err
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
		return nil, err
	}
	defer rows.Close()

	var timeStandards []*TimeStandard
	for rows.Next() {
		timeStandard := &TimeStandard{
			Season: season,
		}
		err = rows.Scan(&timeStandard.ID, &timeStandard.Name, &timeStandard.MinAgeTime, &timeStandard.MaxAgeTime)
		if err != nil && err.Error() != storage.ErrNoRows {
			return nil, err
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
	err := row.Scan(&timeStandard.Season.Name, &timeStandard.Name, &timeStandard.MinAgeTime, &timeStandard.MaxAgeTime)
	if err != nil {
		return nil, err
	}

	return timeStandard, nil
}

func findStandardTimes(example StandardTime, db storage.Database) ([]*StandardTime, error) {
	stmt := `select st.stroke, st.distance, st.standard
			 from standard_time st
			 where st.age = $1 
			   and st.gender = $2 
			   and st.course = $3 
			   and st.time_standard = $4`
	rows, err := db.Query(context.Background(), stmt, example.Age, example.Gender, example.Course, example.TimeStandard.ID)
	if err != nil && err.Error() != storage.ErrNoRows {
		return nil, err
	}
	defer rows.Close()

	var times []*StandardTime
	for rows.Next() {
		time := &StandardTime{}
		err = rows.Scan(&time.Stroke, &time.Distance, &time.Standard)
		if err != nil {
			return nil, err
		}
		times = append(times, time)
	}

	return times, nil
}
