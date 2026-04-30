package monitor

import (
	"sync"
	"time"
)

type StateStore struct {
	cooldown  time.Duration
	lastAlert map[string]time.Time
	status    map[string]string
	mu        sync.Mutex
}

func NewStateStore(cooldown time.Duration) *StateStore {
	return &StateStore{
		cooldown:  cooldown,
		lastAlert: make(map[string]time.Time),
		status:    make(map[string]string),
	}
}

func (s *StateStore) ShouldAlert(nodeName string, now time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	last, ok := s.lastAlert[nodeName]
	if !ok || now.Sub(last) >= s.cooldown {
		s.lastAlert[nodeName] = now
		return true
	}

	return false
}

func (s *StateStore) MarkError(nodeName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status[nodeName] = StatusError
}

func (s *StateStore) MarkSuccess(nodeName string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	wasError := s.status[nodeName] == StatusError
	s.status[nodeName] = StatusSuccess
	return wasError
}
