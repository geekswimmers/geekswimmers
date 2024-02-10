package content

import "time"

type Article struct {
	Reference string
	Title     string
	Published time.Time
	Content   string
}
