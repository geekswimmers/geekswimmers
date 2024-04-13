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
	Modalities   []*Modality
	Modality     *Modality

	BaseTemplateContext *utils.BaseTemplateContext
	AcceptedCookies     bool
}

func (mc *MeetController) MeetModalitiesView(res http.ResponseWriter, req *http.Request) {
	modalities, err := findModalities(mc.DB)
	if err != nil {
		log.Printf("meets.%v", err)
	}

	ctx := &webContext{
		Modalities:          modalities,
		BaseTemplateContext: mc.BaseTemplateContext,
		AcceptedCookies:     storage.GetSessionValue(req, "profile", "acceptedCookies") == "true",
	}

	html := utils.GetTemplateWithFunctions("base", "modalities", template.FuncMap{
		"Title": utils.Title,
	})

	err = html.Execute(res, ctx)
	if err != nil {
		log.Printf("meets.MeetModalitiesView: %v", err)
	}
}

func (mc *MeetController) MeetModalityView(res http.ResponseWriter, req *http.Request) {
	stroke := req.URL.Query().Get(":stroke")

	modality, err := findModality(stroke, mc.DB)
	if err != nil {
		log.Printf("meets.%v", err)
	}

	instructions, err := findInstructions(modality, mc.DB)
	if err != nil {
		log.Printf("meets.%v", err)
	}

	ctx := &webContext{
		Modality:            modality,
		Instructions:        instructions,
		BaseTemplateContext: mc.BaseTemplateContext,
		AcceptedCookies:     storage.GetSessionValue(req, "profile", "acceptedCookies") == "true",
	}

	html := utils.GetTemplateWithFunctions("base", "modality", template.FuncMap{
		"Title": utils.Title,
	})

	err = html.Execute(res, ctx)
	if err != nil {
		log.Printf("meets.MeetModalityView: %v", err)
	}
}
