// Package config encapsulates viper, so it is not directly referenced all around.
package config

import (
	"fmt"
	"geekswimmers/utils"
	"io/fs"
	"log"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

const (
	// DefaultConfigFile is the file to be used if no custom config file is informed through a flag.
	DefaultConfigFile = "./config.toml"

	// DatabaseURL contains all the parameters required to connect to the database.
	DatabaseURL = "database.url"
	// DatabaseMaxOpenConns limits the number of open connections (in-use + idle) at the same time. The default is unlimited.
	DatabaseMaxOpenConns = "database.maxopenconns"
	// DatabaseConnMaxLifetime defines the time the idle connections will remain active before being discarded. The default is 2.
	DatabaseConnMaxLifetime = "database.connmaxlifetime"

	// EmailServer is the URL address of the email server
	EmailServer = "email.server"
	// EmailPort is the number of the port in the server open for email exchange
	EmailPort = "email.port"
	// EmailTransport indicates the user transport layer security used
	EmailTransport = "email.transport"
	// EmailUsername is the user with the right to send email messages
	EmailUsername = "email.username"
	// EmailPassword is a secret key used to legitimate the identity of the user
	EmailPassword = "email.password"
	// EmailFrom is the address used as from in an email message
	EmailFrom = "email.from"

	// RecaptchaSiteKey reCAPTCHA Site Key used in the sign up page publically.
	RecaptchaSiteKey = "recaptcha.sitekey"
	// RecaptchaSecretKey reCAPTCHA Secret Key used in the backend to verify the authenticity with reCAPTCHA server.
	RecaptchaSecretKey = "recaptcha.secretkey"

	// ServerPort is the HTTP port used to serve the application.
	ServerPort = "server.port"
	// ServerSessionKey is the key used to encrypt the session cookie.
	ServerSessionKey = "server.sessionkey"
	// ServerURL is used to create absolute links in email messages.
	ServerURL = "server.url"

	// MonitoringGoogleAnalytics is the Google Analytics ID.
	MonitoringGoogleAnalytics = "monitoring.googleanalytics"

	// FeedbackForm is the URL of the feedback form.
	FeedbackForm = "miscellaneous.feedbackform"
)

type Config interface {
	GetString(key string) string
	GetInt32(key string) int32
	GetDuration(key string) time.Duration
}

var configuration *viper.Viper

// GetConfiguration returns the configuration set.
// Ex.: config.GetConfiguration().GetString(config.ServerPort)
func GetConfiguration() Config {
	if configuration == nil {
		_, err := InitConfiguration(DefaultConfigFile)
		if err != nil {
			log.Printf("config: %v", err)
			return nil
		}
	}

	return configuration
}

// InitConfiguration loads configuration from an external file.
func InitConfiguration(filePath string) (Config, error) {
	configuration = viper.New()
	configuration.SetConfigFile(filePath)

	bindEnvironmentVariables(configuration)

	configuration.WatchConfig()
	configuration.OnConfigChange(func(e fsnotify.Event) {
		log.Printf("config file changed: %v", e.Name)
	})

	if err := configuration.ReadInConfig(); err != nil {
		if _, ok := err.(*fs.PathError); ok {
			if len(configuration.GetString(DatabaseURL)) == 0 {
				return nil, fmt.Errorf("config file and environment variables not available")
			}
		} else {
			return nil, err
		}
	}

	return configuration, nil
}

/*
The configuration entries that change from an environment to another are replaced by environment variables.
This function maps environment variables to configuration entries, considering only the ones that change.
*/
func bindEnvironmentVariables(viperConfig *viper.Viper) {
	viperConfig.AutomaticEnv()
	viperConfig.SetEnvPrefix("geekswimmers")
	utils.LogError(viperConfig.BindEnv(DatabaseURL, "DATABASE_URL"), "DATABASE_URL not available")
	utils.LogError(viperConfig.BindEnv(DatabaseMaxOpenConns, "DATABASE_MAXOPENCONNS"), "DATABASE_MAXOPENCONNS not available")
	utils.LogError(viperConfig.BindEnv(DatabaseConnMaxLifetime, "DATABASE_CONNMAXLIFETIME"), "DATABASE_CONNMAXLIFETIME not available")

	utils.LogError(viperConfig.BindEnv(EmailServer, "EMAIL_SERVER"), "EMAIL_SERVER not available")
	utils.LogError(viperConfig.BindEnv(EmailPort, "EMAIL_PORT"), "EMAIL_PORT not available")
	utils.LogError(viperConfig.BindEnv(EmailTransport, "EMAIL_TRANSPORT"), "EMAIL_TRANSPORT not available")
	utils.LogError(viperConfig.BindEnv(EmailUsername, "EMAIL_USERNAME"), "EMAIL_USERNAME not available")
	utils.LogError(viperConfig.BindEnv(EmailPassword, "EMAIL_PASSWORD"), "EMAIL_PASSWORD not available")
	utils.LogError(viperConfig.BindEnv(EmailFrom, "EMAIL_FROM"), "EMAIL_FROM not available")

	utils.LogError(viperConfig.BindEnv(RecaptchaSiteKey, "RECAPTCHA_SITEKEY"), "RECAPTCHA_SITEKEY not available")
	utils.LogError(viperConfig.BindEnv(RecaptchaSecretKey, "RECAPTCHA_SECRETKEY"), "RECAPTCHA_SECRETKEY not available")

	utils.LogError(viperConfig.BindEnv(ServerPort, "PORT"), "PORT not available")
	utils.LogError(viperConfig.BindEnv(ServerSessionKey, "SERVER_SESSION_KEY"), "SERVER_SESSION_KEY not available")
	utils.LogError(viperConfig.BindEnv(ServerURL, "SERVER_URL"), "SERVER_URL not available")

	utils.LogError(viperConfig.BindEnv(MonitoringGoogleAnalytics, "MONITORING_GOOGLE_ANALYTICS"), "MONITORING_GOOGLE_ANALYTICS not available")

	utils.LogError(viperConfig.BindEnv(FeedbackForm, "MISCELLANEOUS_FEEDBACKFORM"), "MISCELLANEOUS_FEEDBACKFORM not available")
}
