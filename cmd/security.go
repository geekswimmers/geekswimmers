package main

import (
	"encoding/base32"
	"geekswimmers/config"
	"log"

	"github.com/gorilla/securecookie"
)

func main() {
	randomKey := securecookie.GenerateRandomKey(32)
	log.Printf("%v", randomKey)
	log.Printf("%v", base32.StdEncoding.EncodeToString(randomKey))

	key := config.GetConfiguration().GetString("server.sessionkey")
	decoded, _ := base32.StdEncoding.DecodeString(key)
	log.Printf("%v", decoded)
}
