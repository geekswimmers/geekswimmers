package swimming

type Style struct {
	ID          int64
	Stroke      string
	Description string
	Sequence    int64
}

type Instruction struct {
	ID          int64
	Style       *Style
	Instruction string
	Sequence    int64
}
