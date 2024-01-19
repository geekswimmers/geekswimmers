package swimmers

import (
	"context"
	"geekswimmers/storage"
)

func FindChampionshipMeets(course string, db storage.Database) ([]*Meet, error) {
	// When we have more than one season in the database, we have to add the swim season in the meet table.
	stmt := `select m.name, m.age_date, m.time_standard, ss.id, ss.name
			 from meet m
			 join swim_season ss on ss.id = m.season
			 where course = $1 
			 	and ss.start_date <= now() and ss.end_date >= now()
			 	and time_standard is not null 
				and age_date is not null
			 order by age_date`
	rows, err := db.Query(context.Background(), stmt, course)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var meets []*Meet
	for rows.Next() {
		meet := &Meet{
			Course: course,
		}
		err = rows.Scan(&meet.Name, &meet.AgeDate, &meet.TimeStandard.ID, &meet.Season.ID, &meet.Season.Name)

		if err != nil {
			return nil, err
		}
		meets = append(meets, meet)
	}

	return meets, nil
}

func FindStandardTimeMeet(example StandardTime, season SwimSeason, db storage.Database) (*StandardTime, error) {
	stmt := `select ts.name, ts.summary, st.standard 
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
	err := row.Scan(&standardTime.TimeStandard.Name, &standardTime.TimeStandard.Summary, &standardTime.Standard)
	if err != nil {
		return nil, err
	}

	return standardTime, nil
}
