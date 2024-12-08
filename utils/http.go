package utils

import "net/http"

// GetIP gets a requests IP address by reading off the forwarded-for
// header (for proxies) and falls back to use the remote address.
func GetIP(req *http.Request) string {
	forwarded := req.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return req.RemoteAddr
}
