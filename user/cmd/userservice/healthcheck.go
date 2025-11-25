package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func registerHealthcheck(router *mux.Router) {
	router.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}
