package main

import (
	"encoding/json"
	"net/http"
)

func (a *app) helloHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"message": "Hello, Frontend!",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
