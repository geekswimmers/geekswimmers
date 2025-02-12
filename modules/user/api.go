package user

import (
	"encoding/json"
	"geekswimmers/storage"
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
	}

	evts, err := json.Marshal(events)
	if err != nil {
		log.Printf("swimming.EventsResource: %v", err)
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	_, err = res.Write(evts)
	if err != nil {
		log.Printf("swimming.EventsResource: %v", err)
	}
}

func (uc *Controller) SwimmerBestTimeDelete(res http.ResponseWriter, req *http.Request) {
	sessionData := storage.NewSessionData(req)
	if !sessionData.IsAuthenticated() {
		http.Redirect(res, req, "/auth/signin/", http.StatusSeeOther)
		return
	}

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
