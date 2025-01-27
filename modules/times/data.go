package times

import (
	"geekswimmers/storage"
	"geekswimmers/utils"
	"time"
)

type benchmaskTimeViewData struct {
	Distance         int64
	Course           string
	Style            string
	Meets            []*Meet
	FormatedTime     string
	Records          []Record
	BaseTemplateData *utils.BaseTemplateData
	SessionData      *storage.SessionData
}

type timeStandardsViewData struct {
	SwimSeason       *SwimSeason
	SwimSeasons      []*SwimSeason
	TimeStandards    []*TimeStandard
	BaseTemplateData *utils.BaseTemplateData
	SessionData      *storage.SessionData
}

type timeStandardViewData struct {
	Age                int64
	Gender             string
	Course             string
	TimeStandard       *TimeStandard
	StandardTimes      []*StandardTime
	Ages               []int64
	LatestTimeStandard *TimeStandard
	Meets              []*Meet
	BaseTemplateData   *utils.BaseTemplateData
	SessionData        *storage.SessionData
}

type recordsListViewData struct {
	RecordSets       []*RecordSet
	BaseTemplateData *utils.BaseTemplateData
	SessionData      *storage.SessionData
}

type recordsViewData struct {
	Age              int64
	AgeRange         string
	AgeRanges        []*RecordDefinition
	Gender           string
	Course           string
	RecordSet        *RecordSet
	RecordDefinition RecordDefinition
	Records          []Record
	BaseTemplateData *utils.BaseTemplateData
	SessionData      *storage.SessionData
}

type recordHistoryViewData struct {
	RecordDefinition *RecordDefinition
	RecordSet        RecordSet
	Records          []*Record
	Jurisdiction     Jurisdiction
	BaseTemplateData *utils.BaseTemplateData
	SessionData      *storage.SessionData
}

type standardsEventViewData struct {
	Age              int64
	Ages             []int64
	Course           string
	Distance         int64
	Event            string
	Gender           string
	Style            string
	StandardTimes    []*StandardTime
	BaseTemplateData *utils.BaseTemplateData
	SessionData      *storage.SessionData
}

type clubRecordsReportData struct {
	Records    []*RecordPoster
	LastUpdate time.Time
}
