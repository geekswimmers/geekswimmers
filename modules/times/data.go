package times

import (
	"geekswimmers/modules/swimming"
	"geekswimmers/storage"
	"geekswimmers/utils"
)

type benchmaskTimeViewData struct {
	Distance         int64
	Course           string
	Style            string
	Meets            []*Meet
	FormatedTime     string
	Records          []*Record
	BaseTemplateData *utils.BaseTemplateData
	SessionData      *storage.SessionData
}

type timeStandardsViewData struct {
	Jurisdiction     swimming.Jurisdiction
	Jurisdictions    []*swimming.Jurisdiction
	SwimSeason       *SwimSeason
	SwimSeasons      []*SwimSeason
	TimeStandards    []*TimeStandard
	BaseTemplateData *utils.BaseTemplateData
	SessionData      *storage.SessionData
}

type timeStandardViewData struct {
	Age              int64
	Gender           string
	Course           string
	TimeStandard     *TimeStandard
	StandardTimes    []*StandardTime
	Ages             []int64
	Meets            []*Meet
	BaseTemplateData *utils.BaseTemplateData
	SessionData      *storage.SessionData
}

type recordsListViewData struct {
	RecordSets       []*RecordSet
	BaseTemplateData *utils.BaseTemplateData
	SessionData      *storage.SessionData
}

type recordHistoryViewData struct {
	RecordDefinition *RecordDefinition
	RecordSet        RecordSet
	Records          []*Record
	Jurisdiction     swimming.Jurisdiction
	BaseTemplateData *utils.BaseTemplateData
	SessionData      *storage.SessionData
}
