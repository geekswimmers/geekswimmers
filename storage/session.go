package storage

import (
	"encoding/base32"
	"geekswimmers/config"
	"net/http"

	"github.com/gorilla/sessions"
)

var sessionStore *sessions.CookieStore

func InitSessionStore(c config.Config) {
	sessionKey := c.GetString(config.ServerSessionKey)
	decodedKey, _ := base32.StdEncoding.DecodeString(sessionKey)
	sessionStore = sessions.NewCookieStore(decodedKey)
}

func SessionAvailable() bool {
	return sessionStore != nil
}

func GetSessionValue(req *http.Request, store, key string) string {
	var value string

	session, err := sessionStore.Get(req, store)
	if err != nil {
		return value
	}

	if session.Values[key] != nil {
		value = session.Values[key].(string)
	}
	return value
}

func AddSessionEntry(res http.ResponseWriter, req *http.Request, store, key, value string) error {
	session, err := sessionStore.Get(req, store)
	if err != nil {
		return err
	}

	session.Values[key] = value

	if err = session.Save(req, res); err != nil {
		return err
	}
	return nil
}

func RemoveSessionEntry(res http.ResponseWriter, req *http.Request, store, key string) error {
	session, err := sessionStore.Get(req, store)
	if err != nil {
		return err
	}

	session.Values[key] = nil
	err = session.Save(req, res)
	return err
}
