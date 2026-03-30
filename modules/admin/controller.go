package admin

import (
	"encoding/csv"
	"fmt"
	"geekswimmers/modules"
	"geekswimmers/modules/content"
	"geekswimmers/modules/times"
	"geekswimmers/storage"
	"geekswimmers/utils"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"
)

type AdminController struct {
	DB               storage.Database
	BaseTemplateData *utils.BaseTemplateData
}

func (ac *AdminController) ConsoleView(res http.ResponseWriter, _ *http.Request, sessionData *storage.SessionData) {
	ctx := modules.InitializeRequestContext(sessionData, ac.BaseTemplateData)

	html := utils.GetTemplate("admin", "admin-console")
	err := html.Execute(res, ctx)
	if err != nil {
		log.Printf("admin.ConsoleView: %v", err)
	}
}

func (ac *AdminController) StandardsView(res http.ResponseWriter, _ *http.Request, sessionData *storage.SessionData) {
	ctx := modules.InitializeRequestContext(sessionData, ac.BaseTemplateData)

	timeStandards, err := times.FindAllTimeStandards(ac.DB)
	if err != nil {
		log.Printf("times.%v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	ctx["TimeStandards"] = timeStandards

	html := utils.GetTemplate("admin", "admin-standards")
	err = html.Execute(res, ctx)
	if err != nil {
		log.Printf("admin.ConsoleView: %v", err)
	}
}

func (ac *AdminController) ServiceUpdatesView(res http.ResponseWriter, _ *http.Request, sessionData *storage.SessionData) {
	ctx := modules.InitializeRequestContext(sessionData, ac.BaseTemplateData)

	serviceUpdates, err := content.FindServiceUpdates(ac.DB)
	if err != nil {
		log.Printf("admin.%v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	ctx["ServiceUpdates"] = serviceUpdates

	html := utils.GetTemplate("admin", "admin-updates")
	err = html.Execute(res, ctx)
	if err != nil {
		log.Printf("admin.ConsoleView: %v", err)
	}
}

func (ac *AdminController) MeetsView(res http.ResponseWriter, _ *http.Request, sessionData *storage.SessionData) {
	ctx := modules.InitializeRequestContext(sessionData, ac.BaseTemplateData)

	html := utils.GetTemplate("admin", "admin-results")
	err := html.Execute(res, ctx)
	if err != nil {
		log.Printf("admin.ConsoleView: %v", err)
	}
}

func (ac *AdminController) TimeStandardFormView(res http.ResponseWriter, req *http.Request, sessionData *storage.SessionData) {
	ctx := modules.InitializeRequestContext(sessionData, ac.BaseTemplateData)

	ctx["Error"] = nil

	id, _ := strconv.ParseInt(req.URL.Query().Get(":id"), 10, 64)
	timeStandard, err := times.GetTimeStandard(id, ac.DB)
	if err != nil || timeStandard == nil {
		log.Printf("times.%v (%d)", err, id)
		utils.ErrorHandler(res, req, ctx, http.StatusNotFound)
		return
	}

	ctx["TimeStandard"] = timeStandard

	renderTimeStandardFormView(res, ctx)
}

func (ac *AdminController) TimeStandardForm(res http.ResponseWriter, req *http.Request, sessionData *storage.SessionData) {
	ctx := modules.InitializeRequestContext(sessionData, ac.BaseTemplateData)

	timeStandardId, _ := strconv.ParseInt(req.URL.Query().Get(":id"), 10, 64)
	timeStandard, err := times.GetTimeStandard(timeStandardId, ac.DB)
	if err != nil || timeStandard == nil {
		log.Printf("times.%v (%d)", err, timeStandardId)
		utils.ErrorHandler(res, req, ctx, http.StatusNotFound)
		return
	}

	ctx["TimeStandard"] = timeStandard

	// Limit upload file size to 1MB
	err = req.ParseMultipartForm(1 << 20)
	if err != nil {
		log.Printf("admin.TimeStandardForm: %v", err)
		ctx["ErrorCSVFile"] = "Error uploading the file. It is larger than 1MB."
		renderTimeStandardFormView(res, ctx)
		return
	}

	// Validate date format
	date := req.FormValue("publicationDate")
	publicationDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		log.Printf("admin.TimeStandardForm: %v", err)
		ctx["PublicationDate"] = date
		ctx["ErrorPublicationDate"] = "Invalid publication date format. Please use YYYY-MM-DD."
		renderTimeStandardFormView(res, ctx)
		return
	}
	ctx["PublicationDate"] = publicationDate

	csvFile, header, err := req.FormFile("csvFile")
	if err != nil {
		log.Printf("admin.TimeStandardForm: %v", err)
		ctx["CsvFile"] = header.Filename
		ctx["ErrorCSVFile"] = "Error uploading the file. Please check the file format."
		renderTimeStandardFormView(res, ctx)
		return
	}
	defer utils.CloseMultipartFile(csvFile)

	ctx["CsvFile"] = header.Filename

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".csv" {
		ctx["ErrorCSVFile"] = "Invalid file extension. Please upload a CSV file."
		renderTimeStandardFormView(res, ctx)
		return
	}

	// Read CSV directly from the uploaded file
	reader := csv.NewReader(csvFile)

	standardTimes, err := reader.ReadAll()
	if err != nil {
		log.Printf("admin.TimeStandardForm: %v", err)
		ctx["ErrorCSVFile"] = "Error reading CSV file. Please check the file format."
		renderTimeStandardFormView(res, ctx)
		return
	}

	if len(standardTimes) < 2 {
		ctx["ErrorCSVFile"] = "CSV file is empty."
		renderTimeStandardFormView(res, ctx)
		return
	}

	importedRecords, notImportedRecords, failedRecords := processStandardRecords(standardTimes, timeStandard, publicationDate, ac.DB)
	ctx["ImportedRecords"] = importedRecords
	ctx["NotImportedRecords"] = notImportedRecords
	ctx["FailedRecords"] = failedRecords

	html := utils.GetTemplateWithFunctions("admin", "admin-standards-update", template.FuncMap{
		"Title":              utils.Title,
		"FormatMilliseconds": utils.FormatMilliseconds,
	})
	err = html.Execute(res, ctx)
	if err != nil {
		log.Printf("admin.TimeStandardFormView: %v", err)
	}
}

func renderTimeStandardFormView(res http.ResponseWriter, ctx map[string]any) {
	html := utils.GetTemplateWithFunctions("admin", "admin-standards-form", template.FuncMap{
		"Title": utils.Title,
	})
	err := html.Execute(res, ctx)
	if err != nil {
		log.Printf("admin.TimeStandardFormView: %v", err)
	}
}

func (ac *AdminController) ServiceUpdateFormView(res http.ResponseWriter, req *http.Request, sessionData *storage.SessionData) {
	ctx := modules.InitializeRequestContext(sessionData, ac.BaseTemplateData)
	ctx["Error"] = nil
	ctx["ID"] = 0
	ctx["Published"] = time.Now()

	renderServiceUpdateFormView(res, ctx)
}

func (ac *AdminController) ServiceUpdateFormEditView(res http.ResponseWriter, req *http.Request, sessionData *storage.SessionData) {
	ctx := modules.InitializeRequestContext(sessionData, ac.BaseTemplateData)
	ctx["Error"] = nil

	id, _ := strconv.ParseInt(req.URL.Query().Get(":id"), 10, 64)
	serviceUpdate, err := content.GetServiceUpdate(id, ac.DB)
	if err != nil {
		log.Printf("admin.ServiceUpdateFormEditView: %v", err)
		utils.ErrorHandler(res, req, ctx, http.StatusNotFound)
		return
	}

	ctx["ID"] = id
	ctx["Title"] = serviceUpdate.Title
	ctx["Content"] = serviceUpdate.Content
	ctx["Published"] = serviceUpdate.Published

	renderServiceUpdateFormView(res, ctx)
}

func (ac *AdminController) ServiceUpdateForm(res http.ResponseWriter, req *http.Request, sessionData *storage.SessionData) {
	ctx := modules.InitializeRequestContext(sessionData, ac.BaseTemplateData)

	err := req.ParseForm()
	if err != nil {
		log.Print(err)
	}

	id, _ := strconv.ParseInt(req.FormValue("id"), 10, 64)

	ctx["ID"] = id
	ctx["Error"] = nil
	ctx["Title"] = req.PostForm.Get("title")
	ctx["Content"] = req.PostForm.Get("content")

	published, err := time.Parse("2006-01-02", req.PostForm.Get("publicationDate"))
	if err != nil {
		ctx["ErrorPublicationDate"] = "Invalid publication date format. Please use YYYY-MM-DD."
		renderServiceUpdateFormView(res, ctx)
		return
	}
	ctx["Published"] = published.Format("2006-01-02")

	serviceUpdate := &content.ServiceUpdate{
		ID:        id,
		Title:     req.PostForm.Get("title"),
		Content:   req.PostForm.Get("content"),
		Published: published,
	}

	if id > 0 {
		err = content.SaveServiceUpdate(serviceUpdate, ac.DB)
	} else {
		err = content.InsertServiceUpdate(serviceUpdate, ac.DB)
	}

	if err != nil {
		errorMessage := fmt.Sprintf("Error inserting service update: %v", err)
		log.Print(errorMessage)
		ctx["Error"] = errorMessage
		renderServiceUpdateFormView(res, ctx)
	}

	http.Redirect(res, req, "/admin/updates/", http.StatusSeeOther)
}

func renderServiceUpdateFormView(res http.ResponseWriter, ctx map[string]any) {
	html := utils.GetTemplate("admin", "admin-updates-form")
	err := html.Execute(res, ctx)
	if err != nil {
		log.Printf("admin.ServiceUpdateFormView: %v", err)
	}
}
