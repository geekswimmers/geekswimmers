package times

import (
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

type SwimmersController struct {
	DB                  storage.Database
	BaseTemplateContext *utils.BaseTemplateContext
}

type webContext struct {
	Distance     int64
	Course       string
	Stroke       string
	Meets        []*Meet
	FormatedTime string

	Age           int64
	Gender        string
	TimeStandard  *TimeStandard
	Ages          []int64
	StandardTimes []*StandardTime

	BaseTemplateContext *utils.BaseTemplateContext
	AcceptedCookies     bool
}

func (sc *SwimmersController) BenchmarkTime(res http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		log.Print(err)
	}

	// Get values from the form
	birthDate, _ := time.Parse("2006-01-02", req.PostForm.Get("birthDate"))
	gender := req.PostForm.Get("gender")
	course := req.PostForm.Get("course")
	event := strings.Split(req.PostForm.Get("event"), "-")
	minute, _ := strconv.Atoi(req.PostForm.Get("minute"))
	second, _ := strconv.Atoi(req.PostForm.Get("second"))
	milisecond, _ := strconv.Atoi(req.PostForm.Get("milisecond"))

	swimmer := &Swimmer{
		BirthDate: birthDate,
		Gender:    gender,
	}

	if err = storage.AddSessionEntry(res, req, "profile", "birthDate", req.PostForm.Get("birthDate")); err != nil {
		log.Printf("SwimmerController.BenchmarkTime: %v", err)
	}
	if err = storage.AddSessionEntry(res, req, "profile", "gender", req.PostForm.Get("gender")); err != nil {
		log.Printf("SwimmerController.BenchmarkTime: %v", err)
	}

	// Separate the event into distance and stroke
	distance, _ := strconv.ParseInt(event[0], 10, 64)
	stroke := event[1]

	var foundMeets []*Meet
	meets, err := findChampionshipMeets(sc.DB)
	if err != nil {
		log.Printf("swimmers.FindChampionshipMeets: %v", err)
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
		standardTime, err := findStandardTimeMeet(standardTimeExample, meet.Season, sc.DB)
		if err != nil {
			log.Printf("swimmers.FindStandardTimeMeet: %v", err)
		}

		if standardTime != nil {
			// Calculate time difference and percentage of acomplishment
			time := utils.ToMiliseconds(minute, second, milisecond)
			standardTime.Difference = time - standardTime.Standard

			if time <= standardTime.Standard {
				standardTime.Percentage = 100
			} else {
				standardTime.Percentage = (standardTime.Standard * 100) / time
			}
			meet.StandardTime = *standardTime
			foundMeets = append(foundMeets, meet)
		}
	}

	sort.SliceStable(foundMeets, func(i, j int) bool {
		return foundMeets[i].StandardTime.Difference < foundMeets[j].StandardTime.Difference
	})

	ctx := &webContext{
		Meets:               foundMeets,
		FormatedTime:        utils.FormatTime(minute, second, milisecond),
		Distance:            distance,
		Course:              course,
		Stroke:              stroke,
		BaseTemplateContext: sc.BaseTemplateContext,
		AcceptedCookies:     storage.GetSessionValue(req, "profile", "acceptedCookies") == "true",
	}

	html := utils.GetTemplateWithFunctions("base", "benchmark", template.FuncMap{
		"Title":             utils.Title,
		"FormatMiliseconds": utils.FormatMiliseconds,
		"Abs":               utils.Abs,
	})

	err = html.Execute(res, ctx)
	if err != nil {
		log.Print(err)
	}
}

func (sc *SwimmersController) TimeStandardView(res http.ResponseWriter, req *http.Request) {
	id, _ := strconv.ParseInt(req.URL.Query().Get(":id"), 10, 64)
	timeStandard, err := findTimeStandard(id, sc.DB)
	if err != nil {
		log.Printf("Error viewing the time standard: %v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	age, err := strconv.ParseInt(req.URL.Query().Get("age"), 10, 64)
	if err != nil {
		age = timeStandard.MinAgeTime
	}
	gender := req.URL.Query().Get("gender")
	course := req.URL.Query().Get("course")

	example := StandardTime{
		Age:          age,
		Gender:       gender,
		Course:       course,
		TimeStandard: *timeStandard,
	}
	standardTimes, err := findStandardTimes(example, sc.DB)
	if err != nil {
		log.Printf("Error viewing the time standard: %v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	ctx := &webContext{
		Age:                 age,
		Gender:              gender,
		Course:              course,
		TimeStandard:        timeStandard,
		StandardTimes:       standardTimes,
		BaseTemplateContext: sc.BaseTemplateContext,
		AcceptedCookies:     storage.GetSessionValue(req, "profile", "acceptedCookies") == "true",
	}

	for i := timeStandard.MinAgeTime; i <= timeStandard.MaxAgeTime; i++ {
		ctx.Ages = append(ctx.Ages, i)
	}

	html := utils.GetTemplateWithFunctions("base", "timestandard", template.FuncMap{
		"Title":             utils.Title,
		"FormatMiliseconds": utils.FormatMiliseconds,
	})
	err = html.Execute(res, ctx)
	if err != nil {
		log.Print(err)
	}
}
