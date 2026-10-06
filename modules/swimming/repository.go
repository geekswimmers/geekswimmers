package swimming

import (
	"context"
	"fmt"
	"geekswimmers/storage"
	"time"
)

func findStyles(db storage.Database) ([]*Style, error) {
	stm := `select m.stroke, m.description
			from swim_style m
			order by m.sequence`
	rows, err := db.Query(context.Background(), stm)
	if err != nil {
		return nil, fmt.Errorf("findStyles: %v", err)
	}
	defer rows.Close()

	var styles []*Style
	for rows.Next() {
		style := &Style{}
		err = rows.Scan(&style.Stroke, &style.Description)
		if err != nil && err.Error() != storage.ErrNoRows {
			return nil, fmt.Errorf("findStyles: %v", err)
		}
		styles = append(styles, style)
	}

	return styles, nil
}

func findStyle(stroke string, db storage.Database) (*Style, error) {
	stm := `select m.id, m.description, m.sequence 
			from swim_style m 
	        where m.stroke = $1`

	row := db.QueryRow(context.Background(), stm, stroke)

	style := &Style{
		Stroke: stroke,
	}
	err := row.Scan(&style.ID, &style.Description, &style.Sequence)
	if err != nil && err.Error() != storage.ErrNoRows {
		return nil, fmt.Errorf("findStyle: %v", err)
	}

	return style, nil
}

func findStyleBySequence(sequence int64, db storage.Database) (*Style, error) {
	stm := `select id, stroke
			from swim_style
			where sequence = $1`
	row := db.QueryRow(context.Background(), stm, sequence)

	style := &Style{}
	err := row.Scan(&style.ID, &style.Stroke)
	if err != nil && err.Error() != storage.ErrNoRows {
		return nil, fmt.Errorf("findStyleBySequence: %v", err)
	}
	return style, nil
}

func FindEvents(course string, db storage.Database) ([]*Event, error) {
	stm := `select se.distance , ss.stroke 
			from swim_event se 
				join swim_style ss on se.style = ss.id
			where se.course = $1
			    and ss.stroke in (select distinct style from standard_definition)
			order by ss.sequence asc, se.distance asc`
	rows, err := db.Query(context.Background(), stm, course)
	if err != nil {
		return nil, fmt.Errorf("findEvents: %v", err)
	}
	defer rows.Close()

	var events []*Event
	for rows.Next() {
		event := &Event{}
		err = rows.Scan(&event.Distance, &event.Style.Stroke)
		if err != nil && err.Error() != storage.ErrNoRows {
			return nil, fmt.Errorf("findEvents: %v", err)
		}
		events = append(events, event)
	}

	return events, nil
}

func findInstructions(style *Style, db storage.Database) ([]*Instruction, error) {
	stm := `select i.instruction, i.sequence
			from swim_style_instruction i
			where i.style = $1
			order by i.sequence`
	rows, err := db.Query(context.Background(), stm, style.ID)
	if err != nil {
		return nil, fmt.Errorf("findInstructions: %v", err)
	}
	defer rows.Close()

	var instructions []*Instruction
	for rows.Next() {
		instruction := &Instruction{
			Style: style,
		}
		err = rows.Scan(&instruction.Instruction, &instruction.Sequence)
		if err != nil && err.Error() != storage.ErrNoRows {
			return nil, fmt.Errorf("findInstructions: %v", err)
		}
		instructions = append(instructions, instruction)
	}

	return instructions, nil
}

func FindJurisdictionsInUseByStandards(db storage.Database) ([]*Jurisdiction, error) {
	stm := `select j.id, j.country, j.province, j.region, j.city, j.team, j.meet
			from jurisdiction j
			where j.id in (select distinct jurisdiction from time_standard)`

	rows, err := db.Query(context.Background(), stm)
	if err != nil {
		return nil, fmt.Errorf("FindJurisdictionsInUseByStandards: %v", err)
	}
	defer rows.Close()

	var jurisdictions []*Jurisdiction
	for rows.Next() {
		jurisdiction := &Jurisdiction{}
		err = rows.Scan(&jurisdiction.ID, &jurisdiction.Country, &jurisdiction.Province, &jurisdiction.Region, &jurisdiction.City, &jurisdiction.Team, &jurisdiction.Meet)
		if err != nil && err.Error() != storage.ErrNoRows {
			return nil, fmt.Errorf("FindJurisdictionsInUseByStandards: %v", err)
		}
		jurisdiction.Title = jurisdiction.GetTitle()
		jurisdiction.SubTitle = jurisdiction.GetSubTitle()
		jurisdictions = append(jurisdictions, jurisdiction)
	}

	return jurisdictions, nil
}

func FindJurisdictionsByLevel(level string, db storage.Database) ([]*Jurisdiction, error) {
	stm := `select j.id, j.country, j.province, j.region, j.city, j.team, j.meet
	        from jurisdiction j`

	switch level {
	case JurisdictionLevelMeet:
		stm = fmt.Sprintf("%v where j.meet is not null", stm)
	case JurisdictionLevelTeam:
		stm = fmt.Sprintf("%v where j.team is not null and j.meet is null", stm)
	case JurisdictionLevelCity:
		stm = fmt.Sprintf("%v where j.city is not null and j.team is null", stm)
	case JurisdictionLevelRegion:
		stm = fmt.Sprintf("%v where j.region is not null and j.city is null", stm)
	case JurisdictionLevelProvince:
		stm = fmt.Sprintf("%v where j.province is not null and j.region is null", stm)
	case JurisdictionLevelCountry:
		stm = fmt.Sprintf("%v where j.country is not null and j.province is null", stm)
	default:
		return []*Jurisdiction{}, nil
	}
	stm = fmt.Sprintf("%v order by country, province, region, city, team, meet", stm)

	rows, err := db.Query(context.Background(), stm)
	if err != nil {
		return nil, fmt.Errorf("findJurisdictionsByLevel: %v", err)
	}
	defer rows.Close()

	var jurisdictions []*Jurisdiction
	for rows.Next() {
		jurisdiction := &Jurisdiction{}
		err = rows.Scan(&jurisdiction.ID, &jurisdiction.Country, &jurisdiction.Province, &jurisdiction.Region, &jurisdiction.City, &jurisdiction.Team, &jurisdiction.Meet)
		if err != nil && err.Error() != storage.ErrNoRows {
			return nil, fmt.Errorf("findJurisdictionsByLevel: %v", err)
		}
		jurisdiction.Title = jurisdiction.GetTitle()
		jurisdiction.SubTitle = jurisdiction.GetSubTitle()
		jurisdictions = append(jurisdictions, jurisdiction)
	}

	return jurisdictions, nil
}

func FindTeamsByJurisdiction(jurisdiction Jurisdiction, db storage.Database) ([]*Team, error) {
	stm := `select c.id, c.full_name, c.acronym, c.website
	        from team c
	        where c.jurisdiction = $1
	        order by c.full_name`
	rows, err := db.Query(context.Background(), stm, jurisdiction.ID.Int64)
	if err != nil {
		return nil, fmt.Errorf("findTeamsByJurisdiction: %v", err)
	}
	defer rows.Close()

	var teams []*Team
	for rows.Next() {
		team := &Team{
			Jurisdiction: jurisdiction,
		}
		err = rows.Scan(&team.ID, &team.FullName, &team.Acronym, &team.WebSite)
		if err != nil && err.Error() != storage.ErrNoRows {
			return nil, fmt.Errorf("findTeamsByJurisdiction: %v", err)
		}
		teams = append(teams, team)
	}

	return teams, nil
}

func FindSwimmerByNationalNumber(nationalNumber string, db storage.Database) (*Swimmer, error) {
	stm := `select s.id, s.first_name, s.last_name, s.gender, s.national_number
	        from swimmer s
	        where s.national_number = $1`
	row := db.QueryRow(context.Background(), stm, nationalNumber)

	swimmer := &Swimmer{}
	err := row.Scan(&swimmer.ID, &swimmer.FirstName, &swimmer.LastName, &swimmer.Gender, &swimmer.NationalNumber)
	if err != nil && err.Error() == storage.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("FindSwimmerByNationalNumber: %v", err)
	}
	return swimmer, nil
}

func UpdateSwimmerNationalNumber(swimmerID int64, nationalNumber string, db storage.Database) error {
	stm := `update swimmer set national_number = $1 where id = $2 and national_number is null`
	_, err := db.Exec(context.Background(), stm, nationalNumber, swimmerID)
	if err != nil {
		return fmt.Errorf("UpdateSwimmerNationalNumber: %v", err)
	}
	return nil
}

func FindTeamByAcronym(acronym string, db storage.Database) (*Team, error) {
	stm := `select t.id, t.full_name, t.acronym
	        from team t
	        where t.acronym = $1`
	row := db.QueryRow(context.Background(), stm, acronym)

	team := &Team{}
	err := row.Scan(&team.ID, &team.FullName, &team.Acronym)
	if err != nil && err.Error() == storage.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("FindTeamByAcronym: %v", err)
	}
	return team, nil
}

func FindSwimmerByNameAndBirthDate(firstName, lastName string, birthDate time.Time, db storage.Database) (*Swimmer, error) {
	stm := `select s.id, s.first_name, s.last_name, s.gender
	        from swimmer s
	        where lower(s.first_name) = lower($1)
	          and lower(s.last_name) = lower($2)
	          and s.birth_date = $3`
	row := db.QueryRow(context.Background(), stm, firstName, lastName, birthDate)

	swimmer := &Swimmer{}
	err := row.Scan(&swimmer.ID, &swimmer.FirstName, &swimmer.LastName, &swimmer.Gender)
	if err != nil && err.Error() == storage.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("FindSwimmerByNameAndBirthDate: %v", err)
	}
	return swimmer, nil
}

func FindSwimEventByStrokeAndDistance(stroke string, distance int64, db storage.Database) (*Event, error) {
	stm := `select se.id, se.distance, ss.stroke
	        from swim_event se
	        join swim_style ss on se.style = ss.id
	        where ss.stroke = $1 and se.distance = $2`
	row := db.QueryRow(context.Background(), stm, stroke, distance)

	event := &Event{}
	err := row.Scan(&event.ID, &event.Distance, &event.Style.Stroke)
	if err != nil && err.Error() == storage.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("FindSwimEventByStrokeAndDistance: %v", err)
	}
	return event, nil
}

func FindSwimmer(swimmerId int64, db storage.Database) (*Swimmer, error) {
	stm := `select s.first_name, s.last_name
			from swimmer s
			where s.id = $1`

	row := db.QueryRow(context.Background(), stm, swimmerId)
	swimmer := &Swimmer{
		ID: swimmerId,
	}
	err := row.Scan(&swimmer.FirstName, &swimmer.LastName)
	if err != nil && err.Error() != storage.ErrNoRows {
		return nil, fmt.Errorf("FindSwimmer: %v", err)
	}
	return swimmer, nil
}
