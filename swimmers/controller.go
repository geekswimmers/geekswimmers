package swimmers

import (
	"fmt"
	"geekswimmers/storage"
	"geekswimmers/utils"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

type SwimmersController struct {
	DB storage.Database
}

type context struct {
	Example       StandardTime
	StandardTimes []*StandardTime
	FormatedTime  string
}

// ResetPassword
// post: /auth/password/reset/
func (sc *SwimmersController) BenchmarkTime(res http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	utils.Check(err)

	// Get values from the form
	age, _ := strconv.ParseInt(req.PostForm.Get("age"), 10, 64)
	gender := req.PostForm.Get("gender")
	course := req.PostForm.Get("course")
	event := strings.Split(req.PostForm.Get("event"), "-")
	minute, _ := strconv.Atoi(req.PostForm.Get("minute"))
	second, _ := strconv.Atoi(req.PostForm.Get("second"))
	milisecond, _ := strconv.Atoi(req.PostForm.Get("milisecond"))

	// Separate the event into distance and stroke
	distance, _ := strconv.ParseInt(event[0], 10, 64)
	stroke := event[1]

	// Find standard times in the database.
	standardTimeExample := StandardTime{
		Age:      age,
		Gender:   gender,
		Course:   course,
		Stroke:   stroke,
		Distance: distance,
	}
	standardTimes, _ := FindTimeStandards(standardTimeExample, sc.DB)

	// Calculate time difference and percentage of acomplishment
	for _, standardTime := range standardTimes {
		time := utils.ToMiliseconds(minute, second, milisecond)
		standardTime.Difference = time - standardTime.Standard
		fmt.Printf("%d , %d\n", time, standardTime.Standard)

		if time <= standardTime.Standard {
			standardTime.Percentage = 100
		} else {
			standardTime.Percentage = (standardTime.Standard * 100) / time
		}
		fmt.Printf("%d", standardTime.Percentage)
	}

	ctx := &context{
		Example:       standardTimeExample,
		StandardTimes: standardTimes,
		FormatedTime:  utils.FormatTime(minute, second, milisecond),
	}

	html := utils.GetTemplateWithFunctions("base", "benchmark", template.FuncMap{
		"Title":             utils.Title,
		"FormatMiliseconds": utils.FormatMiliseconds,
		"Abs":               utils.Abs,
	})

	err = html.Execute(res, ctx)
	utils.Check(err)
}
