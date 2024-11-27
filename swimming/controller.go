package swimming

import (
	"geekswimmers/storage"
	"geekswimmers/utils"
	"html/template"
	"log"
	"net/http"
	"strings"
)

type MeetController struct {
	DB                  storage.Database
	BaseTemplateContext *utils.BaseTemplateContext
}

type webContext struct {
	Instructions  []*Instruction
	Styles        []*Style
	Style         *Style
	PreviousStyle *Style
	NextStyle     *Style

	BaseTemplateContext *utils.BaseTemplateContext
	AcceptedCookies     bool
}

func (mc *MeetController) SwimStylesView(res http.ResponseWriter, req *http.Request) {
	styles, err := findStyles(mc.DB)
	if err != nil {
		log.Printf("meets.%v", err)
	}

	ctx := &webContext{
		Styles:              styles,
		BaseTemplateContext: mc.BaseTemplateContext,
		AcceptedCookies:     storage.GetSessionEntryValue(req, "profile", "acceptedCookies") == "true",
	}

	html := utils.GetTemplateWithFunctions("base", "styles", template.FuncMap{
		"Title":     utils.Title,
		"Lowercase": utils.Lowercase,
	})

	err = html.Execute(res, ctx)
	if err != nil {
		log.Printf("meets.MeetStylesView: %v", err)
	}
}

func (mc *MeetController) SwimStyleView(res http.ResponseWriter, req *http.Request) {
	stroke := req.URL.Query().Get(":stroke")

	style, err := findStyle(strings.ToUpper(stroke), mc.DB)
	if err != nil {
		log.Printf("meets.%v", err)
	}

	previousStyle, err := findStyleBySequence(style.Sequence-1, mc.DB)
	if err != nil {
		log.Printf("meets.%v", err)
	}

	nextStyle, err := findStyleBySequence(style.Sequence+1, mc.DB)
	if err != nil {
		log.Printf("meets.%v", err)
	}

	instructions, err := findInstructions(style, mc.DB)
	if err != nil {
		log.Printf("meets.%v", err)
	}

	ctx := &webContext{
		Style:               style,
		PreviousStyle:       previousStyle,
		NextStyle:           nextStyle,
		Instructions:        instructions,
		BaseTemplateContext: mc.BaseTemplateContext,
		AcceptedCookies:     storage.GetSessionEntryValue(req, "profile", "acceptedCookies") == "true",
	}

	html := utils.GetTemplateWithFunctions("base", "style", template.FuncMap{
		"Title":     utils.Title,
		"Lowercase": utils.Lowercase,
	})

	err = html.Execute(res, ctx)
	if err != nil {
		log.Printf("meets.MeetStyleView: %v", err)
	}
}
