package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"
)

const integrationTestTimeout = 5 * time.Second

type presenceTestConnection struct {
	reader *bufio.Reader
	body   io.ReadCloser
	cancel context.CancelFunc
}

func (c *presenceTestConnection) close() {
	c.cancel()
	_ = c.body.Close()
}

func (c *presenceTestConnection) next(t *testing.T) []string {
	t.Helper()

	for {
		line, err := c.reader.ReadString('\n')
		if err != nil {
			t.Fatalf("waiting for presence update: %v", err)
		}

		data, ok := strings.CutPrefix(strings.TrimRight(line, "\r\n"), "data: ")
		if !ok {
			continue
		}

		var payload struct {
			Viewers []string `json:"viewers"`
		}
		if err := json.Unmarshal([]byte(data), &payload); err != nil {
			t.Fatalf("decoding presence update %q: %v", data, err)
		}
		return payload.Viewers
	}
}

func newPresenceTestServer(t *testing.T) (*app, string) {
	t.Helper()

	a := &app{
		logger:   slog.New(slog.DiscardHandler),
		presence: newPresenceService(),
	}
	server := httptest.NewServer(a.routes())
	t.Cleanup(server.Close)
	return a, server.URL
}

func connectPresence(t *testing.T, baseURL, reportID, name string) *presenceTestConnection {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), integrationTestTimeout)
	target := fmt.Sprintf(
		"%s/api/reports/%s/presence?name=%s",
		baseURL,
		url.PathEscape(reportID),
		url.QueryEscape(name),
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		cancel()
		t.Fatalf("creating presence request: %v", err)
	}

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		cancel()
		t.Fatalf("connecting %q: %v", name, err)
	}
	if response.StatusCode != http.StatusOK {
		_ = response.Body.Close()
		cancel()
		t.Fatalf("connecting %q: status = %d, want %d", name, response.StatusCode, http.StatusOK)
	}
	if contentType := response.Header.Get("Content-Type"); contentType != "text/event-stream" {
		_ = response.Body.Close()
		cancel()
		t.Fatalf("Content-Type = %q, want text/event-stream", contentType)
	}

	connection := &presenceTestConnection{
		reader: bufio.NewReader(response.Body),
		body:   response.Body,
		cancel: cancel,
	}
	t.Cleanup(connection.close)
	return connection
}

func TestPresenceHandlerStreamsViewerLifecycle(t *testing.T) {
	_, baseURL := newPresenceTestServer(t)

	cole := connectPresence(t, baseURL, "weekly-performance", "Cole")
	assertViewers(t, cole.next(t), []string{"Cole"}, "first viewer connects")

	sarah := connectPresence(t, baseURL, "weekly-performance", "Sarah")
	assertViewers(t, cole.next(t), []string{"Cole", "Sarah"}, "second viewer connects")
	assertViewers(t, sarah.next(t), []string{"Cole", "Sarah"}, "second viewer receives initial state")

	sarah.close()
	assertViewers(t, cole.next(t), []string{"Cole"}, "second viewer disconnects")
}

func TestPresenceHandlerValidatesName(t *testing.T) {
	_, baseURL := newPresenceTestServer(t)

	for name, target := range map[string]string{
		"missing":  baseURL + "/api/reports/weekly-performance/presence",
		"too long": baseURL + "/api/reports/weekly-performance/presence?name=" + url.QueryEscape(strings.Repeat("é", 51)),
	} {
		t.Run(name, func(t *testing.T) {
			response, err := http.Get(target)
			if err != nil {
				t.Fatalf("requesting presence: %v", err)
			}
			defer response.Body.Close()

			if response.StatusCode != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusBadRequest)
			}
		})
	}
}

func TestPresenceHubBroadcastsJoinsAndLeaves(t *testing.T) {
	hub := &presenceHub{subscriptions: make(map[*presenceSubscription]struct{})}
	cole := hub.subscribe("Cole")
	assertUpdate(t, cole, []string{"Cole"})

	sarah := hub.subscribe("Sarah")
	assertUpdate(t, cole, []string{"Cole", "Sarah"})
	assertUpdate(t, sarah, []string{"Cole", "Sarah"})

	hub.unsubscribe(cole)
	assertUpdate(t, sarah, []string{"Sarah"})
}

func TestPresenceHubDeduplicatesNamesAcrossConnections(t *testing.T) {
	hub := &presenceHub{subscriptions: make(map[*presenceSubscription]struct{})}
	first := hub.subscribe("Cole")
	second := hub.subscribe("Cole")

	assertUpdate(t, first, []string{"Cole"})
	assertUpdate(t, second, []string{"Cole"})

	hub.unsubscribe(first)
	assertUpdate(t, second, []string{"Cole"})

	hub.unsubscribe(second)
	hub.mu.Lock()
	viewers := hub.snapshotLocked()
	hub.mu.Unlock()
	if len(viewers) != 0 {
		t.Fatalf("expected no viewers, got %v", viewers)
	}
}

func TestPresenceServiceSeparatesReports(t *testing.T) {
	service := newPresenceService()
	weeklyViewer := service.subscribe("weekly", "Cole")
	monthlyViewer := service.subscribe("monthly", "Sarah")

	assertUpdate(t, weeklyViewer, []string{"Cole"})
	assertUpdate(t, monthlyViewer, []string{"Sarah"})
}

func TestPresenceServiceRemovesEmptyReports(t *testing.T) {
	service := newPresenceService()
	first := service.subscribe("weekly", "Cole")
	second := service.subscribe("weekly", "Sarah")
	assertUpdate(t, first, []string{"Cole", "Sarah"})
	assertUpdate(t, second, []string{"Cole", "Sarah"})

	service.unsubscribe("weekly", first)
	service.mu.Lock()
	_, existsAfterFirstLeave := service.reports["weekly"]
	service.mu.Unlock()
	if !existsAfterFirstLeave {
		t.Fatal("report removed while it still had an active subscription")
	}

	service.unsubscribe("weekly", second)
	service.mu.Lock()
	_, existsAfterLastLeave := service.reports["weekly"]
	service.mu.Unlock()
	if existsAfterLastLeave {
		t.Fatal("empty report was not removed")
	}
}

func assertViewers(t *testing.T, got, want []string, event string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s: got viewers %v, want %v", event, got, want)
	}
}

func assertUpdate(t *testing.T, subscription *presenceSubscription, want []string) {
	t.Helper()
	select {
	case got := <-subscription.updates:
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got viewers %v, want %v", got, want)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for a presence update")
	}
}
