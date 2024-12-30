package content

import (
	"geekswimmers/storage"
	"geekswimmers/utils"
)

type articleViewData struct {
	Article          *Article
	OtherArticles    []*Article
	BaseTemplateData *utils.BaseTemplateData
	SessionData      *storage.SessionData
}
