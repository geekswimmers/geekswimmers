package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"geekswimmers/modules/swimming"
	"geekswimmers/storage"
	"log"
	"strings"

	"github.com/jackc/pgx/v5"
)

func InsertUserAccount(userAccount *UserAccount, db storage.Database) (int64, error) {
	var lastInsertId int64

	stm := `insert into user_account (email, first_name, last_name, human_score, confirmation, access_role) 
			values ($1, $2, $3, $4, $5, $6) returning id`

	err := db.QueryRow(context.Background(), stm,
		userAccount.CleanEmail(),
		userAccount.FirstName,
		userAccount.LastName,
		userAccount.HumanScore,
		userAccount.Confirmation,
		userAccount.Role).Scan(&lastInsertId)
	if err != nil {
		return 0, fmt.Errorf("user.InsertUserAccount(%v): %v", userAccount.Email, err)
	}

	return lastInsertId, nil
}

func InsertSignInAttempt(signInAttempt SignInAttempt, db storage.Database) error {
	stm := `insert into sign_in_attempt (identifier, human_score, status, ip_address, failed_match) 
             values ($1, $2, $3, $4, $5)`

	_, err := db.Exec(context.Background(), stm,
		signInAttempt.Identifier,
		signInAttempt.HumanScore,
		signInAttempt.Status,
		signInAttempt.IPAddress,
		signInAttempt.FailedMatch)
	if err != nil {
		return fmt.Errorf("user.InsertSignInAttempt(%v, %v, %v, %v, %v): %v", signInAttempt.Identifier,
			signInAttempt.HumanScore, signInAttempt.Status, signInAttempt.IPAddress, signInAttempt.FailedMatch, err)
	}
	return nil
}

func InsertSwimmer(swimmer *UserSwimmer, db storage.Database) (int64, error) {
	var lastInsertId int64

	stm := `insert into swimmer (first_name, last_name, birth_date, gender, user_account)
			values ($1, $2, $3, $4, $5) returning id`

	var userAccountId sql.NullInt64
	if swimmer.UserAccount != nil {
		userAccountId = sql.NullInt64{
			Int64: swimmer.UserAccount.ID,
			Valid: true,
		}
	} else {
		userAccountId = sql.NullInt64{}
	}

	err := db.QueryRow(context.Background(), stm,
		swimmer.Swimmer.FirstName,
		swimmer.Swimmer.LastName,
		swimmer.Swimmer.BirthDate.Time,
		swimmer.Swimmer.Gender.String,
		userAccountId).Scan(&lastInsertId)
	if err != nil {
		return 0, fmt.Errorf("user.InsertSwimmer(%v %v): %v", swimmer.Swimmer.FirstName, swimmer.Swimmer.LastName, err)
	}

	return lastInsertId, nil
}

func insertSwimmerBestTime(bestTime *SwimmerBestTime, db storage.Database) (int64, error) {
	var lastInsertId int64

	stm := `insert into swimmer_best_time (swimmer, event, course, best_time, updated) 
			values ($1, $2, $3, $4, current_timestamp) returning id`

	err := db.QueryRow(context.Background(), stm,
		bestTime.Swimmer.ID,
		bestTime.Event.ID,
		bestTime.Course,
		bestTime.BestTime).Scan(&lastInsertId)
	if err != nil {
		return 0, fmt.Errorf("user.InsertSwimmerBestTime(%v, %v, %v, %v): %v", bestTime.Swimmer.ID,
			bestTime.Event.ID, bestTime.Course, bestTime.BestTime, err)
	}

	return lastInsertId, nil
}

func updateProfile(userAccount *UserAccount, db storage.Database) error {
	stm := `update user_account 
			set first_name = $1,
                last_name = $2,
                email = $3
            where id = $4`

	_, err := db.Exec(context.Background(), stm, userAccount.FirstName, userAccount.LastName, userAccount.Email, userAccount.ID)
	if err != nil {
		return fmt.Errorf("user.updateProfile(%v): %v", userAccount.ID, err)
	}

	return nil
}

func updateSwimmerProfile(userAccount *UserAccount, swimmer *UserSwimmer, db storage.Database) error {
	err := updateProfile(userAccount, db)
	if err != nil {
		return fmt.Errorf("user.updateSwimmerProfile(%v): %v", userAccount.Email, err)
	}

	stm := `update swimmer 
			set first_name = $1,
                last_name = $2,
                birth_date = $3, 
                gender = $4
             where id = $5`

	_, err = db.Exec(context.Background(), stm, swimmer.Swimmer.FirstName, swimmer.Swimmer.LastName,
		swimmer.Swimmer.BirthDate.Time, swimmer.Swimmer.Gender.String, swimmer.ID)
	if err != nil {
		return fmt.Errorf("user.updateSwimmerProfile(%v): %v", swimmer.ID, err)
	}

	return nil
}

func updateUserAccount(userAccount *UserAccount, db storage.Database) error {
	stm := `update user_account set confirmation = $1, 
                                     modified = current_timestamp,
                                     first_name = $2,
                       	             last_name = $3,
                       	             promotional_msg = $4
             where id = $5`

	_, err := db.Exec(context.Background(), stm, userAccount.Confirmation, userAccount.FirstName, userAccount.LastName, userAccount.PromotionalMsg, userAccount.ID)
	if err != nil {
		return fmt.Errorf("user.UpdateUserAccount(%v, %v, %v, %v, %v): %v", userAccount.Confirmation, userAccount.FirstName,
			userAccount.LastName, userAccount.PromotionalMsg, userAccount.ID, err)
	}

	return nil
}

func setUserAccountNewPassword(userAccount *UserAccount, db storage.Database) error {
	stm := `update user_account set password = $1, 
	                                 confirmation = null, 
									 modified = current_timestamp
             where email = $2 and confirmation = $3`

	_, err := db.Exec(context.Background(), stm, userAccount.Password, userAccount.Email, userAccount.Confirmation)
	if err != nil {
		return fmt.Errorf("user.setUserAccountNewPassword(%v, %v): %v", userAccount.Email, userAccount.Confirmation, err)
	}

	return nil
}

func SetUserAccountNewEmail(userAccount *UserAccount, newEmail string, db storage.Database) error {
	stm := `update user_account set email = $1,
									 confirmation = null, 
									 modified = current_timestamp
             where email = $2 and confirmation = $3`

	currentEmail := userAccount.Email
	userAccount.Email = newEmail

	_, err := db.Exec(context.Background(), stm, userAccount.CleanEmail(), currentEmail, userAccount.Confirmation)
	if err != nil {
		return fmt.Errorf("user.SetUserAccountNewEmail(%v, %v, %v): %v", newEmail, userAccount.Email, userAccount.Confirmation, err)
	}

	return nil
}

func StartUserAccountSignOffPeriod(userAccount *UserAccount, feedback string, db storage.Database) error {
	stm := `update user_account set modified = current_timestamp,
                                     sign_off = current_timestamp,
                                     sign_off_feedback = $1
             where id = $2`

	_, err := db.Exec(context.Background(), stm, feedback, userAccount.ID)
	if err != nil {
		return fmt.Errorf("user.StartUserAccountSignOffPeriod(%v, %v): %v", userAccount.ID, feedback, err)
	}

	return nil
}

func ResetUserAccountSignOffPeriod(userAccount *UserAccount, db storage.Database) error {
	stm := `update user_account set modified = current_timestamp,
                                     sign_off = null,
                                     sign_off_feedback = null
             where id = $1`

	_, err := db.Exec(context.Background(), stm, userAccount.ID)
	if err != nil {
		return fmt.Errorf("user.ResetUserAccountSignOffPeriod(%v): %v", userAccount.ID, err)
	}

	return nil
}

func FindUserAccountByEmail(email string, db storage.Database) *UserAccount {
	stm := `select id, email, first_name, last_name, access_role, password, sign_off, promotional_msg
             from user_account where email = $1`

	email = strings.ToLower(email)
	email = strings.TrimSpace(email)

	row := db.QueryRow(context.Background(), stm, email)

	userAccount := &UserAccount{}
	err := row.Scan(&userAccount.ID, &userAccount.Email,
		&userAccount.FirstName, &userAccount.LastName, &userAccount.Role, &userAccount.Password,
		&userAccount.SignOff, &userAccount.PromotionalMsg)
	if err != nil {
		log.Printf("user.FindUserAccountByEmail(%v) : %v", email, err)
		return nil
	}

	return userAccount
}

func FindUserAccountByConfirmation(confirmation, email string, db storage.Database) *UserAccount {
	stm := `select id, email, first_name, last_name, access_role 
             from user_account where confirmation = $1`

	var row pgx.Row

	if len(email) > 0 {
		stm = fmt.Sprintf("%s and email = $2", stm)
		row = db.QueryRow(context.Background(), stm, confirmation, email)
	} else {
		row = db.QueryRow(context.Background(), stm, confirmation)
	}

	userAccount := &UserAccount{
		Confirmation: &confirmation,
	}
	err := row.Scan(&userAccount.ID, &userAccount.Email, &userAccount.FirstName, &userAccount.LastName, &userAccount.Role)
	if err != nil {
		log.Printf("user.FindUserAccountByConfirmation(%v, %v): %v", confirmation, email, err)
		return nil
	}
	return userAccount
}

func FindSwimmerByUserAccount(userAccount *UserAccount, db storage.Database) *UserSwimmer {
	stm := `select a.id, a.first_name, a.last_name, a.birth_date, a.gender
			from swimmer a
			where a.user_account = $1`

	row := db.QueryRow(context.Background(), stm, userAccount.ID)

	swimmer := &UserSwimmer{
		UserAccount: userAccount,
		Swimmer:     &swimming.Swimmer{},
	}
	err := row.Scan(&swimmer.ID, &swimmer.Swimmer.FirstName, &swimmer.Swimmer.LastName, &swimmer.Swimmer.BirthDate, &swimmer.Swimmer.Gender)
	if err != nil {
		log.Printf("user.FindSwimmerByUserAccount(%v): %v", userAccount.ID, err)
		return nil
	}

	return swimmer
}

func FindSwimmerByID(id int64, db storage.Database) *UserSwimmer {
	stm := `select a.id, a.first_name, a.last_name, a.birth_date, a.gender, a.user_account
			from swimmer a
			where a.id = $1`

	row := db.QueryRow(context.Background(), stm, id)

	swimmer := &UserSwimmer{
		Swimmer: &swimming.Swimmer{},
	}
	err := row.Scan(&swimmer.ID, &swimmer.Swimmer.FirstName, &swimmer.Swimmer.LastName, &swimmer.Swimmer.BirthDate, &swimmer.Swimmer.Gender, &swimmer.UserAccountID)
	if err != nil {
		log.Printf("user.FindSwimmerByID(%v): %v", id, err)
		return nil
	}

	return swimmer
}

func FindSwimmerByEmail(email string, db storage.Database) *UserSwimmer {
	stm := `select s.id, s.first_name, s.last_name, s.birth_date, s.gender 
            from swimmer s
    			join user_account ua on ua.id = s.user_account 
			where ua.email = $1`

	row := db.QueryRow(context.Background(), stm, email)

	swimmer := &UserSwimmer{
		Swimmer: &swimming.Swimmer{},
	}
	err := row.Scan(&swimmer.ID, &swimmer.Swimmer.FirstName, &swimmer.Swimmer.LastName, &swimmer.Swimmer.BirthDate, &swimmer.Swimmer.Gender)
	if err != nil {
		log.Printf("user.FindSwimmerByEmail(%v) : %v", email, err)
		return nil
	}

	return swimmer
}

func findLinkableSwimmerByEmail(email string, parent *UserAccount, db storage.Database) ([]*UserSwimmer, error) {
	// First, it checks if there is an swimmer with a user account
	stm := `select a.id, a.first_name, a.last_name, a.gender, ua.email
			from swimmer a
				join user_account ua on a.user_account = ua.id
			where ua.email = $1
				and a.id not in (select swimmer from parent_swimmer where parent = $2)`

	rows, err := db.Query(context.Background(), stm, email, parent.ID)
	if err != nil {
		return nil, fmt.Errorf("user.findLinkableSwimmerByEmail(%v, %v): %v", email, parent.ID, err)
	}
	defer rows.Close()

	var swimmers []*UserSwimmer
	for rows.Next() {
		swimmer := &UserSwimmer{
			UserAccount: &UserAccount{},
			Swimmer:     &swimming.Swimmer{},
		}
		if err := rows.Scan(&swimmer.ID, &swimmer.Swimmer.FirstName, &swimmer.Swimmer.LastName, &swimmer.Swimmer.Gender, &swimmer.UserAccount.Email); err != nil && err.Error() != storage.ErrNoRows {
			return nil, fmt.Errorf("user.findLinkableSwimmerByEmail(%v, %v): %v", email, parent.ID, err)
		}
		swimmers = append(swimmers, swimmer)
	}

	// If no swimmer is found, it checks if the email belongs to a parent and, if yes, then it
	// returns all the swimmers linked to that parent.
	if len(swimmers) == 0 {
		stm = `select a.id, a.first_name, a.last_name, a.gender, ua.email
			   from parent_swimmer pa
				   join swimmer a on pa.swimmer = a.id
				   join user_account ua on pa.parent = ua.id
			   where ua.email = $1
			   	   and a.id not in (select swimmer from parent_swimmer where parent = $2)`
		rows, err := db.Query(context.Background(), stm, email, parent.ID)
		if err != nil {
			return nil, fmt.Errorf("user.findLinkableSwimmerByEmail(%v, %v): %v", email, parent.ID, err)
		}
		defer rows.Close()

		for rows.Next() {
			swimmer := &UserSwimmer{
				UserAccount: &UserAccount{},
				Swimmer:     &swimming.Swimmer{},
			}
			if err := rows.Scan(&swimmer.ID, &swimmer.Swimmer.FirstName, &swimmer.Swimmer.LastName, &swimmer.Swimmer.Gender, &swimmer.UserAccount.Email); err != nil && err.Error() != storage.ErrNoRows {
				return nil, fmt.Errorf("user.findLinkableSwimmerByEmail(%v, %v): %v", email, parent.ID, err)
			}
			swimmers = append(swimmers, swimmer)
		}
	}

	return swimmers, nil
}

func linkSwimmersToParent(parent *UserAccount, swimmers []*UserSwimmer, db storage.Database) error {
	stm := `insert into parent_swimmer (parent, swimmer, approved) values ($1, $2, $3)`

	for _, swimmer := range swimmers {
		approved := false
		if !swimmer.UserAccountID.Valid {
			approved = true
		}

		_, err := db.Exec(context.Background(), stm, parent.ID, swimmer.ID, approved)
		if err != nil {
			return fmt.Errorf("user.linkSwimmersToParent(%v, %v): %v", parent.ID, swimmer.ID, err)
		}
	}

	return nil
}

func findSwimmersParent(parent *UserAccount, db storage.Database) ([]*UserSwimmer, error) {
	stm := `select a.id, a.first_name, a.last_name, a.birth_date, a.gender, pa.approved
			from swimmer a
			    join parent_swimmer pa on a.id = pa.swimmer
			where pa.parent = $1
			order by a.first_name`

	rows, err := db.Query(context.Background(), stm, parent.ID)
	if err != nil {
		return nil, fmt.Errorf("FindSwimmersParent: %v", err)
	}
	defer rows.Close()

	var swimmers []*UserSwimmer
	for rows.Next() {
		swimmer := &UserSwimmer{
			Swimmer: &swimming.Swimmer{},
		}
		err = rows.Scan(&swimmer.ID, &swimmer.Swimmer.FirstName, &swimmer.Swimmer.LastName, &swimmer.Swimmer.BirthDate,
			&swimmer.Swimmer.Gender, &swimmer.LinkApproved)
		if err != nil && err.Error() != storage.ErrNoRows {
			return nil, fmt.Errorf("FindSwimmersParent: %v", err)
		}
		swimmers = append(swimmers, swimmer)
	}

	return swimmers, nil
}

func findSwimmerBestTimes(swimmer *UserSwimmer, db storage.Database) ([]*SwimmerBestTime, error) {
	stm := `select sbt.id, sbt.course, best_time, updated,
				ss.stroke,
    			se.distance
			from swimmer_best_time sbt 
				left join swim_event se on sbt.event = se.id
				left join swim_style ss on se.style = ss.id
			where sbt.swimmer = $1
			order by sbt.course, ss.sequence`
	rows, err := db.Query(context.Background(), stm, swimmer.ID)
	if err != nil {
		return nil, fmt.Errorf("findSwimmerBestTimes: %v", err)
	}
	defer rows.Close()

	var bestTimes []*SwimmerBestTime
	for rows.Next() {
		bestTime := &SwimmerBestTime{}
		err := rows.Scan(&bestTime.ID, &bestTime.Course, &bestTime.BestTime, &bestTime.Updated, &bestTime.Event.Style.Stroke, &bestTime.Event.Distance)
		if err != nil && err.Error() != storage.ErrNoRows {
			return nil, fmt.Errorf("findSwimmerBestTimes: %v", err)
		}
		bestTimes = append(bestTimes, bestTime)
	}

	return bestTimes, nil
}

func findSwimmerBestTime(swimmer *UserSwimmer, bestId int64, db storage.Database) *SwimmerBestTime {
	stm := `select sbt.id, sbt.course, sbt.best_time, sbt.updated,
				   s.id, s.first_name, s.last_name, s.gender, s.birth_date,
				   se.distance, ss.stroke
			from swimmer_best_time sbt
				left join swimmer s on s.id = sbt.swimmer
				left join swim_event se on se.id = sbt.event
				left join swim_style ss on ss.id = se.style
			where sbt.id = $1 and sbt.swimmer = $2`
	row := db.QueryRow(context.Background(), stm, bestId, swimmer.ID)

	bestTime := &SwimmerBestTime{
		Swimmer: UserSwimmer{
			Swimmer: &swimming.Swimmer{},
		},
	}
	err := row.Scan(&bestTime.ID, &bestTime.Course, &bestTime.BestTime, &bestTime.Updated,
		&bestTime.Swimmer.ID, &bestTime.Swimmer.Swimmer.FirstName, &bestTime.Swimmer.Swimmer.LastName, &bestTime.Swimmer.Swimmer.Gender, &bestTime.Swimmer.Swimmer.BirthDate,
		&bestTime.Event.Distance, &bestTime.Event.Style.Stroke)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		log.Printf("findSwimmerBestTime: %v", err)
		return nil
	}

	return bestTime
}

func findSwimmerMissingBestTimes(swimmer *UserSwimmer, course string, db storage.Database) ([]*swimming.Event, error) {
	stm := `select se.id, ss.stroke, se.distance
			from swim_event se
				join swim_style ss on se.style = ss.id
			where se.id not in (select event from swimmer_best_time where swimmer = $1 and course = $2)
				and se.course = $2
			order by ss.sequence asc, se.distance asc`
	rows, err := db.Query(context.Background(), stm, swimmer.ID, course)
	if err != nil {
		return nil, fmt.Errorf("findSwimmerMissingBestTimes: %v", err)
	}
	defer rows.Close()

	var events []*swimming.Event
	for rows.Next() {
		event := &swimming.Event{}
		err := rows.Scan(&event.ID, &event.Style.Stroke, &event.Distance)
		if err != nil && err.Error() != storage.ErrNoRows {
			return nil, fmt.Errorf("findSwimmerMissingBestTimes: %v", err)
		}
		events = append(events, event)
	}

	return events, nil
}

func userAccountExists(db storage.Database) bool {
	stm := `select count(id) from user_account`

	row := db.QueryRow(context.Background(), stm)

	var count int
	err := row.Scan(&count)
	if err != nil {
		log.Printf("user.UserAccountExists(): %v", err)
		return false
	}

	return count > 0
}

func TooManySignInAttempts(ipAddress string, db storage.Database) bool {
	stm := `select count(id) 
			 from sign_in_attempt
             where status = $1 and ip_address = $2  and created >= (current_timestamp - interval '1 HOURS') 
             limit 10`

	row := db.QueryRow(context.Background(), stm, StatusFailed, ipAddress)

	var numFailedAttempts int
	err := row.Scan(&numFailedAttempts)
	if err != nil {
		log.Printf("user.TooManySignInAttempts(%v, %v): %v", StatusFailed, ipAddress, err)
	}
	log.Printf("Number of failed attempts: %v", numFailedAttempts)

	return numFailedAttempts > 5
}
