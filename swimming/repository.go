package swimming

import (
	"context"
	"fmt"
	"geekswimmers/storage"
)

func findStyles(db storage.Database) ([]*Style, error) {
	stmt := `select m.stroke, m.description
			 from swim_style m
			 order by m.sequence`
	rows, err := db.Query(context.Background(), stmt)
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
	stmt := `select m.id, m.description 
			 from swim_style m 
	         where m.stroke = $1`

	row := db.QueryRow(context.Background(), stmt, stroke)

	style := &Style{
		Stroke: stroke,
	}
	err := row.Scan(&style.ID, &style.Description)
	if err != nil && err.Error() != storage.ErrNoRows {
		return nil, fmt.Errorf("findStyle: %v", err)
	}

	return style, nil
}

func findInstructions(style *Style, db storage.Database) ([]*Instruction, error) {
	stmt := `select i.instruction, i.sequence
			 from swim_style_instruction i
			 where i.style = $1
			 order by i.sequence`
	rows, err := db.Query(context.Background(), stmt, style.ID)
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
