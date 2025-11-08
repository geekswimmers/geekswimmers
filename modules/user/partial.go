package user

import (
	"geekswimmers/storage"
	"geekswimmers/utils"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

func (uc *Controller) BestTimesPartial(res http.ResponseWriter, req *http.Request, sessionData *storage.SessionData) {
	id := req.URL.Query().Get(":id")
	swimmerId, _ := strconv.ParseInt(id, 10, 64)
	swimmer := FindSwimmerByID(swimmerId, uc.DB)

	course := req.URL.Query().Get("course")

	bestTimes, err := findSwimmerBestTimes(swimmer, course, uc.DB)
	if err != nil {
		log.Printf("swimming.BestTimesPartial: %v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
	bestTimes = addSwimmingPointsToRecords(swimmer, course, bestTimes, uc.DB)

	data := &swimmerData{
		Swimmer:   swimmer,
		BestTimes: bestTimes,
		Course:    course,
	}

	html, err := template.New("bestTimes").Funcs(
		template.FuncMap{
			"Title":              utils.Title,
			"FormatMilliseconds": utils.FormatMilliseconds,
		}).Parse(
		`{{range .BestTimes}}
			<tr>
				<td><a href="/profile/swimmers/{{$.Swimmer.ID}}/besttimes/{{.ID}}/">{{.Event.Distance}} {{.Event.Style.Stroke | Title}}</a></td>
				<td>{{.BestTime | FormatMilliseconds}}</td>
				<td>{{.Points}}</td>
			</tr>
		{{else}}
			<tr>
				<td colspan="3">No best time listed yet.</td>
			</tr>
		{{end}}`)
	if err != nil {
		log.Printf("swimming.BestTimesPartial: %v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	err = html.Execute(res, data)
	if err != nil {
		log.Print(err)
	}
}
