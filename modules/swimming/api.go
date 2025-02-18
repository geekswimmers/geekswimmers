package swimming

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

func (sc *Controller) ClubResource(res http.ResponseWriter, req *http.Request) {
	j := req.URL.Query().Get("jurisdiction")
	jurisdictionId, _ := strconv.ParseInt(j, 10, 64)
	jurisdiction := Jurisdiction{
		ID: sql.NullInt64{
			Int64: jurisdictionId,
		},
	}

	clubs, err := FindClubsByJurisdiction(jurisdiction, sc.DB)
	if err != nil {
		log.Printf("swimming.ClubResource: %v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	clubsJson, err := json.Marshal(clubs)
	if err != nil {
		log.Printf("swimming.ClubResource: %v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	_, err = res.Write(clubsJson)
	if err != nil {
		log.Printf("swimming.ClubResource: %v", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}
