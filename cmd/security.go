package main

import (
	"encoding/base32"
	"log"

	"github.com/gorilla/securecookie"
)

func main() {
	randomKey := securecookie.GenerateRandomKey(32)
	log.Printf("%v", base32.StdEncoding.EncodeToString(randomKey))
}
