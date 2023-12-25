package swimmers

import (
	"geekswimmers/storage"
	"geekswimmers/utils"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type SwimmersController struct {
	DB storage.Database
}

type context struct {
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
	meets, _ := FindChampionshipMeets(course, sc.DB)
	for _, meet := range meets {
		meet.Age = swimmer.AgeAt(meet.AgeDate)

		// Find standard times in the database.
		standardTimeExample := StandardTime{
			Age:          meet.Age,
			Gender:       gender,
			Course:       course,
			Stroke:       stroke,
			Distance:     distance,
			TimeStandard: meet.TimeStandard,
		}
		standardTime, _ := FindStandardTimeMeet(standardTimeExample, meet.Season, sc.DB)

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

	ctx := &context{
		Meets:        foundMeets,
		FormatedTime: utils.FormatTime(minute, second, milisecond),
		Distance:     distance,
		Course:       course,
		Stroke:       stroke,
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
