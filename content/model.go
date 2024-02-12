package content

import "time"

type Article struct {
	Reference   string
	Title       string
	Abstract    string
	Highlighted bool
	Published   time.Time
	Content     string
}
