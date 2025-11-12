package admin

import (
	"encoding/csv"
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
	timeStandards, err := times.FindAllTimeStandards(ac.DB)
	if err != nil {
		log.Printf("times.%v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	ctx := map[string]any{
		"TimeStandards":    timeStandards,
		"BaseTemplateData": ac.BaseTemplateData,
		"SessionData":      sessionData,
	}

	html := utils.GetTemplate("base", "admin-console")
	err = html.Execute(res, ctx)
	if err != nil {
		log.Printf("admin.ConsoleView: %v", err)
	}
}

func (ac *AdminController) TimeStandardView(res http.ResponseWriter, req *http.Request, sessionData *storage.SessionData) {
	ctx := map[string]any{
		"BaseTemplateData": ac.BaseTemplateData,
		"SessionData":      sessionData,
		"Error":            nil,
	}

	id, _ := strconv.ParseInt(req.URL.Query().Get(":id"), 10, 64)
	timeStandard, err := times.GetTimeStandard(id, ac.DB)
	if err != nil || timeStandard == nil {
		log.Printf("times.%v (%d)", err, id)
		utils.ErrorHandler(res, req, ctx, http.StatusNotFound)
		return
	}

	ctx["TimeStandard"] = timeStandard

	renderTimeStandardView(res, ctx)
}

func renderTimeStandardView(res http.ResponseWriter, ctx map[string]any) {
	html := utils.GetTemplateWithFunctions("base", "admin-standard", template.FuncMap{
		"Title": utils.Title,
	})
	err := html.Execute(res, ctx)
	if err != nil {
		log.Printf("admin.TimeStandardView: %v", err)
	}
}

func (ac *AdminController) TimeStandardForm(res http.ResponseWriter, req *http.Request, sessionData *storage.SessionData) {
	ctx := map[string]any{
		"BaseTemplateData": ac.BaseTemplateData,
		"SessionData":      sessionData,
	}

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
		renderTimeStandardView(res, ctx)
		return
	}

	// Validate date format
	date := req.FormValue("publicationDate")
	publicationDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		log.Printf("admin.TimeStandardForm: %v", err)
		ctx["PublicationDate"] = date
		ctx["ErrorPublicationDate"] = "Invalid publication date format. Please use YYYY-MM-DD."
		renderTimeStandardView(res, ctx)
		return
	}
	ctx["PublicationDate"] = publicationDate

	csvFile, header, err := req.FormFile("csvFile")
	if err != nil {
		log.Printf("admin.TimeStandardForm: %v", err)
		ctx["CsvFile"] = header.Filename
		ctx["ErrorCSVFile"] = "Error uploading the file. Please check the file format."
		renderTimeStandardView(res, ctx)
		return
	}
	defer csvFile.Close()
	ctx["CsvFile"] = header.Filename

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".csv" {
		ctx["ErrorCSVFile"] = "Invalid file extension. Please upload a CSV file."
		renderTimeStandardView(res, ctx)
		return
	}

	// Read CSV directly from the uploaded file
	reader := csv.NewReader(csvFile)

	standardTimes, err := reader.ReadAll()
	if err != nil {
		log.Printf("admin.TimeStandardForm: %v", err)
		ctx["ErrorCSVFile"] = "Error reading CSV file. Please check the file format."
		renderTimeStandardView(res, ctx)
		return
	}

	if len(standardTimes) < 2 {
		ctx["ErrorCSVFile"] = "CSV file is empty."
		renderTimeStandardView(res, ctx)
		return
	}

	importedRecords, notImportedRecords, failedRecords := processStandardRecords(standardTimes, timeStandard, publicationDate, ac.DB)
	ctx["ImportedRecords"] = importedRecords
	ctx["NotImportedRecords"] = notImportedRecords
	ctx["FailedRecords"] = failedRecords

	html := utils.GetTemplateWithFunctions("base", "admin-standard-update", template.FuncMap{
		"Title":              utils.Title,
		"FormatMilliseconds": utils.FormatMilliseconds,
	})
	err = html.Execute(res, ctx)
	if err != nil {
		log.Printf("admin.TimeStandardView: %v", err)
	}
}
