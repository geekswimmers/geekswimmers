package messaging

type EmailMessageSent struct {
	ID        int64
	Recipient string
	Subject   string
	Body      string
}

type EmailData struct {
	ServerUrl    string
	Username     string
	CurrentEmail string
	NewEmail     string
	Confirmation string
	Name         string
	Recipe       string
	Step         string
	Summary      string
	Message      string
	ID           int64
}
