package meets

import (
	"geekswimmers/storage"
	"geekswimmers/utils"
	"html/template"
	"log"
	"net/http"
)

type MeetController struct {
	DB                  storage.Database
	BaseTemplateContext *utils.BaseTemplateContext
}

type webContext struct {
	Instructions []*Instruction
	Styles       []*Style
	Style        *Style

	BaseTemplateContext *utils.BaseTemplateContext
	AcceptedCookies     bool
}

func (mc *MeetController) MeetStylesView(res http.ResponseWriter, req *http.Request) {
	styles, err := findStyles(mc.DB)
	if err != nil {
		log.Printf("meets.%v", err)
	}

	ctx := &webContext{
		Styles:              styles,
		BaseTemplateContext: mc.BaseTemplateContext,
		AcceptedCookies:     storage.GetSessionValue(req, "profile", "acceptedCookies") == "true",
	}

	html := utils.GetTemplateWithFunctions("base", "styles", template.FuncMap{
		"Title": utils.Title,
	})

	err = html.Execute(res, ctx)
	if err != nil {
		log.Printf("meets.MeetStylesView: %v", err)
	}
}

func (mc *MeetController) MeetStyleView(res http.ResponseWriter, req *http.Request) {
	stroke := req.URL.Query().Get(":stroke")

	style, err := findStyle(stroke, mc.DB)
	if err != nil {
		log.Printf("meets.%v", err)
	}

	instructions, err := findInstructions(style, mc.DB)
	if err != nil {
		log.Printf("meets.%v", err)
	}

	ctx := &webContext{
		Style:               style,
		Instructions:        instructions,
		BaseTemplateContext: mc.BaseTemplateContext,
		AcceptedCookies:     storage.GetSessionValue(req, "profile", "acceptedCookies") == "true",
	}

	html := utils.GetTemplateWithFunctions("base", "style", template.FuncMap{
		"Title": utils.Title,
	})

	err = html.Execute(res, ctx)
	if err != nil {
		log.Printf("meets.MeetStyleView: %v", err)
	}
}
