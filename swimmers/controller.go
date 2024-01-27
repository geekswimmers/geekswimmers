package swimmers

import (
	"geekswimmers/storage"
	"geekswimmers/utils"
	"geekswimmers/web"
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
	BaseTemplateContext web.BaseTemplateContext
}

type webContext struct {
	Example      StandardTime
	Distance     int64
	Course       string
	Stroke       string
	Meets        []*Meet
	FormatedTime string
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

	// Separate the event into distance and stroke
	distance, _ := strconv.ParseInt(event[0], 10, 64)
	stroke := event[1]

	var foundMeets []*Meet
	meets, err := FindChampionshipMeets(course, sc.DB)
	if err != nil {
		log.Print(err)
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

		// Find standard times in the database.
		standardTimeExample := StandardTime{
			Age:          searchAge,
			Gender:       gender,
			Course:       course,
			Stroke:       stroke,
			Distance:     distance,
			TimeStandard: meet.TimeStandard,
		}
		standardTime, err := FindStandardTimeMeet(standardTimeExample, meet.Season, sc.DB)
		if err != nil {
			log.Print(err)
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
		Meets:        foundMeets,
		FormatedTime: utils.FormatTime(minute, second, milisecond),
		Distance:     distance,
		Course:       course,
		Stroke:       stroke,
	}
	sc.BaseTemplateContext.Page = ctx

	html := utils.GetTemplateWithFunctions("base", "benchmark", template.FuncMap{
		"Title":             utils.Title,
		"FormatMiliseconds": utils.FormatMiliseconds,
		"Abs":               utils.Abs,
	})

	err = html.Execute(res, sc.BaseTemplateContext)
	if err != nil {
		log.Print(err)
	}
}
