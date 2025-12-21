package utils

import (
	"log"
	"mime/multipart"
	"net/http"
)

// GetIP gets a requests IP address by reading off the forwarded-for
// header (for proxies) and falls back to use the remote address.
func GetIP(req *http.Request) string {
	forwarded := req.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return req.RemoteAddr
}

func CloseMultipartFile(csvFile multipart.File) {
	err := csvFile.Close()
	if err != nil {
		log.Printf("CloseMultipartFile: %v", err)
	}
}
