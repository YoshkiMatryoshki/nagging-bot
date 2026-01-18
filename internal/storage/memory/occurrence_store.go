package memory

import (
	"context"
	"sync"
	"time"

	"naggingbot/internal/domain"
)

// InMemoryOccurrenceStore is an in-memory implementation of domain.OccurrenceStore.
type InMemoryOccurrenceStore struct {
	mu     sync.Mutex
	nextID int64
	byID   map[int64]*domain.Occurrence
}

// NewInMemoryOccurrenceStore constructs an empty in-memory store.
func NewInMemoryOccurrenceStore() *InMemoryOccurrenceStore {
	return &InMemoryOccurrenceStore{
		nextID: 1,
		byID:   make(map[int64]*domain.Occurrence),
	}
}

func (s *InMemoryOccurrenceStore) GetByID(ctx context.Context, id int64) (*domain.Occurrence, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	occ, ok := s.byID[id]
	if !ok {
		return nil, nil
	}

	return cloneOccurrence(occ), nil
}

func (s *InMemoryOccurrenceStore) ListByReminder(ctx context.Context, reminderID int64) ([]*domain.Occurrence, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var out []*domain.Occurrence
	for _, occ := range s.byID {
		if occ.ReminderID == reminderID {
			out = append(out, cloneOccurrence(occ))
		}
	}

	return out, nil
}

func (s *InMemoryOccurrenceStore) ListPendingInRange(ctx context.Context, startUTC, endUTC time.Time) ([]*domain.Occurrence, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var out []*domain.Occurrence
	for _, occ := range s.byID {
		if occ.Status != domain.OccurrenceCreated {
			continue
		}

		if !startUTC.IsZero() && occ.FireAtUtc.Before(startUTC) {
			continue
		}

		if occ.FireAtUtc.After(endUTC) {
			continue
		}

		out = append(out, cloneOccurrence(occ))
	}

	return out, nil
}

func (s *InMemoryOccurrenceStore) Create(ctx context.Context, occ *domain.Occurrence) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if occ.ID == 0 {
		occ.ID = s.nextID
		s.nextID++
	}

	s.byID[occ.ID] = cloneOccurrence(occ)
	return nil
}

func (s *InMemoryOccurrenceStore) UpdateStatus(ctx context.Context, id int64, status domain.OccurrenceStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	occ, ok := s.byID[id]
	if !ok {
		return nil
	}

	occ.Status = status
	return nil
}

func cloneOccurrence(occ *domain.Occurrence) *domain.Occurrence {
	c := *occ
	return &c
}
