package user

import (
	"context"
	"fmt"
	"geekswimmers/storage"
	"log"
	"strings"

	"github.com/jackc/pgx/v5"
)

func InsertUserAccount(userAccount *UserAccount, db storage.Database) (int64, error) {
	var lastInsertId int64

	if userAccount.Role == "ATHLETE" {
		stmt := `insert into user_account (email, first_name, last_name, human_score, confirmation, access_role, birth_date, gender)
		values ($1, $2, $3, $4, $5, $6, $7, $8) returning id`

		err := db.QueryRow(context.Background(), stmt,
			userAccount.CleanEmail(),
			userAccount.FirstName,
			userAccount.LastName,
			userAccount.HumanScore,
			userAccount.Confirmation,
			userAccount.Role,
			userAccount.BirthDate.Time,
			userAccount.Gender.String).Scan(&lastInsertId)
		if err != nil {
			return 0, fmt.Errorf("user.InsertUserAccount(%v): %v", userAccount.Email, err)
		}
	} else {
		stmt := `insert into user_account (email, first_name, last_name, human_score, confirmation, access_role) 
		values ($1, $2, $3, $4, $5, $6) returning id`

		err := db.QueryRow(context.Background(), stmt,
			userAccount.CleanEmail(),
			userAccount.FirstName,
			userAccount.LastName,
			userAccount.HumanScore,
			userAccount.Confirmation,
			userAccount.Role).Scan(&lastInsertId)
		if err != nil {
			return 0, fmt.Errorf("user.InsertUserAccount(%v): %v", userAccount.Email, err)
		}
	}

	return lastInsertId, nil
}

func InsertSignInAttempt(signInAttempt SignInAttempt, db storage.Database) error {
	stmt := `insert into sign_in_attempt (identifier, human_score, status, ip_address, failed_match) 
             values ($1, $2, $3, $4, $5)`

	_, err := db.Exec(context.Background(), stmt,
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

func UpdateUserAccount(userAccount *UserAccount, db storage.Database) error {
	stmt := `update user_account set confirmation = $1, 
                                     modified = current_timestamp,
                                     first_name = $2,
                       	             last_name = $3,
                       	             promotional_msg = $4
             where id = $5`

	_, err := db.Exec(context.Background(), stmt, userAccount.Confirmation, userAccount.FirstName, userAccount.LastName, userAccount.PromotionalMsg, userAccount.ID)
	if err != nil {
		return fmt.Errorf("user.UpdateUserAccount(%v, %v, %v, %v, %v): %v", userAccount.Confirmation, userAccount.FirstName,
			userAccount.LastName, userAccount.PromotionalMsg, userAccount.ID, err)
	}

	return nil
}

func setUserAccountNewPassword(userAccount *UserAccount, db storage.Database) error {
	stmt := `update user_account set password = $1, 
	                                 confirmation = null, 
									 modified = current_timestamp
             where email = $2 and confirmation = $3`

	_, err := db.Exec(context.Background(), stmt, userAccount.Password, userAccount.Email, userAccount.Confirmation)
	if err != nil {
		return fmt.Errorf("user.setUserAccountNewPassword(%v, %v): %v", userAccount.Email, userAccount.Confirmation, err)
	}

	return nil
}

func SetUserAccountNewEmail(userAccount *UserAccount, newEmail string, db storage.Database) error {
	stmt := `update user_account set email = $1,
									 confirmation = null, 
									 modified = current_timestamp
             where email = $2 and confirmation = $3`

	currentEmail := userAccount.Email
	userAccount.Email = newEmail

	_, err := db.Exec(context.Background(), stmt, userAccount.CleanEmail(), currentEmail, userAccount.Confirmation)
	if err != nil {
		return fmt.Errorf("user.SetUserAccountNewEmail(%v, %v, %v): %v", newEmail, userAccount.Email, userAccount.Confirmation, err)
	}

	return nil
}

func StartUserAccountSignOffPeriod(userAccount *UserAccount, feedback string, db storage.Database) error {
	stmt := `update user_account set modified = current_timestamp,
                                     sign_off = current_timestamp,
                                     sign_off_feedback = $1
             where id = $2`

	_, err := db.Exec(context.Background(), stmt, feedback, userAccount.ID)
	if err != nil {
		return fmt.Errorf("user.StartUserAccountSignOffPeriod(%v, %v): %v", userAccount.ID, feedback, err)
	}

	return nil
}

func ResetUserAccountSignOffPeriod(userAccount *UserAccount, db storage.Database) error {
	stmt := `update user_account set modified = current_timestamp,
                                     sign_off = null,
                                     sign_off_feedback = null
             where id = $1`

	_, err := db.Exec(context.Background(), stmt, userAccount.ID)
	if err != nil {
		return fmt.Errorf("user.ResetUserAccountSignOffPeriod(%v): %v", userAccount.ID, err)
	}

	return nil
}

func FindUserAccountByEmail(email string, db storage.Database) *UserAccount {
	stmt := `select id, email, first_name, last_name, access_role, password, sign_off, promotional_msg, birth_date, gender
             from user_account where email = $1`

	email = strings.ToLower(email)
	email = strings.TrimSpace(email)

	row := db.QueryRow(context.Background(), stmt, email)

	userAccount := &UserAccount{}
	err := row.Scan(&userAccount.ID, &userAccount.Email,
		&userAccount.FirstName, &userAccount.LastName, &userAccount.Role, &userAccount.Password,
		&userAccount.SignOff, &userAccount.PromotionalMsg, &userAccount.BirthDate, &userAccount.Gender)
	if err != nil {
		log.Printf("user.FindUserAccountByEmail(%v) : %v", email, err)
		return nil
	}

	return userAccount
}

func FindUserAccountByConfirmation(confirmation, email string, db storage.Database) *UserAccount {
	stmt := `select id, email, first_name, last_name, access_role 
             from user_account where confirmation = $1`

	var row pgx.Row

	if len(email) > 0 {
		stmt = fmt.Sprintf("%s and email = $2", stmt)
		row = db.QueryRow(context.Background(), stmt, confirmation, email)
	} else {
		row = db.QueryRow(context.Background(), stmt, confirmation)
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

func UserAccountExists(db storage.Database) bool {
	stmt := `select count(id) from user_account`

	row := db.QueryRow(context.Background(), stmt)

	var count int
	err := row.Scan(&count)
	if err != nil {
		log.Printf("user.UserAccountExists(): %v", err)
		return false
	}

	return count > 0
}

func TooManySignInAttempts(ipAddress string, db storage.Database) bool {
	stmt := `select count(id) 
			 from sign_in_attempt
             where status = $1 and ip_address = $2  and created >= (current_timestamp - interval '1 HOURS') 
             limit 10`

	row := db.QueryRow(context.Background(), stmt, StatusFailed, ipAddress)

	var numFailedAttempts int
	err := row.Scan(&numFailedAttempts)
	if err != nil {
		log.Printf("user.TooManySignInAttempts(%v, %v): %v", StatusFailed, ipAddress, err)
	}
	log.Printf("Number of failed attempts: %v", numFailedAttempts)

	return numFailedAttempts > 5
}
