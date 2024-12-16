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
	DB               storage.Database
	BaseTemplateData *utils.BaseTemplateData
}

func (mc *MeetController) SwimStylesView(res http.ResponseWriter, req *http.Request) {
	styles, err := findStyles(mc.DB)
	if err != nil {
		log.Printf("meets.%v", err)
	}

	ctx := &swimStylesViewData{
		Styles:           styles,
		BaseTemplateData: mc.BaseTemplateData,
		AcceptedCookies:  storage.GetSessionEntryValue(req, "profile", "acceptedCookies") == "true",
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

	ctx := &swimStyleViewData{
		Style:            style,
		PreviousStyle:    previousStyle,
		NextStyle:        nextStyle,
		Instructions:     instructions,
		BaseTemplateData: mc.BaseTemplateData,
		AcceptedCookies:  storage.GetSessionEntryValue(req, "profile", "acceptedCookies") == "true",
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
