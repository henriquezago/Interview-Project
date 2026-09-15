package main

import (
	"reflect"
	"testing"
	"time"
)

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
	weekly := service.forReport("weekly")
	monthly := service.forReport("monthly")

	weeklyViewer := weekly.subscribe("Cole")
	monthlyViewer := monthly.subscribe("Sarah")

	assertUpdate(t, weeklyViewer, []string{"Cole"})
	assertUpdate(t, monthlyViewer, []string{"Sarah"})
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
