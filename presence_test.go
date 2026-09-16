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
