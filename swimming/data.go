package swimming

import (
	"geekswimmers/storage"
	"geekswimmers/utils"
)

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
