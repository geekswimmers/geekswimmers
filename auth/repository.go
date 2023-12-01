package auth

import (
	"database/sql"
	"fmt"
	"geekswimmers/storage"
	"log"
	"strings"
)

func InsertUserAccount(userAccount *UserAccount, db storage.Database) (int64, error) {
	stmt := `insert into user_account (email, username, human_score, confirmation) 
             values ($1, $2, $3, $4) returning id`

	var lastInsertId int64
	err := db.QueryRow(stmt,
		userAccount.Email,
		userAccount.Username,
		userAccount.HumanScore,
		userAccount.Confirmation).Scan(&lastInsertId)
	if err != nil {
		return 0, fmt.Errorf("auth.InsertUserAccount(%v, %v, %v, %v): %v", userAccount.Email, userAccount.Username,
			userAccount.HumanScore, userAccount.Confirmation, err)
	}

	return lastInsertId, nil
}

func InsertSignInAttempt(signInAttempt SignInAttempt, db storage.Database) error {
	stmt := `insert into sign_in_attempt (identifier, human_score, status, ip_address, failed_match) 
             values ($1, $2, $3, $4, $5)`

	_, err := db.Exec(stmt,
		signInAttempt.Identifier,
		signInAttempt.HumanScore,
		signInAttempt.Status,
		signInAttempt.IPAddress,
		signInAttempt.FailedMatch)
	if err != nil {
		return fmt.Errorf("auth.InsertSignInAttempt(%v, %v, %v, %v, %v): %v", signInAttempt.Identifier,
			signInAttempt.HumanScore, signInAttempt.Status, signInAttempt.IPAddress, signInAttempt.FailedMatch, err)
	}
	return nil
}

func UpdateUserAccount(userAccount *UserAccount, db storage.Database) error {
	stmt := `update user_account 
	         set confirmation = $1, 
                 modified = current_timestamp,
                 first_name = $2,
                 last_name = $3,
                 notification_promo = $4
             where id = $5`

	_, err := db.Exec(stmt, userAccount.Confirmation, userAccount.FirstName, userAccount.LastName, userAccount.NotificationPromo, userAccount.ID)
	if err != nil {
		return fmt.Errorf("auth.UpdateUserAccount(%v, %v, %v, %v, %v): %v", userAccount.Confirmation, userAccount.FirstName,
			userAccount.LastName, userAccount.NotificationPromo, userAccount.ID, err)
	}

	return nil
}

func setUserAccountNewPassword(userAccount *UserAccount, db storage.Database) error {
	stmt := `update user_account set password = $1, confirmation = null 
             where email = $2 and confirmation = $3`

	_, err := db.Exec(stmt, userAccount.Password, userAccount.Email, userAccount.Confirmation)
	if err != nil {
		return fmt.Errorf("auth.setUserAccountNewPassword(%v, %v): %v", userAccount.Email, userAccount.Confirmation, err)
	}

	return nil
}

func SetUserAccountNewEmail(userAccount *UserAccount, newEmail string, db storage.Database) error {
	stmt := `update user_account set email = $1, confirmation = null 
             where email = $2 and confirmation = $3`

	_, err := db.Exec(stmt, newEmail, userAccount.Email, userAccount.Confirmation)
	if err != nil {
		return fmt.Errorf("auth.SetUserAccountNewEmail(%v, %v, %v): %v", newEmail, userAccount.Email, userAccount.Confirmation, err)
	}

	return nil
}

func StartUserAccountSignOffPeriod(userAccount *UserAccount, feedback string, db storage.Database) error {
	stmt := `update user_account set modified = current_timestamp,
                                     sign_off = current_timestamp,
                                     sign_off_feedback = $1
             where id = $2`

	_, err := db.Exec(stmt, feedback, userAccount.ID)
	if err != nil {
		return fmt.Errorf("auth.StartUserAccountSignOffPeriod(%v, %v): %v", userAccount.ID, feedback, err)
	}

	return nil
}

func ResetUserAccountSignOffPeriod(userAccount *UserAccount, db storage.Database) error {
	stmt := `update user_account set modified = current_timestamp,
                                     sign_off = null,
                                     sign_off_feedback = null
             where id = $1`

	_, err := db.Exec(stmt, userAccount.ID)
	if err != nil {
		return fmt.Errorf("auth.ResetUserAccountSignOffPeriod(%v): %v", userAccount.ID, err)
	}

	return nil
}

func FindUserAccountByEmail(email string, db storage.Database) *UserAccount {
	stmt := `select id, email, username, first_name, last_name, password, sign_off, notification_promo, access_role
             from user_account where email = $1`

	email = strings.ToLower(email)
	email = strings.TrimSpace(email)

	row := db.QueryRow(stmt, email)

	userAccount := &UserAccount{}
	err := row.Scan(&userAccount.ID, &userAccount.Email, &userAccount.Username,
		&userAccount.FirstName, &userAccount.LastName, &userAccount.Password,
		&userAccount.SignOff, &userAccount.NotificationPromo, &userAccount.Role)
	if err != nil {
		log.Printf("auth.FindUserAccountByEmail(%v) : %v", email, err)
		return nil
	}

	return userAccount
}

func FindUserAccountByUsername(username string, db storage.Database) *UserAccount {
	stmt := `select ua.id, ua.email, ua.username, ua.first_name, ua.last_name, ua.password, ua.sign_off, ua.notification_promo, ua.access_role
			 from user_account ua
			 where ua.username = $1`

	row := db.QueryRow(stmt, username)

	userAccount := &UserAccount{}
	err := row.Scan(&userAccount.ID, &userAccount.Email, &userAccount.Username, &userAccount.FirstName,
		&userAccount.LastName, &userAccount.Password, &userAccount.SignOff, &userAccount.NotificationPromo, &userAccount.Role)
	if err != nil {
		log.Printf("auth.FindUserAccountByUsername(%v): %v", username, err)
		return nil
	}

	return userAccount
}

func FindUserAccountByConfirmation(confirmation, email string, db storage.Database) *UserAccount {
	stmt := `select id, email, username, first_name, last_name, confirmation 
             from user_account where confirmation = $1`

	var row *sql.Row

	if len(email) > 0 {
		stmt = fmt.Sprintf("%s and email = $2", stmt)
		row = db.QueryRow(stmt, confirmation, email)
	} else {
		row = db.QueryRow(stmt, confirmation)
	}

	userAccount := &UserAccount{}
	err := row.Scan(&userAccount.ID, &userAccount.Email, &userAccount.Username, &userAccount.FirstName,
		&userAccount.LastName, &userAccount.Confirmation)
	if err != nil {
		log.Printf("auth.FindUserAccountByConfirmation(%v, %v): %v", confirmation, email, err)
		return nil
	}
	return userAccount
}

func TooManySignInAttempts(ipAddress string, db storage.Database) bool {
	stmt := `select count(id) 
			 from sign_in_attempt
             where status = $1 and ip_address = $2  and created >= (current_timestamp - interval '1 HOURS') 
             limit 5`

	row := db.QueryRow(stmt, StatusFailed, ipAddress)

	var numFailedAttempts int
	err := row.Scan(&numFailedAttempts)
	if err != nil {
		log.Printf("auth.TooManySignInAttempts(%v, %v): %v", StatusFailed, ipAddress, err)
	}
	log.Printf("Number of failed attempts: %v", numFailedAttempts)

	return numFailedAttempts > 10
}
