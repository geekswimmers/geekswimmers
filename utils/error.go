package utils

import (
	"log"
	"net/http"
)

func ErrorHandler(res http.ResponseWriter, req *http.Request, ctx any, status int) {
	res.WriteHeader(status)

	if status == http.StatusNotFound {
		html := GetTemplate("base", "not-found")
		err := html.Execute(res, ctx)
		if err != nil {
			log.Print(err)
		}
	}
}
