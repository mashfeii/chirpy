package api

import (
	"net/http"
)

func HealthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)

	_, err := w.Write([]byte("OK"))
	if err != nil {
		panic(err)
	}
}
