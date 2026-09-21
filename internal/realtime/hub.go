// Package realtime manages WebSocket subscriptions and broadcasts.
package realtime

import (
	"sync"
)

// Hub keeps WebSocket subscribers grouped by event ID
// and fans out availability updates to them.
type Hub struct {
	mu    sync.RWMutex
	rooms map[string]map[*Subscriber]struct{}
}

// NewHub creates an empty hub.
func NewHub() *Hub {
	return &Hub{rooms: make(map[string]map[*Subscriber]struct{})}
}

// Subscriber is one connected client.
type Subscriber struct {
	ID      string
	EventID string
	Ch      chan []byte
}

// Subscribe registers a subscriber in the event's room.
func (h *Hub) Subscribe(s *Subscriber) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[s.EventID] == nil {
		h.rooms[s.EventID] = make(map[*Subscriber]struct{})
	}
	h.rooms[s.EventID][s] = struct{}{}
}

// Unsubscribe removes the subscriber and closes its channel.
func (h *Hub) Unsubscribe(s *Subscriber) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if room, ok := h.rooms[s.EventID]; ok {
		delete(room, s)
		if len(room) == 0 {
			delete(h.rooms, s.EventID)
		}
	}
	close(s.Ch)
}

// BroadcastRaw sends the pre-encoded payload to every subscriber
// of the event. Slow subscribers are dropped, never blocking the hub.
func (h *Hub) BroadcastRaw(eventID string, payload []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for s := range h.rooms[eventID] {
		select {
		case s.Ch <- payload:
		default:
		}
	}
}
