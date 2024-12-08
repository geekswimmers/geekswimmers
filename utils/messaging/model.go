package messaging

type EmailMessageSent struct {
	ID        int64
	Recipient string
	Username  string
	Subject   string
	Body      string
}
