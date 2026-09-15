package main

import (
	"sort"
	"strings"
	"sync"
)

// presenceService separates viewers by report while keeping all state in memory.
type presenceService struct {
	mu      sync.Mutex
	reports map[string]*presenceHub
}

func newPresenceService() *presenceService {
	return &presenceService{reports: make(map[string]*presenceHub)}
}

func (s *presenceService) forReport(reportID string) *presenceHub {
	s.mu.Lock()
	defer s.mu.Unlock()

	if hub, ok := s.reports[reportID]; ok {
		return hub
	}

	hub := &presenceHub{subscriptions: make(map[*presenceSubscription]struct{})}
	s.reports[reportID] = hub
	return hub
}

// Each browser connection gets its own subscription. The displayed snapshot is
// deduplicated by name so closing one of two tabs with the same name is safe.
type presenceHub struct {
	mu            sync.Mutex
	subscriptions map[*presenceSubscription]struct{}
}

type presenceSubscription struct {
	name    string
	updates chan []string
}

func (h *presenceHub) subscribe(name string) *presenceSubscription {
	subscription := &presenceSubscription{
		name:    name,
		updates: make(chan []string, 1),
	}

	h.mu.Lock()
	h.subscriptions[subscription] = struct{}{}
	h.broadcastLocked()
	h.mu.Unlock()

	return subscription
}

func (h *presenceHub) unsubscribe(subscription *presenceSubscription) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.subscriptions[subscription]; !ok {
		return
	}

	delete(h.subscriptions, subscription)
	h.broadcastLocked()
}

func (h *presenceHub) broadcastLocked() {
	viewers := h.snapshotLocked()
	for subscription := range h.subscriptions {
		// Presence is state, not an event log. If a slow client has not consumed
		// the previous state, replace it with the newest snapshot.
		select {
		case subscription.updates <- viewers:
		default:
			select {
			case <-subscription.updates:
			default:
			}
			subscription.updates <- viewers
		}
	}
}

func (h *presenceHub) snapshotLocked() []string {
	unique := make(map[string]struct{}, len(h.subscriptions))
	for subscription := range h.subscriptions {
		unique[subscription.name] = struct{}{}
	}

	viewers := make([]string, 0, len(unique))
	for name := range unique {
		viewers = append(viewers, name)
	}
	sort.Slice(viewers, func(i, j int) bool {
		left, right := strings.ToLower(viewers[i]), strings.ToLower(viewers[j])
		if left == right {
			return viewers[i] < viewers[j]
		}
		return left < right
	})
	return viewers
}
