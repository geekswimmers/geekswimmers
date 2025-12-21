package times

import (
	"database/sql"
	"fmt"
	"geekswimmers/modules"
	"geekswimmers/modules/swimming"
	"geekswimmers/storage"
	"geekswimmers/utils"
	"geekswimmers/utils/reporting"
	"log"
	"net/http"
	"slices"
	"sort"
	"strconv"
	"strings"
	"text/template"
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
	ctx := modules.InitializeRequestContext(storage.NewSessionData(req), bc.BaseTemplateData)

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
	swimmerTime := utils.ToMilliseconds(minute, second, millisecond)

	ctx["Course"] = course
	ctx["FormatedTime"] = utils.FormatTime(minute, second, millisecond)

	// Separate the event into distance and stroke
	distance, _ := strconv.ParseInt(event[0], 10, 64)
	style := event[1]

	ctx["Distance"] = distance
	ctx["Style"] = style

	jurisdictionId, err := strconv.ParseInt(jurisdiction, 10, 64)
	if err != nil {
		jurisdictionId = 0
	}
	meets, err := FindMeetsWithTimeStandardByJurisdiction(jurisdictionId, bc.DB)
	if err != nil {
		log.Printf("times.%v", err)
	}

	swimmer := &swimming.Swimmer{
		BirthDate: sql.NullTime{
			Time:  birthDate,
			Valid: true,
		},
		Gender: sql.NullString{
			String: gender,
			Valid:  true,
		},
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

		standardDefinition := StandardDefinition{
			Age:      &searchAge,
			Gender:   gender,
			Course:   course,
			Style:    style,
			Distance: distance,
		}

		standardTimeExample := StandardTime{
			TimeStandard: meet.TimeStandard,
			Definition:   standardDefinition,
		}

		standardTime, err := GetStandardTimeMeetByExample(standardTimeExample, bc.DB)
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
		Style:    style,
		Distance: distance,
	}
	records, err := FindRecordsByExample(recordExample, bc.DB)
	if err != nil {
		log.Printf("times.%v", err)
	}
	for _, record := range records {
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

	ctx["Meets"] = foundMeets
	ctx["Records"] = records

	html := utils.GetTemplateWithFunctions("base", "benchmark", template.FuncMap{
		"Title":              utils.Title,
		"FormatMilliseconds": utils.FormatMilliseconds,
		"Abs":                utils.Abs,
		"Lowercase":          utils.Lowercase,
	})

	err = html.Execute(res, ctx)
	if err != nil {
		log.Printf("times.BenchmarkTime: %v", err)
	}
}

func (sc *StandardsController) TimeStandardsView(res http.ResponseWriter, req *http.Request) {
	ctx := modules.InitializeRequestContext(storage.NewSessionData(req), sc.BaseTemplateData)

	swimSeasonID, err := strconv.ParseInt(req.URL.Query().Get("season"), 10, 64)
	var swimSeason *SwimSeason
	if err != nil {
		swimSeason, err = GetLatestSwimSeason(sc.DB)
		if err != nil {
			log.Printf("times.%v", err)
			http.Error(res, err.Error(), http.StatusInternalServerError)
		}
	} else {
		swimSeason = &SwimSeason{
			ID: swimSeasonID,
		}
	}

	jurisdictionID, _ := strconv.ParseInt(req.URL.Query().Get("jurisdiction"), 10, 64)
	jurisdiction := swimming.Jurisdiction{
		ID: sql.NullInt64{
			Int64: jurisdictionID,
		},
	}

	ctx["Jurisdiction"] = jurisdiction

	seasons, err := findSwimSeasons(sc.DB)
	if err != nil {
		log.Printf("times.%v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	ctx["SwimSeasons"] = seasons

	jurisdictions, err := swimming.FindJurisdictionsByLevel(swimming.JurisdictionLevelRegion, sc.DB)
	if err != nil {
		log.Printf("times.%v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	ctx["Jurisdictions"] = jurisdictions

	if swimSeasonID == 0 && len(seasons) > 0 {
		swimSeason.ID = seasons[0].ID
	}

	ctx["SwimSeason"] = swimSeason

	timeStandards, err := FindTimeStandards(*swimSeason, jurisdiction, sc.DB)
	if err != nil {
		log.Printf("times.%v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	ctx["TimeStandards"] = timeStandards

	html := utils.GetTemplate("base", "timestandards")
	err = html.Execute(res, ctx)
	if err != nil {
		log.Printf("times.TimeStandardsView: %v", err)
	}
}

func (sc *StandardsController) TimeStandardView(res http.ResponseWriter, req *http.Request) {
	ctx := modules.InitializeRequestContext(storage.NewSessionData(req), sc.BaseTemplateData)

	id, _ := strconv.ParseInt(req.URL.Query().Get(":id"), 10, 64)
	timeStandard, err := GetTimeStandard(id, sc.DB)
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
		gender = swimming.GenderFemale
	}
	course := req.URL.Query().Get("course")
	if course == "" {
		course = swimming.DefaultCourse
	}

	standardDefinition := StandardDefinition{
		Age:    &age,
		Gender: gender,
		Course: course,
	}

	example := StandardTime{
		TimeStandard: *timeStandard,
		Definition:   standardDefinition,
	}
	standardTimes, err := findStandardTimes(example, sc.DB)
	if err != nil {
		log.Printf("times.%v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	ctx["Age"] = age
	ctx["Gender"] = gender
	ctx["Course"] = course
	ctx["TimeStandard"] = timeStandard
	ctx["StandardTimes"] = standardTimes

	var ages []int64
	if timeStandard.MaxAgeTime != nil {
		for i := *timeStandard.MinAgeTime; i <= *timeStandard.MaxAgeTime; i++ {
			ages = append(ages, i)
		}
	}

	ctx["Ages"] = ages

	meets, err := FindMeetsByTimeStandard(*timeStandard, sc.DB)
	if err != nil {
		log.Printf("TimeStandardView.%v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	ctx["Meets"] = meets

	html := utils.GetTemplateWithFunctions("base", "timestandard", template.FuncMap{
		"Title":              utils.Title,
		"FormatMilliseconds": utils.FormatMilliseconds,
	})
	err = html.Execute(res, ctx)
	if err != nil {
		log.Printf("times.TimeStandardView: %v", err)
	}
}

func (sc *RecordsController) RecordsListView(res http.ResponseWriter, req *http.Request) {
	ctx := modules.InitializeRequestContext(storage.NewSessionData(req), sc.BaseTemplateData)

	recordSets, err := findRecordSets(sc.DB)
	if err != nil {
		log.Printf("times.%v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	ctx["RecordSets"] = recordSets

	html := utils.GetTemplate("base", "records-list")
	err = html.Execute(res, ctx)
	if err != nil {
		log.Printf("times.RecordsListView: %v", err)
	}
}

func (rc *RecordsController) RecordsView(res http.ResponseWriter, req *http.Request) {
	ctx := modules.InitializeRequestContext(storage.NewSessionData(req), rc.BaseTemplateData)

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
	age := int64(0)
	if ageParam != "All" {
		age, err = strconv.ParseInt(ageParam, 10, 64)
		if err != nil && len(ageParam) > 0 {
			minMaxAge := strings.Split(ageParam, "-")
			minAge, err := strconv.ParseInt(minMaxAge[0], 10, 64)
			if err == nil {
				age = minAge
			} else {
				maxAge, err := strconv.ParseInt(minMaxAge[1], 10, 64)
				if err == nil {
					age = maxAge
				}
			}
		} else if len(ageParam) == 0 {
			if ageRanges[0].MinAge != nil {
				age = *ageRanges[0].MinAge
			} else if ageRanges[0].MaxAge != nil {
				age = *ageRanges[0].MaxAge
			}
		}
	}

	ctx["AgeRange"] = ageParam
	ctx["AgeRanges"] = ageRanges

	gender := req.URL.Query().Get("gender")
	if gender == "" {
		gender = swimming.GenderFemale
	}
	course := req.URL.Query().Get("course")
	if course == "" {
		course = swimming.DefaultCourse
	}

	definition := RecordDefinition{
		Age:    age,
		Gender: gender,
		Course: course,
	}
	records, err := FindRecordsByRecordSet(*recordSet, definition, rc.DB)
	if err != nil {
		log.Printf("times.%v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
	groupedRecords := groupRecordsByDefinition(records)

	ctx["Age"] = age
	ctx["Gender"] = gender
	ctx["Course"] = course
	ctx["RecordSet"] = recordSet
	ctx["RecordDefinition"] = definition
	ctx["Records"] = groupedRecords

	html := utils.GetTemplateWithFunctions("base", "records", template.FuncMap{
		"Title":              utils.Title,
		"Lowercase":          utils.Lowercase,
		"FormatMilliseconds": utils.FormatMilliseconds,
	})
	err = html.Execute(res, ctx)
	if err != nil {
		log.Printf("times.RecordsView: %v", err)
	}
}

func (rc *RecordsController) RecordHistoryView(res http.ResponseWriter, req *http.Request) {
	ctx := modules.InitializeRequestContext(storage.NewSessionData(req), rc.BaseTemplateData)

	id, _ := strconv.ParseInt(req.URL.Query().Get(":id"), 10, 64)
	recordSet := RecordSet{
		ID: id,
	}

	defId, _ := strconv.ParseInt(req.URL.Query().Get(":defId"), 10, 64)
	recordDefinition, err := getRecordDefinition(defId, rc.DB)
	if err != nil || recordDefinition == nil {
		log.Printf("times.RecordHistoryView (%d): %v", defId, err)
		utils.ErrorHandler(res, req, ctx, http.StatusNotFound)
		return
	}

	ctx["RecordDefinition"] = recordDefinition

	records, err := findRecordsByDefinition(*recordDefinition, recordSet, rc.DB)
	if err != nil {
		log.Printf("times.RecordHistoryView (%d): %v", defId, err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	ctx["Records"] = records

	if len(records) > 0 {
		ctx["RecordSet"] = records[0].RecordSet
		ctx["Jurisdiction"] = records[0].RecordSet.Jurisdiction
	}

	html := utils.GetTemplateWithFunctions("base", "record-history", template.FuncMap{
		"Title":              utils.Title,
		"FormatMilliseconds": utils.FormatMilliseconds,
	})
	err = html.Execute(res, ctx)
	if err != nil {
		log.Printf("times.RecordHistoryView: %v", err)
	}
}

func (rc *RecordsController) RecordSwimmerView(res http.ResponseWriter, req *http.Request) {
	sessionData := storage.NewSessionData(req)
	ctx := map[string]any{
		"BaseTemplateData": rc.BaseTemplateData,
		"SessionData":      sessionData,
	}

	id, _ := strconv.ParseInt(req.URL.Query().Get(":id"), 10, 64)
	recordSet, err := findRecordSet(id, rc.DB)
	if err != nil || recordSet == nil {
		log.Printf("times.RecordSwimmerView (%d): %v", id, err)
		utils.ErrorHandler(res, req, ctx, http.StatusNotFound)
		return
	}
	ctx["RecordSet"] = recordSet

	swimmerId, _ := strconv.ParseInt(req.URL.Query().Get(":swimmerId"), 10, 64)
	swimmer, err := swimming.FindSwimmer(swimmerId, rc.DB)
	if err != nil || swimmer == nil {
		log.Printf("times.RecordSwimmerView (%d): %v", swimmerId, err)
		utils.ErrorHandler(res, req, ctx, http.StatusNotFound)
		return
	}
	ctx["Swimmer"] = swimmer

	records, err := findRecordsBySwimmer(recordSet, swimmer, rc.DB)
	if err != nil {
		log.Printf("times.RecordSwimmerView (%d): %v", swimmerId, err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
	ctx["Records"] = records

	html := utils.GetTemplateWithFunctions("base", "record-holder", template.FuncMap{
		"Title":              utils.Title,
		"FormatMilliseconds": utils.FormatMilliseconds,
	})
	err = html.Execute(res, ctx)
	if err != nil {
		log.Printf("times.RecordSwimmerView: %v", err)
	}
}

func (rc *RecordsController) RecordPosterView(res http.ResponseWriter, req *http.Request) {
	id, _ := strconv.ParseInt(req.URL.Query().Get(":id"), 10, 64)
	recordSet := RecordSet{
		ID: id,
	}
	records, err := findRecordsPoster(recordSet, rc.DB)
	if err != nil {
		log.Printf("times.RecordPosterView: %v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
	records = groupPosterRecordsByDefinition(records)

	for _, record := range records {
		if record.Field == "TIME" {
			value, err := strconv.ParseInt(record.Value, 10, 64)
			if err != nil {
				log.Printf("Error parsing record.Value to int64: %v", err)
				continue
			}
			record.Value = utils.FormatMilliseconds(value)
		}
	}

	ctx := map[string]any{
		"Records":    records,
		"LastUpdate": time.Now(),
	}

	report := reporting.GetReportTemplate("records-team-poster")
	res.Header().Set("Content-Type", "image/svg+xml")

	err = report.Execute(res, ctx)
	if err != nil {
		log.Printf("times.RecordPosterView: %v", err)
	}
}

func (sc *StandardsController) StandardsEventView(res http.ResponseWriter, req *http.Request) {
	ctx := modules.InitializeRequestContext(storage.NewSessionData(req), sc.BaseTemplateData)

	id, _ := strconv.ParseInt(req.URL.Query().Get(":id"), 10, 64)
	timeStandard, err := GetTimeStandard(id, sc.DB)
	if err != nil || timeStandard == nil {
		log.Printf("times.%v (%d)", err, id)
		utils.ErrorHandler(res, req, ctx, http.StatusNotFound)
		return
	}
	ctx["TimeStandard"] = timeStandard

	definitionId, _ := strconv.ParseInt(req.URL.Query().Get(":eventId"), 10, 64)
	standardDefinition, err := GetStandardDefinition(definitionId, sc.DB)
	if err != nil || standardDefinition == nil {
		log.Printf("times.%v (%d)", err, definitionId)
		utils.ErrorHandler(res, req, ctx, http.StatusNotFound)
		return
	}

	ctx["Distance"] = standardDefinition.Distance
	ctx["Style"] = standardDefinition.Style
	ctx["Event"] = fmt.Sprintf("%d-%s", standardDefinition.Distance, standardDefinition.Style)

	minimum, maximum, err := FindMinAndMaxStandardsAges(sc.DB)
	if err != nil {
		log.Printf("times.%v", err)
	}

	ctx["Age"] = standardDefinition.Age

	var ages []int64
	for i := minimum; i <= maximum; i++ {
		ages = append(ages, i)
	}
	ctx["Ages"] = ages

	ctx["Gender"] = standardDefinition.Gender
	ctx["Course"] = standardDefinition.Course

	standardTimes, err := findStandardsEvent(timeStandard, standardDefinition, sc.DB)
	standardTimes = calculateDifferences(standardTimes)
	if err != nil {
		log.Printf("times.%v (%d-%s)", err, standardDefinition.Distance, utils.Title(standardDefinition.Style))
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	ctx["StandardTimes"] = standardTimes
	reversedStandardTimes := slices.Clone(standardTimes)
	slices.Reverse(reversedStandardTimes)
	ctx["InvertedStandardTimes"] = reversedStandardTimes

	html := utils.GetTemplateWithFunctions("base", "standards-event", template.FuncMap{
		"Title":              utils.Title,
		"Lowercase":          utils.Lowercase,
		"FormatMilliseconds": utils.FormatMilliseconds,
	})
	err = html.Execute(res, ctx)
	if err != nil {
		log.Printf("times.StandardsEventView: %v", err)
	}
}
