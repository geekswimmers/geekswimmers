package swimming

import (
	"geekswimmers/modules"
	"geekswimmers/storage"
	"geekswimmers/utils"
	"html/template"
	"log"
	"net/http"
	"strings"
)

type Controller struct {
	DB               storage.Database
	BaseTemplateData *utils.BaseTemplateData
}

func (mc *Controller) SwimStylesView(res http.ResponseWriter, req *http.Request) {
	ctx := modules.InitializeRequestContext(storage.NewSessionData(req), mc.BaseTemplateData)

	styles, err := findStyles(mc.DB)
	if err != nil {
		log.Printf("meets.%v", err)
	}

	ctx["Styles"] = styles

	html := utils.GetTemplateWithFunctions("base", "styles", template.FuncMap{
		"Title":     utils.Title,
		"Lowercase": utils.Lowercase,
	})

	err = html.Execute(res, ctx)
	if err != nil {
		log.Printf("meets.MeetStylesView: %v", err)
	}
}

func (mc *Controller) SwimStyleView(res http.ResponseWriter, req *http.Request) {
	ctx := modules.InitializeRequestContext(storage.NewSessionData(req), mc.BaseTemplateData)

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

	ctx["Style"] = style
	ctx["PreviousStyle"] = previousStyle
	ctx["NextStyle"] = nextStyle
	ctx["Instructions"] = instructions

	html := utils.GetTemplateWithFunctions("base", "style", template.FuncMap{
		"Title":     utils.Title,
		"Lowercase": utils.Lowercase,
	})

	err = html.Execute(res, ctx)
	if err != nil {
		log.Printf("meets.MeetStyleView: %v", err)
	}
}
