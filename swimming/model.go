package swimming

import (
	"geekswimmers/storage"
	"geekswimmers/utils"
)

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

type Event struct {
	Distance int64
	Stroke   string
}

type swimStylesViewData struct {
	Styles           []*Style
	BaseTemplateData *utils.BaseTemplateData
	SessionData      *storage.SessionData
}

type swimStyleViewData struct {
	Instructions     []*Instruction
	Style            *Style
	PreviousStyle    *Style
	NextStyle        *Style
	BaseTemplateData *utils.BaseTemplateData
	SessionData      *storage.SessionData
}
