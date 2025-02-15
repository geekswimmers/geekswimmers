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

func (uc *Controller) SwimmerBestTimeDelete(res http.ResponseWriter, req *http.Request) {
	sessionData := storage.NewSessionData(req)
	if !sessionData.IsAuthenticated() {
		http.Error(res, "Permission denied", http.StatusUnauthorized)
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

func (uc *Controller) SwimmerDelete(res http.ResponseWriter, req *http.Request) {
	sessionData := storage.NewSessionData(req)
	if !sessionData.IsAuthenticated() {
		http.Error(res, "Permission denied", http.StatusUnauthorized)
		return
	}

	id := req.URL.Query().Get(":id")
	swimmerId, _ := strconv.ParseInt(id, 10, 64)
	swimmer := FindSwimmerByID(swimmerId, uc.DB)

	err := deleteSwimmer(swimmer, uc.DB)
	if err != nil {
		log.Printf("Error deleting the swimmer: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
	}

	res.WriteHeader(http.StatusOK)
}

func (uc *Controller) AcceptParentLink(res http.ResponseWriter, req *http.Request) {
	sessionData := storage.NewSessionData(req)
	if !sessionData.IsAuthenticated() {
		http.Error(res, "Permission denied", http.StatusUnauthorized)
		return
	}

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

func (uc *Controller) DismissParentLink(res http.ResponseWriter, req *http.Request) {
	sessionData := storage.NewSessionData(req)
	if !sessionData.IsAuthenticated() {
		http.Error(res, "Permission denied", http.StatusUnauthorized)
		return
	}

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
