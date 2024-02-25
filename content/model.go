package content

import "time"

type Article struct {
	Reference      string
	Title          string
	SubTitle       string
	Abstract       string
	Highlighted    bool
	Published      time.Time
	Content        string
	Image          string
	ImageCopyright string
}
