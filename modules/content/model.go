package content

import (
	"fmt"
	"strings"
	"time"
)

type Article struct {
	Reference      string
	Title          string
	SubTitle       string
	Highlighted    bool
	Published      time.Time
	ContentFile    string
	Image          string
	ImageCopyright string

	// Transient fields
	abstract string
	content  string
}

func (a *Article) Abstract() string {
	if a.abstract != "" {
		return a.abstract
	}

	markdownContent, err := LoadMarkdownContent(fmt.Sprintf("web/content/%s", a.ContentFile))
	if err != nil {
		return ""
	}

	lines := strings.Split(markdownContent, "\n")
	if len(lines) == 0 {
		return ""
	}

	var abstract string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "<!-- more -->" {
			break
		}

		abstract += "\n" + line
	}

	a.abstract = abstract
	return strings.TrimSpace(abstract)
}

func (a *Article) Content() string {
	if a.content != "" {
		return a.content
	}

	markdownContent, err := LoadMarkdownContent(fmt.Sprintf("web/content/%s", a.ContentFile))
	if err != nil {
		return ""
	}

	lines := strings.Split(markdownContent, "\n")
	if len(lines) == 0 {
		return ""
	}

	var abstract, content string
	contentFound := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "<!-- more -->" && !contentFound {
			contentFound = true
			continue
		}

		if contentFound {
			content += "\n" + line
			continue
		}

		abstract += "\n" + line
	}

	a.content = content
	a.abstract = abstract

	return strings.TrimSpace(content)
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
