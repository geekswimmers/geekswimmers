package times

import (
	"fmt"
	"geekswimmers/storage"
	"geekswimmers/utils"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

type BenchmarkController struct {
	DB                  storage.Database
	BaseTemplateContext *utils.BaseTemplateContext
}

type StandardsController struct {
	DB                  storage.Database
	BaseTemplateContext *utils.BaseTemplateContext
}

type RecordsController struct {
	DB                  storage.Database
	BaseTemplateContext *utils.BaseTemplateContext
}

type webContext struct {
	Event        string
	Distance     int64
	Course       string
	Style        string
	Meets        []*Meet
	FormatedTime string

	Age              int64
	AgeRange         string
	Gender           string
	TimeStandard     *TimeStandard
	Ages             []int64
	AgeRanges        []*RecordDefinition
	StandardTimes    []*StandardTime
	Records          []Record
	Jurisdiction     *Jurisdiction
	Jurisdictions    []*Jurisdiction
	RecordDefinition RecordDefinition
	Source           Source

	SwimSeason    *SwimSeason
	SwimSeasons   []*SwimSeason
	TimeStandards []*TimeStandard

	BaseTemplateContext *utils.BaseTemplateContext
	AcceptedCookies     bool
}

func (bc *BenchmarkController) BenchmarkTime(res http.ResponseWriter, req *http.Request) {
	birthDate, _ := time.Parse("2006-01-02", req.URL.Query().Get("birthDate"))
	gender := req.URL.Query().Get("gender")
	course := req.URL.Query().Get("course")
	event := strings.Split(req.URL.Query().Get("event"), "-")

	minute, _ := strconv.Atoi(req.URL.Query().Get("minute"))
	second, _ := strconv.Atoi(req.URL.Query().Get("second"))
	milisecond, _ := strconv.Atoi(req.URL.Query().Get("milisecond"))
	swimmerTime := utils.ToMiliseconds(minute, second, milisecond)

	swimmer := &Swimmer{
		BirthDate: birthDate,
		Gender:    gender,
	}

	if err := storage.AddSessionEntry(res, req, "profile", "birthDate", req.URL.Query().Get("birthDate")); err != nil {
		log.Printf("storage.%v", err)
	}
	if err := storage.AddSessionEntry(res, req, "profile", "gender", req.URL.Query().Get("gender")); err != nil {
		log.Printf("storage.%v", err)
	}

	// Separate the event into distance and stroke
	distance, _ := strconv.ParseInt(event[0], 10, 64)
	stroke := event[1]

	var foundMeets []*Meet
	meets, err := findChampionshipMeets(bc.DB)
	if err != nil {
		log.Printf("times.%v", err)
	}

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
		record.RecordSet.Jurisdiction.SetTitle(record.Definition.Age)
		record.RecordSet.Jurisdiction.SetSubTitle()

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

	ctx := &webContext{
		Meets:               foundMeets,
		Records:             groupedRecords,
		FormatedTime:        utils.FormatTime(minute, second, milisecond),
		Distance:            distance,
		Course:              course,
		Style:               stroke,
		BaseTemplateContext: bc.BaseTemplateContext,
		AcceptedCookies:     storage.GetSessionValue(req, "profile", "acceptedCookies") == "true",
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

	ctx := &webContext{
		SwimSeason:          swimSeason,
		SwimSeasons:         swimSeasons,
		TimeStandards:       timeStandards,
		BaseTemplateContext: sc.BaseTemplateContext,
		AcceptedCookies:     storage.GetSessionValue(req, "profile", "acceptedCookies") == "true",
	}

	html := utils.GetTemplate("base", "timestandards")
	err = html.Execute(res, ctx)
	if err != nil {
		log.Printf("times.TimeStandardsView: %v", err)
	}
}

func (sc *StandardsController) TimeStandardView(res http.ResponseWriter, req *http.Request) {
	ctx := &webContext{
		BaseTemplateContext: sc.BaseTemplateContext,
		AcceptedCookies:     storage.GetSessionValue(req, "profile", "acceptedCookies") == "true",
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
		course = CourseLong
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
	jurisdictions, err := findJurisdictions(sc.DB)
	if err != nil {
		log.Printf("times.%v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	ctx := &webContext{
		Jurisdictions:       jurisdictions,
		BaseTemplateContext: sc.BaseTemplateContext,
		AcceptedCookies:     storage.GetSessionValue(req, "profile", "acceptedCookies") == "true",
	}

	html := utils.GetTemplate("base", "records-list")
	err = html.Execute(res, ctx)
	if err != nil {
		log.Printf("times.RecordsListView: %v", err)
	}
}

func (rc *RecordsController) RecordsView(res http.ResponseWriter, req *http.Request) {
	ctx := &webContext{
		BaseTemplateContext: rc.BaseTemplateContext,
		AcceptedCookies:     storage.GetSessionValue(req, "profile", "acceptedCookies") == "true",
	}

	id, _ := strconv.ParseInt(req.URL.Query().Get(":id"), 10, 64)
	jurisdiction, err := findJurisdiction(id, rc.DB)
	if err != nil || jurisdiction == nil {
		log.Printf("times.%v (%d)", err, id)
		utils.ErrorHandler(res, req, ctx, http.StatusNotFound)
		return
	}

	var ageRanges []*RecordDefinition
	ageRanges, err = findRecordsAgeRanges(*jurisdiction, rc.DB)
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

	jurisdiction.SetTitle(age)
	jurisdiction.SetSubTitle()

	gender := req.URL.Query().Get("gender")
	if gender == "" {
		gender = GenderFemale
	}
	course := req.URL.Query().Get("course")
	if course == "" {
		course = CourseLong
	}

	definition := RecordDefinition{
		Age:    age,
		Gender: gender,
		Course: course,
	}
	records, err := findRecordsByJurisdiction(*jurisdiction, definition, rc.DB)
	if err != nil {
		log.Printf("times.%v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
	groupedRecords := groupRecordsByDefinition(records)

	ctx.Age = age
	ctx.AgeRanges = ageRanges
	ctx.Gender = gender
	ctx.Course = course
	ctx.Jurisdiction = jurisdiction
	ctx.RecordDefinition = definition
	ctx.Records = groupedRecords

	if len(records) > 0 {
		ctx.Source = records[0].RecordSet.Source
	}

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

func (sc *StandardsController) StandardsEventView(res http.ResponseWriter, req *http.Request) {
	ctx := &webContext{
		BaseTemplateContext: sc.BaseTemplateContext,
		AcceptedCookies:     storage.GetSessionValue(req, "profile", "acceptedCookies") == "true",
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
		course = CourseLong
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
