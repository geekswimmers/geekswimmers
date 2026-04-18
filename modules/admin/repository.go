package admin

import (
	"context"
	"fmt"
	"geekswimmers/modules/times"
	"geekswimmers/storage"
)

func findMeetEvent(meetID, eventID int64, gender string, db storage.Database) (*times.MeetEvent, error) {
	stm := `select id from meet_event where meet = $1 and event = $2 and gender = $3`
	row := db.QueryRow(context.Background(), stm, meetID, eventID, gender)

	meetEvent := &times.MeetEvent{
		MeetID: meetID,
		Gender: gender,
	}
	err := row.Scan(&meetEvent.ID)
	if err != nil && err.Error() == storage.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("findMeetEvent: %v", err)
	}
	meetEvent.Event.ID = eventID
	return meetEvent, nil
}

func insertMeetEvent(event *times.MeetEvent, db storage.Database) error {
	stm := `insert into meet_event (meet, event, gender) values ($1, $2, $3) returning id`
	row := db.QueryRow(context.Background(), stm, event.MeetID, event.Event.ID, event.Gender)
	err := row.Scan(&event.ID)
	if err != nil {
		return fmt.Errorf("insertMeetEvent: %v", err)
	}
	return nil
}

func insertMeetResult(result *times.MeetResult, db storage.Database) error {
	stm := `insert into meet_result (meet_event, swimmer, team, result_time, dq)
	        values ($1, $2, $3, $4, $5)
	        on conflict (meet_event, swimmer) do nothing`
	_, err := db.Exec(context.Background(), stm,
		result.MeetEvent.ID,
		result.Swimmer.ID,
		result.Team.ID.Int64,
		result.ResultTime,
		result.DQ)
	if err != nil {
		return fmt.Errorf("insertMeetResult: %v", err)
	}
	return nil
}
