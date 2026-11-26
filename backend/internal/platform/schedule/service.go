package schedule

import (
	"context"
	"log"
	"sync"
	"time"
)

type ChangeRepository interface {
	Listen(ctx context.Context, notify func()) error
}

type ScheduleService struct {
	repo    ChangeRepository
	mu      sync.Mutex
	clients map[chan struct{}]struct{}
}

func NewScheduleService(repo ChangeRepository) *ScheduleService {
	return &ScheduleService{repo: repo, clients: make(map[chan struct{}]struct{})}
}

func (s *ScheduleService) Subscribe() (<-chan struct{}, func()) {
	changes := make(chan struct{}, 1)
	changes <- struct{}{}
	s.mu.Lock()
	s.clients[changes] = struct{}{}
	s.mu.Unlock()
	return changes, func() {
		s.mu.Lock()
		delete(s.clients, changes)
		s.mu.Unlock()
	}
}

func (s *ScheduleService) broadcast() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for changes := range s.clients {
		select {
		case changes <- struct{}{}:
		default:
		}
	}
}

func (s *ScheduleService) Run(ctx context.Context) {
	for ctx.Err() == nil {
		err := s.repo.Listen(ctx, s.broadcast)
		if ctx.Err() != nil {
			return
		}
		log.Printf("schedule: lost database subscription, reconnecting: %v", err)
		select {
		case <-ctx.Done():
			return
		case <-time.After(2 * time.Second):
		}
	}
}
