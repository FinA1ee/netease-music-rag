package service

import (
	"sync"
)

// ─── Event types ──────────────────────────────────────────────────────────────

type EventType string

const (
	EvJobStarted         EventType = "job_started"
	EvPlaylistProcessing EventType = "playlist_processing"
	EvSongAnalysed       EventType = "song_analysed"
	EvSongSkipped        EventType = "song_skipped"
	EvJobCompleted       EventType = "job_completed"
	EvEmbeddingStarted   EventType = "embedding_started"
	EvEmbeddingDone      EventType = "embedding_done"
)

// JobEvent is the payload pushed to connected SSE clients.
type JobEvent struct {
	Type    EventType      `json:"type"`
	Payload map[string]any `json:"payload,omitempty"`
}

// ─── EventBus ─────────────────────────────────────────────────────────────────

// EventBus is a thread-safe fan-out broadcaster. Each connected SSE client
// gets its own buffered channel. Slow clients are dropped rather than blocking.
// The last recentMax events are buffered so reconnecting clients can replay them.
type EventBus struct {
	mu        sync.RWMutex
	clients   map[chan JobEvent]struct{}
	recent    []JobEvent // ring buffer
	recentMax int
}

const defaultRecentMax = 20

func NewEventBus() *EventBus {
	return &EventBus{
		clients:   make(map[chan JobEvent]struct{}),
		recentMax: defaultRecentMax,
	}
}

// Subscribe returns a channel that will receive all published events.
// The caller must call Unsubscribe when done to avoid leaks.
func (b *EventBus) Subscribe() chan JobEvent {
	ch := make(chan JobEvent, 32) // buffered so slow clients don't block publish
	b.mu.Lock()
	b.clients[ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

// Unsubscribe removes and closes the channel.
func (b *EventBus) Unsubscribe(ch chan JobEvent) {
	b.mu.Lock()
	delete(b.clients, ch)
	b.mu.Unlock()
	close(ch)
}

// Publish sends an event to all subscribers and appends it to the replay buffer.
func (b *EventBus) Publish(ev JobEvent) {
	b.mu.Lock()
	// Append to ring buffer
	b.recent = append(b.recent, ev)
	if len(b.recent) > b.recentMax {
		b.recent = b.recent[len(b.recent)-b.recentMax:]
	}
	clients := make([]chan JobEvent, 0, len(b.clients))
	for ch := range b.clients {
		clients = append(clients, ch)
	}
	b.mu.Unlock()

	for _, ch := range clients {
		select {
		case ch <- ev:
		default: // client too slow — skip rather than block
		}
	}
}

// RecentEvents returns a copy of the replay buffer.
func (b *EventBus) RecentEvents() []JobEvent {
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := make([]JobEvent, len(b.recent))
	copy(out, b.recent)
	return out
}

// ─── Convenience helpers ──────────────────────────────────────────────────────

func (b *EventBus) emit(t EventType, payload map[string]any) {
	b.Publish(JobEvent{Type: t, Payload: payload})
}
