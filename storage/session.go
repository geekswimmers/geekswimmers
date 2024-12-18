package storage

import (
	"encoding/base32"
	"fmt"
	"geekswimmers/config"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
)

var sessionStore *sessions.CookieStore

type SessionData struct {
	AcceptedCookies bool
	BirthDate       string
	Confirmation    string
	Course          string
	Email           string
	Event           string
	Gender          string
	Jurisdiction    string
	Millisecond     string
	Minute          string
	Roles           []string
	Second          string
}

func NewSessionData(req *http.Request) *SessionData {
	return &SessionData{
		AcceptedCookies: GetSessionEntryValue(req, "profile", "acceptedCookies") == "true",
		Jurisdiction:    GetSessionEntryValue(req, "profile", "jurisdiction"),
		BirthDate:       GetSessionEntryValue(req, "profile", "birthDate"),
		Email:           GetSessionEntryValue(req, "profile", "email"),
		Gender:          GetSessionEntryValue(req, "profile", "gender"),
		Course:          GetSessionEntryValue(req, "profile", "course"),
		Event:           GetSessionEntryValue(req, "profile", "event"),
		Minute:          GetSessionEntryValue(req, "profile", "minute"),
		Second:          GetSessionEntryValue(req, "profile", "second"),
		Millisecond:     GetSessionEntryValue(req, "profile", "millisecond"),
	}
}

func InitSessionStore(c config.Config) error {
	sessionKey := c.GetString(config.ServerSessionKey)
	decodedKey, err := base32.StdEncoding.DecodeString(sessionKey)
	if err != nil {
		return fmt.Errorf("InitSessionStore: %v", err)
	}
	sessionStore = sessions.NewCookieStore(decodedKey)
	sessionStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   60 * 60 * 24 * 30, // 30 days
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	return nil
}

func SessionStoreAvailable() bool {
	return sessionStore != nil
}

func GetSessionEntryValue(req *http.Request, store, key string) string {
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
		log.Printf("AddSessionEntry.Get: %v, %v", err, session)
	}

	session.Values[key] = value

	if err = session.Save(req, res); err != nil {
		return fmt.Errorf("AddSessionEntry.Save: %v", err)
	}
	return nil
}

func RemoveSessionEntry(res http.ResponseWriter, req *http.Request, store, key string) error {
	session, err := sessionStore.Get(req, store)
	if err != nil {
		return fmt.Errorf("RemoveSessionEntry: %v", err)
	}

	session.Values[key] = nil
	if err = session.Save(req, res); err != nil {
		return fmt.Errorf("RemoveSessionEntry: %v", err)
	}

	return nil
}
