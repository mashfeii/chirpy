package api

import (
	"log"
	"net/http"
)

func MiddlewareLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s\t%s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
