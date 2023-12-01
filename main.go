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

	"github.com/bmizerany/pat"
	"github.com/pkg/errors"
	"github.com/rs/cors"
)

var (
	flgConfigPath = flag.String("cfg", config.DefaultConfigFile, "Path to the server configuration file")
)

func main() {
	if err := run(); err != nil {
		log.Printf("%s\n", err)
		os.Exit(1)
	}
}

func run() error {
	config, err := loadConfiguration()
	if err != nil {
		return err
	}

	if err := storage.MigrateDatabase(config); err != nil {
		return err
	}

	storage.InitSessionStore(config)

	db, err := storage.InitializeConnectionPool(config)
	if err != nil {
		return err
	}

	s := server.CreateServer(db, pat.New())

	runHTTPServer(s, defineHTTPPort(config))

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

	log.Printf("Serving on port: %v", port)
	if err := http.ListenAndServe(":"+port, c.Handler(server.Router)); err != nil {
		log.Fatal(err)
	}
}
