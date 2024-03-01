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

	// ServerPort is the HTTP port used to serve the application.
	ServerPort = "server.port"
	// ServerSessionKey is the key used to encrypt the session cookie.
	ServerSessionKey = "server.sessionkey"

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
	utils.LogError(viperConfig.BindEnv(DatabaseURL, "DATABASE_URL"), "Error binding environment variable DATABASE_URL")
	utils.LogError(viperConfig.BindEnv(DatabaseMaxOpenConns, "DATABASE_MAXOPENCONNS"), "Error binding environment variable DATABASE_MAXOPENCONNS")
	utils.LogError(viperConfig.BindEnv(DatabaseConnMaxLifetime, "DATABASE_CONNMAXLIFETIME"), "Error binding environment variable DATABASE_CONNMAXLIFETIME")

	utils.LogError(viperConfig.BindEnv(ServerPort, "PORT"), "Error binding environment variable PORT")
	utils.LogError(viperConfig.BindEnv(ServerSessionKey, "SERVER_SESSION_KEY"), "Error binding environment variable SERVER_SESSION_KEY")

	utils.LogError(viperConfig.BindEnv(MonitoringGoogleAnalytics, "MONITORING_GOOGLE_ANALYTICS"), "Error binding environment variable MONITORING_GOOGLE_ANALYTICS")

	utils.LogError(viperConfig.BindEnv(FeedbackForm, "MISCELLANEOUS_FEEDBACKFORM"), "Error binding environment variable MISCELLANEOUS_FEEDBACKFORM")
}
