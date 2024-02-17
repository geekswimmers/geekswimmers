package main

import (
	"flag"
	"geekswimmers/config"
	"geekswimmers/server"
	"geekswimmers/storage"
	"geekswimmers/utils"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/cors"

	_ "github.com/heroku/x/hmetrics/onload"
)

var (
	flgConfigPath = flag.String("cfg", config.DefaultConfigFile, "Path to the server configuration file")
)

func run() error {
	log.Println("Loading configuration")
	conf, err := loadConfiguration()
	if err != nil {
		return err
	}

	log.Println("Migrating database")
	if err = storage.MigrateDatabase(conf); err != nil {
		return err
	}

	log.Println("Initializing database connection pool")
	db, err := storage.InitializeConnectionPool(conf)
	if err != nil {
		return err
	}

	if conf.GetString(config.ServerSessionKey) != "" {
		storage.InitSessionStore(conf)
	}

	log.Println("Creating GeekSwimmers server")
	s := server.CreateServer(conf, db)

	runHTTPServer(s, defineHTTPPort(conf))

	return nil
}

func loadConfiguration() (config.Config, error) {
	config, err := config.InitConfiguration(*flgConfigPath)
	if err != nil {
		return nil, errors.Wrap(err, "loading configuration")
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
