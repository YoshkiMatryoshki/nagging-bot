package memory

import (
	"context"
	"sync"

	"naggingbot/internal/domain"
)

// InMemoryUserStore is an in-memory implementation of domain.UserStore.
type InMemoryUserStore struct {
	mu     sync.Mutex
	nextID int64
	byID   map[int64]*domain.User
	byTGID map[int64]*domain.User
}

// NewInMemoryUserStore constructs an empty user store.
func NewInMemoryUserStore() *InMemoryUserStore {
	return &InMemoryUserStore{
		nextID: 1,
		byID:   make(map[int64]*domain.User),
		byTGID: make(map[int64]*domain.User),
	}
}

func (s *InMemoryUserStore) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	u, ok := s.byID[id]
	if !ok {
		return nil, nil
	}
	return cloneUser(u), nil
}

func (s *InMemoryUserStore) GetByTelegramID(ctx context.Context, telegramID int64) (*domain.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	u, ok := s.byTGID[telegramID]
	if !ok {
		return nil, nil
	}
	return cloneUser(u), nil
}

func (s *InMemoryUserStore) Upsert(ctx context.Context, user *domain.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Assign new ID if needed.
	if user.ID == 0 {
		user.ID = s.nextID
		s.nextID++
	}

	// Overwrite existing entries.
	s.byID[user.ID] = cloneUser(user)
	if user.TelegramID != 0 {
		s.byTGID[user.TelegramID] = cloneUser(user)
	}

	return nil
}

func cloneUser(u *domain.User) *domain.User {
	c := *u
	return &c
}
