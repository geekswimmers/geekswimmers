package swimming

import (
	"context"
	"fmt"
	"geekswimmers/storage"
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

func FindEventByExample(event Event, db storage.Database) (*Event, error) {
	stm := `select se.id, se.distance, ss.stroke
			 from swim_event se
			 join swim_style ss on se.style = ss.id
			 where se.distance = $1 and ss.stroke = $2 and se.course = $3`
	row := db.QueryRow(context.Background(), stm, event.Distance, event.Style.Stroke, event.Course)

	ev := &Event{}
	err := row.Scan(&ev.ID, &ev.Distance, &ev.Style.Stroke)
	if err != nil && err.Error() != storage.ErrNoRows {
		return nil, fmt.Errorf("findEventByExample: %v", err)
	}

	return ev, nil
}

func FindEvents(course string, db storage.Database) ([]*Event, error) {
	stm := `select se.distance , ss.stroke 
			 from swim_event se 
				join swim_style ss on se.style = ss.id
			 where se.course = $1
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

func FindJurisdictionsByLevel(level string, db storage.Database) ([]*Jurisdiction, error) {
	stm := `select j.id, j.country, j.province, j.region, j.city, j.club, j.meet
	         from jurisdiction j`

	if level == JurisdictionLevelMeet {
		stm = fmt.Sprintf("%v where j.meet is not null", stm)
	} else if level == JurisdictionLevelClub {
		stm = fmt.Sprintf("%v where j.club is not null and j.meet is null", stm)
	} else if level == JurisdictionLevelCity {
		stm = fmt.Sprintf("%v where j.city is not null and j.club is null", stm)
	} else if level == JurisdictionLevelRegion {
		stm = fmt.Sprintf("%v where j.region is not null and j.city is null", stm)
	} else if level == JurisdictionLevelProvince {
		stm = fmt.Sprintf("%v where j.province is not null and j.region is null", stm)
	} else if level == JurisdictionLevelCountry {
		stm = fmt.Sprintf("%v where j.country is not null and j.province is null", stm)
	} else {
		return []*Jurisdiction{}, nil
	}
	stm = fmt.Sprintf("%v order by country, province, region, city, club, meet", stm)

	rows, err := db.Query(context.Background(), stm)
	if err != nil {
		return nil, fmt.Errorf("findJurisdictionsByLevel: %v", err)
	}
	defer rows.Close()

	var jurisdictions []*Jurisdiction
	for rows.Next() {
		jurisdiction := &Jurisdiction{}
		err = rows.Scan(&jurisdiction.ID, &jurisdiction.Country, &jurisdiction.Province, &jurisdiction.Region, &jurisdiction.City, &jurisdiction.Club, &jurisdiction.Meet)
		if err != nil && err.Error() != storage.ErrNoRows {
			return nil, fmt.Errorf("findJurisdictionsByLevel: %v", err)
		}
		jurisdiction.SetTitle()
		jurisdiction.SetSubTitle()
		jurisdictions = append(jurisdictions, jurisdiction)
	}

	return jurisdictions, nil
}

func FindClubsByJurisdiction(jurisdiction Jurisdiction, db storage.Database) ([]*Club, error) {
	stm := `select c.id, c.full_name, c.acronym, c.website
	        from club c
	        where c.jurisdiction = $1
	        order by c.full_name`
	rows, err := db.Query(context.Background(), stm, jurisdiction.ID)
	if err != nil {
		return nil, fmt.Errorf("findClubsByJurisdiction: %v", err)
	}
	defer rows.Close()

	var clubs []*Club
	for rows.Next() {
		club := &Club{
			Jurisdiction: jurisdiction,
		}
		err = rows.Scan(&club.ID, &club.FullName, &club.Acronym, &club.WebSite)
		if err != nil && err.Error() != storage.ErrNoRows {
			return nil, fmt.Errorf("findClubsByJurisdiction: %v", err)
		}
		clubs = append(clubs, club)
	}

	return clubs, nil
}
