package memory

import (
	"context"
	"sync"

	"naggingbot/internal/domain"
)

// InMemoryReminderStore is an in-memory implementation of domain.ReminderStore.
type InMemoryReminderStore struct {
	mu     sync.Mutex
	nextID int64
	byID   map[int64]*domain.Reminder
	byUser map[int64][]*domain.Reminder
}

// NewInMemoryReminderStore constructs an empty reminder store.
func NewInMemoryReminderStore() *InMemoryReminderStore {
	return &InMemoryReminderStore{
		nextID: 1,
		byID:   make(map[int64]*domain.Reminder),
		byUser: make(map[int64][]*domain.Reminder),
	}
}

func (s *InMemoryReminderStore) GetByID(ctx context.Context, id int64) (*domain.Reminder, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	r, ok := s.byID[id]
	if !ok {
		return nil, nil
	}
	return cloneReminder(r), nil
}

func (s *InMemoryReminderStore) ListByUser(ctx context.Context, userID int64) ([]*domain.Reminder, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rs := s.byUser[userID]
	out := make([]*domain.Reminder, 0, len(rs))
	for _, r := range rs {
		out = append(out, cloneReminder(r))
	}
	return out, nil
}

func (s *InMemoryReminderStore) Create(ctx context.Context, reminder *domain.Reminder) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if reminder.ID == 0 {
		reminder.ID = s.nextID
		s.nextID++
	}

	s.byID[reminder.ID] = cloneReminder(reminder)
	s.byUser[reminder.UserID] = append(s.byUser[reminder.UserID], cloneReminder(reminder))
	return nil
}

func (s *InMemoryReminderStore) Update(ctx context.Context, reminder *domain.Reminder) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Update primary map.
	s.byID[reminder.ID] = cloneReminder(reminder)

	// Rebuild user's slice entry.
	var updated []*domain.Reminder
	for _, r := range s.byUser[reminder.UserID] {
		if r.ID == reminder.ID {
			updated = append(updated, cloneReminder(reminder))
		} else {
			updated = append(updated, cloneReminder(r))
		}
	}
	s.byUser[reminder.UserID] = updated

	return nil
}

func cloneReminder(r *domain.Reminder) *domain.Reminder {
	c := *r
	if r.TimesOfDay != nil {
		c.TimesOfDay = append([]domain.TimeOfDay{}, r.TimesOfDay...)
	}
	return &c
}
