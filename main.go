package main

import (
	"fmt"
	"geekswimmers/config"
	"geekswimmers/server"
	"geekswimmers/storage"
	"geekswimmers/utils"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/rs/cors"

	_ "github.com/heroku/x/hmetrics/onload"
)

func run() error {
	conf, err := loadConfiguration()
	if err != nil {
		return err
	}
	log.Println("Configuration loaded successfully.")

	if err = storage.MigrateDatabase(conf); err != nil {
		return fmt.Errorf("storage: %v", err)
	}

	db, err := storage.InitializeConnectionPool(conf)
	if err != nil {
		return fmt.Errorf("storage: %v", err)
	}
	log.Println("Database connection pool initialized successfully.")

	if conf.GetString(config.ServerSessionKey) != "" {
		if err := storage.InitSessionStore(conf); err != nil {
			return fmt.Errorf("storage: %v", err)
		}
	}

	s := server.CreateServer(conf, db)

	runHTTPServer(s, defineHTTPPort(conf))

	return nil
}

func loadConfiguration() (config.Config, error) {
	config, err := config.InitConfiguration(config.DefaultConfigFile)
	if err != nil {
		return nil, fmt.Errorf("loading configuration: %v", err)
	}

	return config, nil
}

func defineHTTPPort(c config.Config) string {
	port := c.GetString(config.ServerPort)
	if port == "" || !utils.IsNumeric(port) {
		port = "8080"
	}
	return port
}

func runHTTPServer(server *server.Server, port string) {
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowCredentials: true,
		Debug:            false,
	})

	log.Printf("Serving GeekSwimmers on port: %v", port)

	srvr := &http.Server{
		Addr:              ":" + port,
		ReadHeaderTimeout: 10 * time.Second,
		Handler:           c.Handler(server.Router),
	}

	if err := srvr.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	if err := run(); err != nil {
		log.Printf("%s\n", err)
		os.Exit(1)
	}
}
