package web

import (
	"geekswimmers/modules/content"
	"geekswimmers/modules/swimming"
	"geekswimmers/storage"
	"geekswimmers/utils"
)

type homeViewData struct {
	Articles         []*content.Article
	Updates          []*content.ServiceUpdate
	Jurisdictions    []*swimming.Jurisdiction
	Events           []*swimming.Event
	Jurisdiction     string
	BirthDate        string
	Gender           string
	Course           string
	Event            string
	Minute           string
	Second           string
	Millisecond      string
	BaseTemplateData *utils.BaseTemplateData
	QuoteOfTheDay    *content.Quote
	SessionData      *storage.SessionData
}

type sitemapViewData struct {
	Articles    []*content.Article
	SessionData *storage.SessionData
}

type notFoundViewData struct {
	BaseTemplateData *utils.BaseTemplateData
	SessionData      *storage.SessionData
}

type LegalData struct {
	Title            string
	Content          string
	BaseTemplateData *utils.BaseTemplateData
	SessionData      *storage.SessionData
}
