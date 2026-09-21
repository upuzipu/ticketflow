// Package realtime manages WebSocket subscriptions and broadcasts.
package realtime

import (
	"encoding/json"
	"sync"

	"github.com/upuzipu/ticketflow/internal/domain"
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
	ID      string // уникальный id для логов
	EventID string
	Ch      chan []byte // outbound messages; writer goroutine reads it
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

// BroadcastAvailability sends the update to everyone subscribed
// to the event. Slow subscribers are dropped (buffer full → skip).
func (h *Hub) BroadcastAvailability(eventID string, stats []domain.CategoryAvailability) {
	payload, err := json.Marshal(map[string]any{
		"type":       "availability",
		"event_id":   eventID,
		"categories": stats,
	})
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	for s := range h.rooms[eventID] {
		select {
		case s.Ch <- payload:
		default:
		}
	}
}
