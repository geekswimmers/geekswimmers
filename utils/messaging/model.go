package messaging

type EmailMessageSent struct {
	ID        int64
	Recipient string
	Subject   string
	Body      string
}
