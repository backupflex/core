package agents

import (
	"sync"
	"time"
)

// StatusTracker manages volatile agent liveness/status in-memory.
//
// It is separate from the Repository because status data is ephemeral and
// should not be persisted to storage (it resets on restart).
type StatusTracker struct {
	mu    sync.RWMutex
	state map[string]RuntimeStatus
}

// NewStatusTracker creates an empty in-memory status store.
func NewStatusTracker() *StatusTracker {
	return &StatusTracker{
		mu:    sync.RWMutex{},
		state: make(map[string]RuntimeStatus),
	}
}

// Touch updates last seen timestamp and marks agent as online.
func (t *StatusTracker) Touch(id string, at time.Time) RuntimeStatus {
	t.mu.Lock()
	t.state[id] = RuntimeStatus{
		Status:     StatusOnline,
		LastSeenAt: at,
		UpdatedAt:  at,
	}
	t.mu.Unlock()

	return RuntimeStatus{
		Status:     StatusOnline,
		LastSeenAt: at,
		UpdatedAt:  at,
	}
}

// SetOffline marks an agent as offline.
func (t *StatusTracker) SetOffline(id string, at time.Time) {
	t.mu.Lock()
	current := t.state[id]
	current.Status = StatusOffline
	current.UpdatedAt = at
	t.state[id] = current
	t.mu.Unlock()
}

// Get returns the current runtime status for an agent.
// If the agent has no recorded status, returns StatusOffline with zero times.
func (t *StatusTracker) Get(id string) RuntimeStatus {
	t.mu.RLock()
	state, ok := t.state[id]
	t.mu.RUnlock()

	if !ok {
		return RuntimeStatus{
			Status:     StatusOffline,
			LastSeenAt: time.Time{},
			UpdatedAt:  time.Time{},
		}
	}
	return state
}

// Remove deletes the runtime status for an agent.
func (t *StatusTracker) Remove(id string) {
	t.mu.Lock()
	delete(t.state, id)
	t.mu.Unlock()
}
