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

	age, _ := strconv.ParseInt(req.PostForm.Get("age"), 10, 64)
	gender := req.PostForm.Get("gender")
	course := req.PostForm.Get("course")
	event := req.PostForm.Get("event")
	minute, _ := strconv.Atoi(req.PostForm.Get("minute"))
	second, _ := strconv.Atoi(req.PostForm.Get("second"))
	milisecond, _ := strconv.Atoi(req.PostForm.Get("milisecond"))

	// When the user types just one digit in the miliseconds field
	// they actually mean hundreds of miliseconds. Since we just
	// show 2 digits for the miliseconds, we simply multiply it by 10.
	if milisecond <= 10 {
		milisecond = milisecond * 10
	}

	distanceAndStroke := strings.Split(event, "-")
	distance, _ := strconv.ParseInt(distanceAndStroke[0], 10, 64)
	stroke := distanceAndStroke[1]

	standardTimeExample := StandardTime{
		Age:      age,
		Gender:   gender,
		Course:   course,
		Stroke:   stroke,
		Distance: distance,
	}
	standardTimes, _ := FindTimeStandards(standardTimeExample, sc.DB)
	for _, standardTime := range standardTimes {
		time := utils.ToMiliseconds(minute, second, milisecond)
		standardTime.Difference = time - standardTime.Standard
		fmt.Printf("%d / %d\n", time, standardTime.Standard)
		standardTime.Percentage = (time / standardTime.Standard) * 100
	}

	ctx := &context{
		Example:       standardTimeExample,
		StandardTimes: standardTimes,
		FormatedTime:  utils.FormatMiliseconds(utils.ToMiliseconds(minute, second, milisecond)),
	}

	html := utils.GetTemplateWithFunctions("base", "benchmark", template.FuncMap{
		"Title":             utils.FirstLetterUppercase,
		"FormatMiliseconds": utils.FormatMiliseconds,
		"Abs":               utils.Abs,
	})

	err = html.Execute(res, ctx)
	utils.Check(err)
}
