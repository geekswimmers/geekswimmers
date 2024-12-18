package web

import (
	"geekswimmers/content"
	"geekswimmers/storage"
	"geekswimmers/swimming"
	"geekswimmers/times"
	"geekswimmers/utils"
)

type homeViewData struct {
	Articles         []*content.Article
	Updates          []*content.ServiceUpdate
	Jurisdictions    []*times.Jurisdiction
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
