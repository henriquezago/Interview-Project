package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"
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

func (a *app) presenceHandler(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.URL.Query().Get("name"))
	if name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	if !utf8.ValidString(name) || utf8.RuneCountInString(name) > 50 {
		http.Error(w, "name must be 50 characters or fewer", http.StatusBadRequest)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming is not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	reportID := r.PathValue("reportID")
	subscription := a.presence.subscribe(reportID, name)
	defer a.presence.unsubscribe(reportID, subscription)

	keepAlive := time.NewTicker(15 * time.Second)
	defer keepAlive.Stop()

	for {
		select {
		case viewers := <-subscription.updates:
			payload, err := json.Marshal(struct {
				Viewers []string `json:"viewers"`
			}{Viewers: viewers})
			if err != nil {
				a.logger.Error("encoding presence update", "error", err)
				return
			}
			if _, err := fmt.Fprintf(w, "data: %s\n\n", payload); err != nil {
				return
			}
			flusher.Flush()
		case <-keepAlive.C:
			if _, err := fmt.Fprint(w, ": keepalive\n\n"); err != nil {
				return
			}
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}
