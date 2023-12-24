package swimmers

import "geekswimmers/storage"

func FindTimeStandards(example StandardTime, db storage.Database) ([]*StandardTime, error) {
	stmt := `select ts.name, ts.summary, st.standard 
			 from standard_time st 
	           join time_standard ts on ts.id = st.time_standard
	           join swim_season ss on ss.id = ts.season 
	         where ss.start_date <= now() and ss.end_date >= now()
		       and st.age = $1
			   and st.gender = $2
			   and st.course  = $3
			   and st.stroke = $4
			   and st.distance = $5`

	rows, err := db.Query(stmt, example.Age, example.Gender, example.Course, example.Stroke, example.Distance)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var standardTimes []*StandardTime
	for rows.Next() {
		standardTime := &StandardTime{
			Age:      example.Age,
			Gender:   example.Gender,
			Course:   example.Course,
			Stroke:   example.Stroke,
			Distance: example.Distance,
		}
		err = rows.Scan(&standardTime.TimeStandard.Name, &standardTime.TimeStandard.Summary, &standardTime.Standard)

		if err != nil {
			return nil, err
		}
		standardTimes = append(standardTimes, standardTime)
	}

	return standardTimes, nil
}
