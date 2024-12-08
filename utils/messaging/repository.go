package messaging

import (
	"context"
	"fmt"
	"geekswimmers/storage"
)

func (emailMessageSent *EmailMessageSent) Insert(db storage.Database) error {
	stmt := `insert into email_message_sent (recipient, username, subject, body) values ($1, $2, $3, $4)`

	_, err := db.Exec(context.Background(), stmt,
		emailMessageSent.Recipient,
		emailMessageSent.Username,
		emailMessageSent.Subject,
		emailMessageSent.Body)

	if err != nil {
		return fmt.Errorf("utils.EmailMessageSent.Insert(%v, %v, %v, %v): %v",
			emailMessageSent.Recipient,
			emailMessageSent.Username,
			emailMessageSent.Subject,
			emailMessageSent.Body, err)
	}
	return nil
}
