package web

// This context is shared globally within the application. Do not put any session-specific data here.
type BaseTemplateContext struct {
	FeedbackForm              string
	MonitoringGoogleAnalytics string
}
