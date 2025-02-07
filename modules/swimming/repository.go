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
	stm := `select ssd.id, ssd.distance, ss.stroke
			 from swim_event ssd
			 join swim_style ss on ssd.style = ss.id
			 where ssd.distance = $1 and ss.stroke = $2`
	row := db.QueryRow(context.Background(), stm, event.Distance, event.Style.Stroke)

	ev := &Event{}
	err := row.Scan(&ev.ID, &ev.Distance, &ev.Style.Stroke)
	if err != nil && err.Error() != storage.ErrNoRows {
		return nil, fmt.Errorf("findEventByExample: %v", err)
	}

	return ev, nil
}

func FindEvents(db storage.Database) ([]*Event, error) {
	stm := `select ssd.distance , ss.stroke 
			 from swim_event ssd 
				join swim_style ss on ssd.style = ss.id
			 order by ss.sequence`
	rows, err := db.Query(context.Background(), stm)
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
