package times

import (
	"context"
	"fmt"
	"geekswimmers/modules/swimming"
	"geekswimmers/storage"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

func findSwimSeasons(db storage.Database) ([]*SwimSeason, error) {
	stm := `select ss.id, ss.name, ss.start_date, ss.end_date
	         from swim_season ss
			 order by ss.start_date desc`
	rows, err := db.Query(context.Background(), stm)
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

func GetCurrentSwimSeason(db storage.Database) (*SwimSeason, error) {
	stm := `select id, name 
			from swim_season ss
			where start_date <= $1 and end_date >= $1`

	row := db.QueryRow(context.Background(), stm, time.Now())

	swimSeason := &SwimSeason{}
	if err := row.Scan(&swimSeason.ID, &swimSeason.Name); err != nil {
		return nil, fmt.Errorf("GetCurrentSwimSeason: %v", err)
	}

	return swimSeason, nil
}

func getRecordDefinition(id int64, db storage.Database) (*RecordDefinition, error) {
	stm := `select rd.gender, rd.course, rd.style, rd.distance, rd.min_age, rd.max_age
			from record_definition rd
			where rd.id = $1`

	row := db.QueryRow(context.Background(), stm, id)

	recordDefinition := &RecordDefinition{
		ID: id,
	}
	err := row.Scan(&recordDefinition.Gender, &recordDefinition.Course, &recordDefinition.Style, &recordDefinition.Distance,
		&recordDefinition.MinAge, &recordDefinition.MaxAge)
	if err != nil {
		return nil, fmt.Errorf("getRecordDefinition: %v", err)
	}

	return recordDefinition, nil
}

func findRecordsByDefinition(definition RecordDefinition, db storage.Database) ([]*Record, error) {
	stm := `select r.record_time, r.year, r.month, coalesce(r.holder, ''),
				rs.id, coalesce(rs.source_title, ''), coalesce(rs.source_link, ''),
				coalesce(j.id, 0), coalesce(j.country, ''), j.province, j.region, j.city, j.team, j.meet
			from record r
                join record_set rs on rs.id = r.record_set
                left join jurisdiction j on j.id = rs.jurisdiction
            where r.definition = $1
            order by r.record_time asc`
	rows, err := db.Query(context.Background(), stm, definition.ID)
	if err != nil {
		return nil, fmt.Errorf("findRecordsByDefinition: %v", err)
	}
	defer rows.Close()

	var records []*Record
	for rows.Next() {
		record := &Record{}
		err = rows.Scan(&record.Time, &record.Year, &record.Month, &record.Holder, &record.RecordSet.ID, &record.RecordSet.Source.Title, &record.RecordSet.Source.Link,
			&record.RecordSet.Jurisdiction.ID, &record.RecordSet.Jurisdiction.Country, &record.RecordSet.Jurisdiction.Province, &record.RecordSet.Jurisdiction.Region,
			&record.RecordSet.Jurisdiction.City, &record.RecordSet.Jurisdiction.Team, &record.RecordSet.Jurisdiction.Meet)
		if err != nil && err.Error() != storage.ErrNoRows {
			return nil, fmt.Errorf("findRecordsByDefinition: %v", err)
		}

		records = append(records, record)
	}

	return records, nil
}

func FindRecordsByExample(example RecordDefinition, db storage.Database) ([]*Record, error) {
	stm := `select r.record_time, r.year, r.month, r.holder,
	            j.id, j.world, j.country, j.province, j.region, j.city, j.team, j.meet,
				rd.min_age, rd.max_age
			from record r
                join record_definition rd on rd.id = r.definition
				join record_set rs on rs.id = r.record_set
                left join jurisdiction j on j.id = rs.jurisdiction
				join (
					select definition, min(record_time) as min_record_time
					from record
					group by definition
				) as min_records on r.definition = min_records.definition and r.record_time = min_records.min_record_time
            where ((rd.min_age is null and rd.max_age is null) or
			       (rd.min_age is null and rd.max_age >= $1) or
				   (rd.min_age <= $1 and rd.max_age is null) or
				   (rd.min_age <= $1 and rd.max_age >= $1)) and
                rd.gender = $2 and
                rd.course = $3 and
                rd.style = $4 and
                rd.distance = $5
            order by r.record_time desc`
	rows, err := db.Query(context.Background(), stm, example.Age, example.Gender, example.Course, example.Style, example.Distance)
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
		err = rows.Scan(&record.Time, &record.Year, &record.Month, &record.Holder, &record.RecordSet.Jurisdiction.ID,
			&record.RecordSet.Jurisdiction.World, &record.RecordSet.Jurisdiction.Country, &record.RecordSet.Jurisdiction.Province,
			&record.RecordSet.Jurisdiction.Region, &record.RecordSet.Jurisdiction.City, &record.RecordSet.Jurisdiction.Team,
			&record.RecordSet.Jurisdiction.Meet, &record.Definition.MinAge, &record.Definition.MaxAge)
		if err != nil && err.Error() != storage.ErrNoRows {
			return nil, fmt.Errorf("findRecordsByExample: %v", err)
		}

		records = append(records, record)
	}

	return records, nil
}

func FindRecordsByRecordSet(recordSet RecordSet, example RecordDefinition, db storage.Database) ([]*Record, error) {
	var rows pgx.Rows
	var err error
	if example.Age > 0 {
		stm := `select r.record_time, r.year, r.month, coalesce(r.holder, ''), coalesce(rs.source_title, 'None'), coalesce(rs.source_link, '#'),
					rd.id, coalesce(rd.min_age, 0), coalesce(rd.max_age, 0), rd.style, rd.distance, ss.sequence
				from record r
					join record_definition rd on rd.id = r.definition
					join record_set rs on rs.id = r.record_set
					join swim_style ss on ss.stroke = rd.style
				where rs.id = $1 and
					((rd.min_age is null and rd.max_age >= $2) or
					(rd.min_age <= $2 and rd.max_age is null) or
					(rd.min_age <= $2 and rd.max_age >= $2)) and
					rd.gender = $3 and
					rd.course = $4
				order by ss.sequence asc, rd.distance asc`
		rows, err = db.Query(context.Background(), stm, recordSet.ID, example.Age, example.Gender, example.Course)
		if err != nil {
			return nil, fmt.Errorf("findRecordsByRecordSet: %v", err)
		}
	} else {
		stm := `select r.record_time, r.year, r.month, coalesce(r.holder, ''), coalesce(rs.source_title, 'None'), coalesce(rs.source_link, '#'),
					rd.id, coalesce(rd.min_age, 0), coalesce(rd.max_age, 0), rd.style, rd.distance, ss.sequence
				from record r
					join record_definition rd on rd.id = r.definition
					join record_set rs on rs.id = r.record_set
					join swim_style ss on ss.stroke = rd.style
				where rs.id = $1 and
					rd.gender = $2 and
					rd.course = $3
				order by ss.sequence asc, rd.distance asc`
		rows, err = db.Query(context.Background(), stm, recordSet.ID, example.Gender, example.Course)
		if err != nil {
			return nil, fmt.Errorf("findRecordsByRecordSet: %v", err)
		}
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
			RecordSet: recordSet,
		}
		err = rows.Scan(&record.Time, &record.Year, &record.Month, &record.Holder, &record.RecordSet.Source.Title, &record.RecordSet.Source.Link,
			&record.Definition.ID, &record.Definition.MinAge, &record.Definition.MaxAge,
			&record.Definition.Style, &record.Definition.Distance, &record.Definition.Sequence)
		if err != nil && err.Error() != storage.ErrNoRows {
			return nil, fmt.Errorf("findRecordsByRecordSet: %v", err)
		}
		records = append(records, record)
	}

	return records, nil
}

func findRecordsPoster(recordSet RecordSet, db storage.Database) ([]*RecordPoster, error) {
	stm := `select distinct r.placeholder, r.field, r.value, rm.coord_x , rm.coord_y 
			from report_mapping rm
				join (select 
						r.id, 
						concat(rd.gender, 
								'_', 
								rd.course, 
								case when rd.min_age is null then '' else concat('_', rd.min_age) end, 
								case when rd.max_age is null then '' else concat('_', rd.max_age) end, 
								'_', 
								rd.style,
								'_', 
								rd.distance,
								'_',
								prt.prtype) as placeholder, 
						prt.prtype as field,
						case when prt.prtype = 'HOLDER' then coalesce(r.holder::varchar(50), '')
								when prt.prtype = 'TIME' then coalesce(r.record_time::varchar(50), '')
								when prt.prtype = 'YEAR' then coalesce(r.year::varchar(50), '')
						end as value
					from record r
						join record_definition rd on rd.id = r.definition
						join (select definition, min(record_time) record_time from record group by definition) as df on df.record_time = r.record_time and df.definition = r.definition,
						(select 'HOLDER' as prtype union select 'TIME' union select 'YEAR') prt 
					where r.record_set = $1
						and rd.distance > 25) r on r.placeholder = rm.placeholder`
	rows, err := db.Query(context.Background(), stm, recordSet.ID)
	if err != nil {
		return nil, fmt.Errorf("findRecordsPoster: %v", err)
	}
	defer rows.Close()

	var records []*RecordPoster
	for rows.Next() {
		record := &RecordPoster{}
		err = rows.Scan(&record.Placeholder, &record.Field, &record.Value, &record.CoordX, &record.CoordY)
		if err != nil && err.Error() != storage.ErrNoRows {
			return nil, fmt.Errorf("findRecordsPoster: %v", err)
		}
		records = append(records, record)
	}

	return records, nil
}

func findRecordsAgeRanges(recordSet RecordSet, db storage.Database) ([]*RecordDefinition, error) {
	stm := `select distinct rd.min_age, rd.max_age
			from record_definition rd
				join record r on r.definition = rd.id
			where r.record_set = $1
			order by rd.max_age, rd.min_age`

	rows, err := db.Query(context.Background(), stm, recordSet.ID)
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

func findRecordSets(db storage.Database) ([]*RecordSet, error) {
	stm := `select rs.id, rs.jurisdiction,
	               j.world, j.country, j.province, j.region, j.city, j.team, j.meet
			from record_set rs
			    left join jurisdiction j on j.id = rs.jurisdiction
			
			order by j.world, j.country, j.province, j.region, j.city, j.team, j.meet`
	rows, err := db.Query(context.Background(), stm)
	if err != nil {
		return nil, fmt.Errorf("findRecordSets: %v", err)
	}
	defer rows.Close()

	var recordSets []*RecordSet
	for rows.Next() {
		recordSet := &RecordSet{}
		err = rows.Scan(&recordSet.ID, &recordSet.Jurisdiction.ID, &recordSet.Jurisdiction.World,
			&recordSet.Jurisdiction.Country, &recordSet.Jurisdiction.Province, &recordSet.Jurisdiction.Region,
			&recordSet.Jurisdiction.City, &recordSet.Jurisdiction.Team, &recordSet.Jurisdiction.Meet)
		if err != nil && err.Error() != storage.ErrNoRows {
			return nil, fmt.Errorf("findRecordSets: %v", err)
		}

		recordSets = append(recordSets, recordSet)
	}

	return recordSets, nil
}

func findRecordSet(id int64, db storage.Database) (*RecordSet, error) {
	stm := `select rs.id, rs.jurisdiction, rs.source_title, rs.source_link, 
	               j.world, j.country, j.province, j.region, j.city, j.team, j.meet
			from record_set rs
			    left join jurisdiction j on j.id = rs.jurisdiction
			where rs.id = $1`
	row := db.QueryRow(context.Background(), stm, id)

	recordSet := &RecordSet{
		ID: id,
	}
	if err := row.Scan(&recordSet.ID, &recordSet.Jurisdiction.ID, &recordSet.Source.Title, &recordSet.Source.Link,
		&recordSet.Jurisdiction.World, &recordSet.Jurisdiction.Country, &recordSet.Jurisdiction.Province,
		&recordSet.Jurisdiction.Region, &recordSet.Jurisdiction.City, &recordSet.Jurisdiction.Team,
		&recordSet.Jurisdiction.Meet); err != nil {
		return nil, fmt.Errorf("findRecordSet: %v", err)
	}

	return recordSet, nil
}

func FindTimeStandards(season SwimSeason, jurisdiction swimming.Jurisdiction, db storage.Database) ([]*TimeStandard, error) {
	stm := `select ts.id, ts.name, ts.min_age_time, ts.max_age_time, ts.benchmark
	        from time_standard ts
				join jurisdiction j on j.id = ts.jurisdiction
			where (j.id = $1 or j.region is null)
				and ts.season = $2
			order by ts.name`
	rows, err := db.Query(context.Background(), stm, jurisdiction.ID.Int64, season.ID)
	if err != nil {
		return nil, fmt.Errorf("findTimeStandards: %v", err)
	}
	defer rows.Close()

	var timeStandards []*TimeStandard
	for rows.Next() {
		timeStandard := &TimeStandard{
			Season: season,
		}
		err = rows.Scan(&timeStandard.ID, &timeStandard.Name, &timeStandard.MinAgeTime, &timeStandard.MaxAgeTime, &timeStandard.Benchmark)
		if err != nil && err.Error() != storage.ErrNoRows {
			return nil, fmt.Errorf("findTimeStandards: %v", err)
		}
		timeStandards = append(timeStandards, timeStandard)
	}

	return timeStandards, nil
}

func FindTimeStandard(id int64, db storage.Database) (*TimeStandard, error) {
	stm := `select ss.name, ts.name, ts.min_age_time, ts.max_age_time, ts.open, coalesce(ts.source_title, 'None'), coalesce(ts.source_link, '#'), coalesce(ts.previous, 0)
			from time_standard ts
			 	join swim_season ss on ss.id = ts.season
	        where ts.id = $1`

	row := db.QueryRow(context.Background(), stm, id)

	timeStandard := &TimeStandard{
		ID:       id,
		Previous: &TimeStandard{},
	}
	if err := row.Scan(&timeStandard.Season.Name, &timeStandard.Name,
		&timeStandard.MinAgeTime, &timeStandard.MaxAgeTime, &timeStandard.Open,
		&timeStandard.Source.Title, &timeStandard.Source.Link, &timeStandard.Previous.ID); err != nil {
		return nil, fmt.Errorf("findTimeStandard: %v", err)
	}

	return timeStandard, nil
}

func findLatestTimeStandard(previousId int64, db storage.Database) (*TimeStandard, error) {
	stm := `select ts.id, ts.name
			from time_standard ts
	        where ts.previous = $1`

	row := db.QueryRow(context.Background(), stm, previousId)

	timeStandard := &TimeStandard{}
	if err := row.Scan(&timeStandard.ID, &timeStandard.Name); err != nil {
		return nil, fmt.Errorf("findLatestTimeStandard: %v", err)
	}

	return timeStandard, nil
}

func findStandardTimes(example StandardTime, db storage.Database) ([]*StandardTime, error) {
	var rows pgx.Rows
	var err error

	if example.TimeStandard.MinAgeTime != nil && example.TimeStandard.MaxAgeTime != nil {
		// Age groups
		stm := `select st.style, st.distance, st.standard
			 	from standard_time st
			     	join swim_style ss on ss.stroke = st.style
			 	where st.age between $1 and $2
			   		and st.gender = $3
			   		and st.course = $4
			   		and st.time_standard = $5
				order by ss.sequence, st.standard asc`

		minAge, maxAge := getStandardAgeInterval(example.Age, example.TimeStandard)

		rows, err = db.Query(context.Background(), stm, minAge, maxAge, example.Gender, example.Course, example.TimeStandard.ID)
		if err != nil && err.Error() != storage.ErrNoRows {
			return nil, fmt.Errorf("findStandardTimes: %v", err)
		}
		defer rows.Close()
	} else {
		// Open
		stm := `select st.style, st.distance, st.standard
				from standard_time st
				where st.gender = $1
		  			and st.course = $2
		  			and st.time_standard = $3
				order by st.style, st.standard asc`

		rows, err = db.Query(context.Background(), stm, example.Gender, example.Course, example.TimeStandard.ID)
		if err != nil && err.Error() != storage.ErrNoRows {
			return nil, fmt.Errorf("findStandardTimes: %v", err)
		}
		defer rows.Close()
	}

	var times []*StandardTime
	if rows != nil {
		for rows.Next() {
			time := &StandardTime{}
			err = rows.Scan(&time.Style, &time.Distance, &time.Standard)
			if err != nil {
				return nil, fmt.Errorf("findStandardTimes: %v", err)
			}
			times = append(times, time)
		}
	}

	return times, nil
}

func FindStandardTimeMeetByExample(example StandardTime, season SwimSeason, db storage.Database) (*StandardTime, error) {
	var row pgx.Row

	if example.TimeStandard.MinAgeTime != nil && example.TimeStandard.MaxAgeTime != nil {
		stm := `select ts.id, ts.name, st.standard
				from standard_time st
					join time_standard ts on ts.id = st.time_standard
					join swim_season ss on ss.id = ts.season
				where ss.id = $1
					and st.time_standard = $2
					and st.age between $3 and $4
					and st.gender = $5
					and st.course  = $6
					and st.style = $7
					and st.distance = $8`

		minAge, maxAge := getStandardAgeInterval(example.Age, example.TimeStandard)

		row = db.QueryRow(context.Background(), stm,
			season.ID, example.TimeStandard.ID, minAge, maxAge, example.Gender, example.Course, example.Style, example.Distance)
	} else {
		stm := `select ts.id, ts.name, st.standard
				from standard_time st
					join time_standard ts on ts.id = st.time_standard
					join swim_season ss on ss.id = ts.season
				where ss.id = $1
					and st.time_standard = $2
					and st.gender = $3
					and st.course  = $4
					and st.style = $5
					and st.distance = $6`

		row = db.QueryRow(context.Background(), stm,
			season.ID, example.TimeStandard.ID, example.Gender, example.Course, example.Style, example.Distance)
	}

	standardTime := &StandardTime{
		Age:      example.Age,
		Gender:   example.Gender,
		Course:   example.Course,
		Style:    example.Style,
		Distance: example.Distance,
	}
	err := row.Scan(&standardTime.TimeStandard.ID, &standardTime.TimeStandard.Name, &standardTime.Standard)
	if err != nil && err.Error() != storage.ErrNoRows {
		return nil, fmt.Errorf("findStandardTimeMeetByExample: %v", err)
	}

	return standardTime, nil
}

func FindStandardTimesBySwimmer(swimmer *swimming.Swimmer, course string, age int64, timeStandard TimeStandard, db storage.Database) ([]*StandardTime, error) {
	stm := `select st.id, st.course, st.style, st.distance, st.standard 
			from standard_time st
			where st.gender = $1
				and st.course = $2
				and (st.age = $3 or st.age is null)
				and st.time_standard = $4`

	rows, err := db.Query(context.Background(), stm, swimmer.Gender.String, course, age, timeStandard.ID)
	if err != nil && err.Error() != storage.ErrNoRows {
		return nil, fmt.Errorf("FindStandardTimesBySwimmer: %v", err)
	}
	defer rows.Close()

	var times []*StandardTime
	for rows.Next() {
		time := &StandardTime{}
		err = rows.Scan(&time.ID, &time.Course, &time.Style, &time.Distance, &time.Standard)
		if err != nil {
			return nil, fmt.Errorf("FindStandardTimesBySwimmer: %v", err)
		}
		times = append(times, time)
	}

	return times, nil
}

func findStandardsEvent(example StandardTime, db storage.Database) ([]*StandardTime, error) {
	stm := `select ts.id , ts.name, st.standard, ss.id, ss.name
			from standard_time st
				join time_standard ts on ts.id = st.time_standard
				join swim_season ss on ts.season = ss.id
			where st.age = $1 and st.gender = $2 and st.course = $3 and st.distance = $4 and st.style = $5
			order by ss.name desc, st.standard desc`

	rows, err := db.Query(context.Background(), stm, example.Age, example.Gender, example.Course, example.Distance, example.Style)
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

func FindMinAndMaxStandardsAges(db storage.Database) (int64, int64, error) {
	stm := `select min(age) as min_age, max(age) as max_age from standard_time`

	row := db.QueryRow(context.Background(), stm)

	var minAge, maxAge int64
	if err := row.Scan(&minAge, &maxAge); err != nil {
		return 0, 0, fmt.Errorf("FindMinAndMaxStandardsAges: %v", err)
	}

	return minAge, maxAge, nil
}

func FindMinAndMaxStandardAges(timeStandard *TimeStandard, db storage.Database) (int64, int64) {
	stm := `select min(st.age) as min_age, max(st.age) as max_age 
			from standard_time st
			where st.time_standard = $1`

	row := db.QueryRow(context.Background(), stm, timeStandard.ID)

	var minAge, maxAge int64
	if err := row.Scan(&minAge, &maxAge); err != nil {
		if !strings.Contains(err.Error(), "cannot scan NULL") {
			log.Printf("FindMinAndMaxStandardAges: %v", err)
		}
		return 0, 0
	}

	return minAge, maxAge
}

func FindMeetsWithTimeStandardByJurisdiction(jurisdictionId int64, db storage.Database) ([]*Meet, error) {
	stm := `select m.name, m.age_date, m.time_standard, m.course, ss.id, ss.name,
	            ts.min_age_time, ts.max_age_time, ts.open, m.min_age_enforced, m.max_age_enforced
			from meet m
			    join swim_season ss on ss.id = m.season
			    join time_standard ts on ts.id = m.time_standard
				join jurisdiction j on j.id = ts.jurisdiction
			where (j.id = $1 or j.region is null)
			    and ss.start_date <= now() and ss.end_date >= now()
			 	and m.end_date >= now()
			 	and m.time_standard is not null
				and m.age_date is not null
				and ts.benchmark = true
			order by m.age_date`
	rows, err := db.Query(context.Background(), stm, jurisdictionId)
	if err != nil {
		return nil, fmt.Errorf("FindChampionshipMeets: %v", err)
	}
	defer rows.Close()

	var meets []*Meet
	for rows.Next() {
		meet := &Meet{}
		err = rows.Scan(&meet.Name, &meet.AgeDate, &meet.TimeStandard.ID, &meet.Course, &meet.Season.ID, &meet.Season.Name,
			&meet.TimeStandard.MinAgeTime, &meet.TimeStandard.MaxAgeTime, &meet.TimeStandard.Open, &meet.MinAgeEnforced, &meet.MaxAgeEnforced)
		if err != nil && err.Error() != storage.ErrNoRows {
			return nil, fmt.Errorf("FindChampionshipMeets: %v", err)
		}
		meets = append(meets, meet)
	}

	return meets, nil
}

func FindMeetsByTimeStandard(timeStandard TimeStandard, db storage.Database) ([]*Meet, error) {
	stm := `select m.id, m.name, m.course
			from meet m
			where m.time_standard = $1
			order by m.name`
	rows, err := db.Query(context.Background(), stm, timeStandard.ID)
	if err != nil {
		return nil, fmt.Errorf("findStandardChampionshipMeets: %v", err)
	}
	defer rows.Close()

	var meets []*Meet
	for rows.Next() {
		meet := &Meet{}
		err = rows.Scan(&meet.ID, &meet.Name, &meet.Course)
		if err != nil && err.Error() != storage.ErrNoRows {
			return nil, fmt.Errorf("findStandardChampionshipMeets: %v", err)
		}
		meets = append(meets, meet)
	}

	return meets, nil
}

func GetMeet(id int64, db storage.Database) *Meet {
	stm := `select m.id, m.name, m.age_date, m.min_age_enforced, m.max_age_enforced
			from meet m
			where m.id = $1`
	row := db.QueryRow(context.Background(), stm, id)

	meet := &Meet{
		ID: id,
	}
	err := row.Scan(&meet.Name, &meet.Course, &meet.AgeDate, &meet.MinAgeEnforced, &meet.MaxAgeEnforced)
	if err != nil {
		if err.Error() == storage.ErrNoRows {
			return nil
		}
		log.Printf("GetMeet: %v", err)
	}

	return meet
}
