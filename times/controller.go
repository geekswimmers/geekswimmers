package times

import (
	"fmt"
	"geekswimmers/storage"
	"geekswimmers/utils"
	"geekswimmers/utils/reporting"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

type BenchmarkController struct {
	DB               storage.Database
	BaseTemplateData *utils.BaseTemplateData
}

type StandardsController struct {
	DB               storage.Database
	BaseTemplateData *utils.BaseTemplateData
}

type RecordsController struct {
	DB               storage.Database
	BaseTemplateData *utils.BaseTemplateData
}

func (bc *BenchmarkController) BenchmarkTime(res http.ResponseWriter, req *http.Request) {
	// Put all the fields in the session cookie
	fields := []string{"jurisdiction", "birthDate", "gender", "course", "event", "minute", "second", "millisecond"}
	for _, field := range fields {
		if err := storage.AddSessionEntry(res, req, "profile", field, req.URL.Query().Get(field)); err != nil {
			log.Printf("storage.%v", err)
		}
	}

	jurisdiction := req.URL.Query().Get("jurisdiction")
	birthDate, _ := time.Parse("2006-01-02", req.URL.Query().Get("birthDate"))
	gender := req.URL.Query().Get("gender")
	course := req.URL.Query().Get("course")
	event := strings.Split(req.URL.Query().Get("event"), "-")
	minute, _ := strconv.Atoi(req.URL.Query().Get("minute"))
	second, _ := strconv.Atoi(req.URL.Query().Get("second"))
	millisecond, _ := strconv.Atoi(req.URL.Query().Get("millisecond"))
	swimmerTime := utils.ToMiliseconds(minute, second, millisecond)

	// Separate the event into distance and stroke
	distance, _ := strconv.ParseInt(event[0], 10, 64)
	stroke := event[1]

	jurisdictionId, err := strconv.Atoi(jurisdiction)
	if err != nil {
		jurisdictionId = 0
	}
	meets, err := findChampionshipMeets(jurisdictionId, bc.DB)
	if err != nil {
		log.Printf("times.%v", err)
	}

	swimmer := &Swimmer{
		BirthDate: birthDate,
		Gender:    gender,
	}

	var foundMeets []*Meet
	for _, meet := range meets {
		meet.Age = swimmer.AgeAt(meet.AgeDate)
		searchAge := meet.Age

		if !meet.MinAgeEnforced && meet.TimeStandard.MinAgeTime != nil && meet.Age < *meet.TimeStandard.MinAgeTime {
			searchAge = *meet.TimeStandard.MinAgeTime
		} else if meet.MinAgeEnforced && meet.Age < *meet.TimeStandard.MinAgeTime {
			continue
		}

		if !meet.MaxAgeEnforced && meet.TimeStandard.MaxAgeTime != nil && meet.Age > *meet.TimeStandard.MaxAgeTime {
			searchAge = *meet.TimeStandard.MaxAgeTime
		} else if meet.MaxAgeEnforced && meet.Age > *meet.TimeStandard.MaxAgeTime {
			continue
		}

		standardTimeExample := StandardTime{
			Age:          searchAge,
			Gender:       gender,
			Course:       course,
			Style:        stroke,
			Distance:     distance,
			TimeStandard: meet.TimeStandard,
		}
		standardTime, err := findStandardTimeMeetByExample(standardTimeExample, meet.Season, bc.DB)
		if err != nil {
			log.Printf("times.%v", err)
		}

		if standardTime.Standard > 0 {
			standardTime.Difference = swimmerTime - standardTime.Standard

			if swimmerTime <= standardTime.Standard {
				standardTime.Percentage = 100
			} else {
				standardTime.Percentage = (standardTime.Standard * 100) / swimmerTime
			}
			meet.StandardTime = *standardTime
			foundMeets = append(foundMeets, meet)
		}
	}

	recordExample := RecordDefinition{
		Age:      swimmer.AgeAt(time.Now()),
		Gender:   gender,
		Course:   course,
		Style:    stroke,
		Distance: distance,
	}
	records, err := findRecordsByExample(recordExample, bc.DB)
	if err != nil {
		log.Printf("times.%v", err)
	}
	groupedRecords := groupRecordsByJurisdiction(records)

	for i, record := range groupedRecords {
		record.Difference = swimmerTime - record.Time

		if swimmerTime <= record.Time {
			record.Percentage = 100
		} else {
			record.Percentage = (record.Time * 100) / swimmerTime
		}
		groupedRecords[i] = record
	}

	sort.SliceStable(foundMeets, func(i, j int) bool {
		return foundMeets[i].StandardTime.Difference < foundMeets[j].StandardTime.Difference
	})

	sessionData := storage.NewSessionData(req)
	ctx := &benchmaskTimeViewData{
		Meets:            foundMeets,
		Records:          groupedRecords,
		FormatedTime:     utils.FormatTime(minute, second, millisecond),
		Distance:         distance,
		Course:           course,
		Style:            stroke,
		BaseTemplateData: bc.BaseTemplateData,
		SessionData:      sessionData,
	}

	html := utils.GetTemplateWithFunctions("base", "benchmark", template.FuncMap{
		"Title":             utils.Title,
		"FormatMiliseconds": utils.FormatMiliseconds,
		"Abs":               utils.Abs,
		"Lowercase":         utils.Lowercase,
	})

	err = html.Execute(res, ctx)
	if err != nil {
		log.Printf("times.BenchmarkTime: %v", err)
	}
}

func (sc *StandardsController) TimeStandardsView(res http.ResponseWriter, req *http.Request) {
	swimSeasonID, _ := strconv.ParseInt(req.URL.Query().Get("season"), 10, 64)
	swimSeason := &SwimSeason{
		ID: swimSeasonID,
	}

	swimSeasons, err := findSwimSeasons(sc.DB)
	if err != nil {
		log.Printf("times.%v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	if swimSeasonID == 0 {
		swimSeason.ID = swimSeasons[0].ID
	}

	timeStandards, err := findTimeStandards(*swimSeason, sc.DB)
	if err != nil {
		log.Printf("times.%v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	sessionData := storage.NewSessionData(req)
	ctx := &timeStandardsViewData{
		SwimSeason:       swimSeason,
		SwimSeasons:      swimSeasons,
		TimeStandards:    timeStandards,
		BaseTemplateData: sc.BaseTemplateData,
		SessionData:      sessionData,
	}

	html := utils.GetTemplate("base", "timestandards")
	err = html.Execute(res, ctx)
	if err != nil {
		log.Printf("times.TimeStandardsView: %v", err)
	}
}

func (sc *StandardsController) TimeStandardView(res http.ResponseWriter, req *http.Request) {
	sessionData := storage.NewSessionData(req)
	ctx := &timeStandardViewData{
		BaseTemplateData: sc.BaseTemplateData,
		SessionData:      sessionData,
	}

	id, _ := strconv.ParseInt(req.URL.Query().Get(":id"), 10, 64)
	timeStandard, err := findTimeStandard(id, sc.DB)
	if err != nil || timeStandard == nil {
		log.Printf("times.%v (%d)", err, id)
		utils.ErrorHandler(res, req, ctx, http.StatusNotFound)
		return
	}

	age, err := strconv.ParseInt(req.URL.Query().Get("age"), 10, 64)
	if err != nil {
		age = *timeStandard.MinAgeTime
	}
	if age < *timeStandard.MinAgeTime {
		age = *timeStandard.MinAgeTime
	}
	if timeStandard.MaxAgeTime != nil && age > *timeStandard.MaxAgeTime {
		age = *timeStandard.MaxAgeTime
	}

	gender := req.URL.Query().Get("gender")
	if gender == "" {
		gender = GenderFemale
	}
	course := req.URL.Query().Get("course")
	if course == "" {
		course = DefaultCourse
	}

	example := StandardTime{
		Age:          age,
		Gender:       gender,
		Course:       course,
		TimeStandard: *timeStandard,
	}
	standardTimes, err := findStandardTimes(example, sc.DB)
	if err != nil {
		log.Printf("times.%v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	ctx.Age = age
	ctx.Gender = gender
	ctx.Course = course
	ctx.TimeStandard = timeStandard
	ctx.StandardTimes = standardTimes

	if timeStandard.MaxAgeTime != nil {
		for i := *timeStandard.MinAgeTime; i <= *timeStandard.MaxAgeTime; i++ {
			ctx.Ages = append(ctx.Ages, i)
		}
	}

	latestTimeStandard, err := findLatestTimeStandard(timeStandard.ID, sc.DB)
	if condition := err == nil && latestTimeStandard != nil; condition {
		ctx.LatestTimeStandard = latestTimeStandard
	}

	meets, err := findStandardChampionshipMeets(*timeStandard, sc.DB)
	if err != nil {
		log.Printf("TimeStandardView.%v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
	ctx.Meets = meets

	html := utils.GetTemplateWithFunctions("base", "timestandard", template.FuncMap{
		"Title":             utils.Title,
		"FormatMiliseconds": utils.FormatMiliseconds,
	})
	err = html.Execute(res, ctx)
	if err != nil {
		log.Printf("times.TimeStandardView: %v", err)
	}
}

func (sc *RecordsController) RecordsListView(res http.ResponseWriter, req *http.Request) {
	recordSets, err := findRecordSets(sc.DB)
	if err != nil {
		log.Printf("times.%v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	sessionData := storage.NewSessionData(req)
	ctx := &recordsListViewData{
		RecordSets:       recordSets,
		BaseTemplateData: sc.BaseTemplateData,
		SessionData:      sessionData,
	}

	html := utils.GetTemplate("base", "records-list")
	err = html.Execute(res, ctx)
	if err != nil {
		log.Printf("times.RecordsListView: %v", err)
	}
}

func (rc *RecordsController) RecordsView(res http.ResponseWriter, req *http.Request) {
	sessionData := storage.NewSessionData(req)
	ctx := &recordsViewData{
		BaseTemplateData: rc.BaseTemplateData,
		SessionData:      sessionData,
	}

	recordSetId, _ := strconv.ParseInt(req.URL.Query().Get(":id"), 10, 64)
	recordSet, err := findRecordSet(recordSetId, rc.DB)
	if err != nil || recordSet == nil {
		log.Printf("times.%v (%d)", err, recordSetId)
		utils.ErrorHandler(res, req, ctx, http.StatusNotFound)
		return
	}

	var ageRanges []*RecordDefinition
	ageRanges, err = findRecordsAgeRanges(*recordSet, rc.DB)
	if err != nil {
		log.Printf("times.%v", err)
	}

	ageParam := req.URL.Query().Get("age")
	age, err := strconv.ParseInt(ageParam, 10, 64)
	if err != nil && len(ageParam) > 0 {
		minMaxAge := strings.Split(ageParam, "-")
		minAge, err := strconv.ParseInt(minMaxAge[0], 10, 64)
		if err == nil {
			age = minAge
		} else {
			maxAge, err := strconv.ParseInt(minMaxAge[1], 10, 64)
			if err == nil {
				age = maxAge
			} else {
				age = 0
			}
		}
	} else if len(ageParam) == 0 {
		if ageRanges[0].MinAge != nil {
			age = *ageRanges[0].MinAge
		} else if ageRanges[0].MaxAge != nil {
			age = *ageRanges[0].MaxAge
		} else {
			age = 0
		}
	}
	ctx.AgeRange = ageParam
	ctx.AgeRanges = ageRanges

	gender := req.URL.Query().Get("gender")
	if gender == "" {
		gender = GenderFemale
	}
	course := req.URL.Query().Get("course")
	if course == "" {
		course = DefaultCourse
	}

	definition := RecordDefinition{
		Age:    age,
		Gender: gender,
		Course: course,
	}
	records, err := findRecordsByRecordSet(*recordSet, definition, rc.DB)
	if err != nil {
		log.Printf("times.%v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
	groupedRecords := groupRecordsByDefinition(records)

	ctx.Age = age
	ctx.Gender = gender
	ctx.Course = course
	ctx.RecordSet = recordSet
	ctx.RecordDefinition = definition
	ctx.Records = groupedRecords

	html := utils.GetTemplateWithFunctions("base", "records", template.FuncMap{
		"Title":             utils.Title,
		"Lowercase":         utils.Lowercase,
		"FormatMiliseconds": utils.FormatMiliseconds,
	})
	err = html.Execute(res, ctx)
	if err != nil {
		log.Printf("times.RecordsView: %v", err)
	}
}

func (rc *RecordsController) RecordHistoryView(res http.ResponseWriter, req *http.Request) {
	sessionData := storage.NewSessionData(req)
	ctx := &recordHistoryViewData{
		BaseTemplateData: rc.BaseTemplateData,
		SessionData:      sessionData,
	}

	id, _ := strconv.ParseInt(req.URL.Query().Get(":id"), 10, 64)
	recordDefinition, err := getRecordDefinition(id, rc.DB)
	if err != nil || recordDefinition == nil {
		log.Printf("times.RecordHistoryView (%d): %v", id, err)
		utils.ErrorHandler(res, req, ctx, http.StatusNotFound)
		return
	}
	ctx.RecordDefinition = recordDefinition

	records, err := findRecordsByDefinition(*recordDefinition, rc.DB)
	if err != nil {
		log.Printf("times.RecordHistoryView (%d): %v", id, err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
	ctx.Records = records

	if len(records) > 0 {
		ctx.RecordSet = records[0].RecordSet
		ctx.Jurisdiction = records[0].RecordSet.Jurisdiction
	}

	html := utils.GetTemplateWithFunctions("base", "record-history", template.FuncMap{
		"Title":             utils.Title,
		"FormatMiliseconds": utils.FormatMiliseconds,
	})
	err = html.Execute(res, ctx)
	if err != nil {
		log.Printf("times.RecordHistoryView: %v", err)
	}
}

func (sc *RecordsController) RecordPosterView(res http.ResponseWriter, req *http.Request) {
	report := reporting.GetReportTemplate("records-club-poster")
	res.Header().Set("Content-Type", "image/svg+xml")

	id, _ := strconv.ParseInt(req.URL.Query().Get(":id"), 10, 64)
	resultSet := RecordSet{
		ID: id,
	}
	records, err := findRecordsPoster(resultSet, sc.DB)
	if err != nil {
		log.Printf("times.RecordPosterView: %v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	reportData := clubRecordsReportData{
		Records:    records,
		LastUpdate: time.Now(),
	}

	err = report.Execute(res, reportData)
	if err != nil {
		log.Printf("times.RecordPosterView: %v", err)
	}
}

func (sc *StandardsController) StandardsEventView(res http.ResponseWriter, req *http.Request) {
	sessionData := storage.NewSessionData(req)
	ctx := &standardsEventViewData{
		BaseTemplateData: sc.BaseTemplateData,
		SessionData:      sessionData,
	}

	// Represents the event in two parts: distance and stroke
	event := strings.Split(req.URL.Query().Get("event"), "-")

	distance, err := strconv.ParseInt(event[0], 10, 64)
	if err != nil {
		distance = 100
	}
	ctx.Distance = distance

	var stroke string
	if len(event) > 1 {
		stroke = event[1]
	}
	if stroke == "" {
		stroke = StrokeFree
	}
	ctx.Style = stroke
	ctx.Event = fmt.Sprintf("%d-%s", distance, stroke)

	min, max, err := findMinAndMaxStandardAges(sc.DB)
	if err != nil {
		log.Printf("times.%v", err)
	}

	age, err := strconv.ParseInt(req.URL.Query().Get("age"), 10, 64)
	if err != nil {
		age = min
	}
	if age < min {
		age = min
	}
	if age > max {
		age = max
	}
	ctx.Age = age

	for i := min; i <= max; i++ {
		ctx.Ages = append(ctx.Ages, i)
	}

	gender := req.URL.Query().Get("gender")
	if gender == "" {
		gender = GenderFemale
	}
	ctx.Gender = gender

	course := req.URL.Query().Get("course")
	if course == "" {
		course = DefaultCourse
	}
	ctx.Course = course

	example := StandardTime{
		Age:      age,
		Gender:   gender,
		Course:   course,
		Style:    stroke,
		Distance: distance,
	}

	standardsEvent, err := findStandardsEvent(example, sc.DB)
	if err != nil {
		log.Printf("times.%v (%d-%s)", err, distance, utils.Title(stroke))
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	ctx.StandardTimes = standardsEvent

	html := utils.GetTemplateWithFunctions("base", "standards-event", template.FuncMap{
		"Title":             utils.Title,
		"Lowercase":         utils.Lowercase,
		"FormatMiliseconds": utils.FormatMiliseconds,
	})
	err = html.Execute(res, ctx)
	if err != nil {
		log.Printf("times.StandardsEventView: %v", err)
	}
}
