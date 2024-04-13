package meets

type Modality struct {
	ID          int64
	Stroke      string
	Description string
}

type Instruction struct {
	ID          int64
	Modality    *Modality
	Instruction string
	Sequence    int64
}
