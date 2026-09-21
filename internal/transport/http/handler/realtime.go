package handler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/coder/websocket"

	"github.com/upuzipu/ticketflow/internal/domain"
	"github.com/upuzipu/ticketflow/internal/realtime"
	"github.com/upuzipu/ticketflow/internal/transport/http/dto"
)

// RealtimeHandler serves the WebSocket endpoint.
type RealtimeHandler struct {
	hub       *realtime.Hub
	loadStats func(ctx context.Context, eventID string) ([]domain.CategoryAvailability, error)
	log       *slog.Logger
}

// NewRealtimeHandler wires the handler.
func NewRealtimeHandler(hub *realtime.Hub,
	loadStats func(ctx context.Context, eventID string) ([]domain.CategoryAvailability, error)) *RealtimeHandler {
	return &RealtimeHandler{hub: hub, loadStats: loadStats, log: slog.Default()}
}

// Subscribe handles GET /ws/events/{id}.
func (h *RealtimeHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	eventID := r.PathValue("id")

	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close(websocket.StatusInternalError, "closing")

	sub := &realtime.Subscriber{
		ID:      fmt.Sprintf("%s-%s", eventID, time.Now().Format("15:04:05.000")),
		EventID: eventID,
		Ch:      make(chan []byte, 8),
	}
	h.hub.Subscribe(sub)
	defer h.hub.Unsubscribe(sub)
	h.log.Info("ws subscriber joined", "id", sub.ID)

	ctx := r.Context()

	// initial snapshot
	go func() {
		stats, err := h.loadStats(r.Context(), eventID)
		if err != nil {
			return
		}
		payload, err := dto.AvailabilityFrameFromDomain(eventID, stats)
		if err != nil {
			return
		}
		select {
		case sub.Ch <- payload:
		default:
		}
	}()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-sub.Ch:
				if !ok {
					return
				}
				writeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
				err := conn.Write(writeCtx, websocket.MessageText, msg)
				cancel()
				if err != nil {
					return
				}
			}
		}
	}()

	conn.Read(ctx)
	<-done
	h.log.Info("ws subscriber left", "id", sub.ID)
}
