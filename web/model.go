package web

import (
	"geekswimmers/content"
	"geekswimmers/swimming"
	"geekswimmers/times"
	"geekswimmers/utils"
)

type homeViewData struct {
	Username            string
	Articles            []*content.Article
	Updates             []*content.ServiceUpdate
	Jurisdictions       []*times.Jurisdiction
	Events              []*swimming.Event
	Jurisdiction        string
	BirthDate           string
	Gender              string
	Course              string
	Event               string
	Minute              string
	Second              string
	Millisecond         string
	BaseTemplateContext *utils.BaseTemplateContext
	AcceptedCookies     bool
	QuoteOfTheDay       *content.Quote
}

type sitemapViewData struct {
	Articles        []*content.Article
	AcceptedCookies bool
}

type notFoundViewData struct {
	BaseTemplateContext *utils.BaseTemplateContext
	AcceptedCookies     bool
}
