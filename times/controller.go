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
	Stroke       string
	Meets        []*Meet
	FormatedTime string

	Age              int64
	AgeRange         string
	Gender           string
	TimeStandard     *TimeStandard
	Ages             []int64
	AgeRanges        []*RecordDefinition
	StandardTimes    []*StandardTime
	Records          []*Record
	Jurisdiction     *Jurisdiction
	RecordDefinition RecordDefinition

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

	if err := storage.AddSessionEntry(res, req, "profile", "birthDate", req.PostForm.Get("birthDate")); err != nil {
		log.Printf("storage.%v", err)
	}
	if err := storage.AddSessionEntry(res, req, "profile", "gender", req.PostForm.Get("gender")); err != nil {
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
		if !meet.MinAgeEnforced && meet.Age < meet.TimeStandard.MinAgeTime {
			searchAge = meet.TimeStandard.MinAgeTime
		} else if meet.MinAgeEnforced && meet.Age < meet.TimeStandard.MinAgeTime {
			continue
		}

		if !meet.MaxAgeEnforced && meet.Age > meet.TimeStandard.MaxAgeTime {
			searchAge = meet.TimeStandard.MaxAgeTime
		} else if meet.MaxAgeEnforced && meet.Age > meet.TimeStandard.MaxAgeTime {
			continue
		}

		standardTimeExample := StandardTime{
			Age:          searchAge,
			Gender:       gender,
			Course:       course,
			Stroke:       stroke,
			Distance:     distance,
			TimeStandard: meet.TimeStandard,
		}
		standardTime, err := findStandardTimeMeetByExample(standardTimeExample, meet.Season, bc.DB)
		if err != nil {
			log.Printf("times.%v", err)
		}

		if standardTime != nil {
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
		Stroke:   stroke,
		Distance: distance,
	}
	records, err := findRecordsByExample(recordExample, bc.DB)
	if err != nil {
		log.Printf("times.%v", err)
	}
	records = groupCurrentAndPreviousRecords(records)

	for _, record := range records {
		record.Jurisdiction.SetTitle(record.Definition.Age)
		record.Jurisdiction.SetSubTitle()

		record.Difference = swimmerTime - record.Time

		if swimmerTime <= record.Time {
			record.Percentage = 100
		} else {
			record.Percentage = (record.Time * 100) / swimmerTime
		}
	}

	sort.SliceStable(foundMeets, func(i, j int) bool {
		return foundMeets[i].StandardTime.Difference < foundMeets[j].StandardTime.Difference
	})

	ctx := &webContext{
		Meets:               foundMeets,
		Records:             records,
		FormatedTime:        utils.FormatTime(minute, second, milisecond),
		Distance:            distance,
		Course:              course,
		Stroke:              stroke,
		BaseTemplateContext: bc.BaseTemplateContext,
		AcceptedCookies:     storage.GetSessionValue(req, "profile", "acceptedCookies") == "true",
	}

	html := utils.GetTemplateWithFunctions("base", "benchmark", template.FuncMap{
		"Title":             utils.Title,
		"FormatMiliseconds": utils.FormatMiliseconds,
		"Abs":               utils.Abs,
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
		age = timeStandard.MinAgeTime
	}
	if age < timeStandard.MinAgeTime {
		age = timeStandard.MinAgeTime
	}
	if age > timeStandard.MaxAgeTime {
		age = timeStandard.MaxAgeTime
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

	for i := timeStandard.MinAgeTime; i <= timeStandard.MaxAgeTime; i++ {
		ctx.Ages = append(ctx.Ages, i)
	}

	html := utils.GetTemplateWithFunctions("base", "timestandard", template.FuncMap{
		"Title":             utils.Title,
		"FormatMiliseconds": utils.FormatMiliseconds,
	})
	err = html.Execute(res, ctx)
	if err != nil {
		log.Printf("times.TimeStandardView: %v", err)
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

	age, err := strconv.ParseInt(req.URL.Query().Get("age"), 10, 64)
	if err != nil {
		ctx.AgeRange = req.URL.Query().Get("age")
		minMaxAge := strings.Split(ctx.AgeRange, "-")
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
	}
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

	var ageRanges []*RecordDefinition
	ageRanges, err = findRecordsAgeRanges(*jurisdiction, rc.DB)
	if err != nil {
		log.Printf("times.%v", err)
	}

	ctx.Age = age
	ctx.AgeRanges = ageRanges
	ctx.Gender = gender
	ctx.Course = course
	ctx.Jurisdiction = jurisdiction
	ctx.RecordDefinition = definition
	ctx.Records = records

	html := utils.GetTemplateWithFunctions("base", "records", template.FuncMap{
		"Title":             utils.Title,
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
	ctx.Stroke = stroke
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
		Stroke:   stroke,
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
		"FormatMiliseconds": utils.FormatMiliseconds,
	})
	err = html.Execute(res, ctx)
	if err != nil {
		log.Printf("times.StandardsEventView: %v", err)
	}
}
