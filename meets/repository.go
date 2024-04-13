package meets

import (
	"context"
	"fmt"
	"geekswimmers/storage"
)

func findModalities(db storage.Database) ([]*Modality, error) {
	stmt := `select m.stroke, m.description
			 from swim_modality m
			 order by m.sequence`
	rows, err := db.Query(context.Background(), stmt)
	if err != nil {
		return nil, fmt.Errorf("findModalities: %v", err)
	}
	defer rows.Close()

	var modalities []*Modality
	for rows.Next() {
		modality := &Modality{}
		err = rows.Scan(&modality.Stroke, &modality.Description)
		if err != nil && err.Error() != storage.ErrNoRows {
			return nil, fmt.Errorf("findModalities: %v", err)
		}
		modalities = append(modalities, modality)
	}

	return modalities, nil
}

func findModality(stroke string, db storage.Database) (*Modality, error) {
	stmt := `select m.id, m.description 
			 from swim_modality m 
	         where m.stroke = $1`

	row := db.QueryRow(context.Background(), stmt, stroke)

	modality := &Modality{
		Stroke: stroke,
	}
	err := row.Scan(&modality.ID, &modality.Description)
	if err != nil && err.Error() != storage.ErrNoRows {
		return nil, fmt.Errorf("findModality: %v", err)
	}

	return modality, nil
}

func findInstructions(modality *Modality, db storage.Database) ([]*Instruction, error) {
	stmt := `select i.instruction, i.sequence
			 from swim_modality_instruction i
			 where i.modality = $1
			 order by i.sequence`
	rows, err := db.Query(context.Background(), stmt, modality.ID)
	if err != nil {
		return nil, fmt.Errorf("findInstructions: %v", err)
	}
	defer rows.Close()

	var instructions []*Instruction
	for rows.Next() {
		instruction := &Instruction{
			Modality: modality,
		}
		err = rows.Scan(&instruction.Instruction, &instruction.Sequence)
		if err != nil && err.Error() != storage.ErrNoRows {
			return nil, fmt.Errorf("findInstructions: %v", err)
		}
		instructions = append(instructions, instruction)
	}

	return instructions, nil
}
