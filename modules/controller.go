package modules

import (
	"geekswimmers/storage"
	"geekswimmers/utils"
)

func InitializeRequestContext(sessionData *storage.SessionData, baseTemplateData *utils.BaseTemplateData) map[string]any {
	return map[string]any{
		"BaseTemplateData": baseTemplateData,
		"SessionData":      sessionData,
	}
}
