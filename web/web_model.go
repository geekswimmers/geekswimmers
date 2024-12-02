package web

import (
	"geekswimmers/content"
	"geekswimmers/swimming"
	"geekswimmers/times"
	"geekswimmers/utils"
)

type HomeViewContext struct {
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

type SitemapViewContext struct {
	Articles        []*content.Article
	AcceptedCookies bool
}

type NotFoundViewContext struct {
	BaseTemplateContext *utils.BaseTemplateContext
	AcceptedCookies     bool
}
