package content

import (
	"time"
)

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

type Quote struct {
	Sequence int64
	Quote    string
	Author   string
}

type ServiceUpdate struct {
	Title     string
	Content   string
	Published time.Time
}
