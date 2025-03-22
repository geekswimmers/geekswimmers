package user

import (
	"encoding/json"
	"geekswimmers/storage"
	"geekswimmers/utils"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

func (uc *Controller) EventsResource(res http.ResponseWriter, req *http.Request) {
	course := req.URL.Query().Get("course")

	id := req.URL.Query().Get(":id")
	swimmerId, _ := strconv.ParseInt(id, 10, 64)
	swimmer := FindSwimmerByID(swimmerId, uc.DB)

	events, err := findSwimmerMissingBestTimes(swimmer, course, uc.DB)
	if err != nil {
		log.Printf("swimming.EventsResource: %v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	evts, err := json.Marshal(events)
	if err != nil {
		log.Printf("swimming.EventsResource: %v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	_, err = res.Write(evts)
	if err != nil {
		log.Printf("swimming.EventsResource: %v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

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
			"Title":             utils.Title,
			"FormatMiliseconds": utils.FormatMiliseconds,
		}).Parse(
		`{{range .BestTimes}}
			<tr>
				<td>{{.Course | Title}}</td>
				<td><a href="/profile/swimmers/{{$.Swimmer.ID}}/besttimes/{{.ID}}/">{{.Event.Distance}} {{.Event.Style.Stroke | Title}}</a></td>
				<td>{{.BestTime | FormatMiliseconds}}</td>
				<td>{{.Points}}</td>
			</tr>
		{{else}}
			<tr>
				<td colspan="4">No best time listed yet.</td>
			</tr>
		{{end}}`)
	if err != nil {
		log.Printf("swimming.BestTimesPartial: %v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	html.Execute(res, data)
}

func (uc *Controller) SwimmerBestTimeDelete(res http.ResponseWriter, req *http.Request, sessionData *storage.SessionData) {
	id := req.URL.Query().Get(":id")
	swimmerId, _ := strconv.ParseInt(id, 10, 64)
	swimmer := FindSwimmerByID(swimmerId, uc.DB)

	id = req.URL.Query().Get(":bestId")
	bestId, _ := strconv.ParseInt(id, 10, 64)
	bestTime := findSwimmerBestTime(swimmer, bestId, uc.DB)

	if bestTime != nil {
		err := deleteSwimmerBestTime(bestTime, uc.DB)
		if err != nil {
			log.Printf("Error deleting the best time: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
		}
	}

	res.WriteHeader(http.StatusOK)
}

func (uc *Controller) SwimmerDelete(res http.ResponseWriter, req *http.Request, sessionData *storage.SessionData) {
	id := req.URL.Query().Get(":id")
	swimmerId, _ := strconv.ParseInt(id, 10, 64)
	swimmer := FindSwimmerByID(swimmerId, uc.DB)

	parent := FindUserAccountByEmail(sessionData.Email, uc.DB)
	parentSwimmer := findParentSwimmer(parent, swimmer, uc.DB)
	if parentSwimmer == nil {
		http.Error(res, "Permission denied", http.StatusUnauthorized)
		return
	}

	err := deleteSwimmer(swimmer, uc.DB)
	if err != nil {
		log.Printf("Error deleting the swimmer: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
	}

	res.WriteHeader(http.StatusOK)
}

func (uc *Controller) AcceptParentLink(res http.ResponseWriter, req *http.Request, sessionData *storage.SessionData) {
	id := req.URL.Query().Get(":id")
	swimmerId, _ := strconv.ParseInt(id, 10, 64)
	swimmer := FindSwimmerByID(swimmerId, uc.DB)

	id = req.URL.Query().Get(":linkId")
	linkId, _ := strconv.ParseInt(id, 10, 64)

	err := approvalParentLink(swimmer, linkId, ParentSwimmerApprovalAccepted, uc.DB)
	if err != nil {
		log.Printf("Error accepting the parent link: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
	}

	res.WriteHeader(http.StatusOK)
}

func (uc *Controller) DismissParentLink(res http.ResponseWriter, req *http.Request, sessionData *storage.SessionData) {
	id := req.URL.Query().Get(":id")
	swimmerId, _ := strconv.ParseInt(id, 10, 64)
	swimmer := FindSwimmerByID(swimmerId, uc.DB)

	id = req.URL.Query().Get(":linkId")
	linkId, _ := strconv.ParseInt(id, 10, 64)

	err := approvalParentLink(swimmer, linkId, ParentSwimmerApprovalDismissed, uc.DB)
	if err != nil {
		log.Printf("Error dismissing the parent link: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
	}

	res.WriteHeader(http.StatusOK)
}
