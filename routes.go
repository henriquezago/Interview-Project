package main

import "net/http"

func (a *app) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api", a.helloHandler)
	mux.HandleFunc("GET /api/reports/{reportID}/presence", a.presenceHandler)

	return mux
}
