package store

import "sync"

type TypeCommand string

var (
	UserAdminCreate TypeCommand = "create"
	UserAdminDelete TypeCommand = "delete"
)

type AdminStore struct {
	MsgID  int64
	UserID int64

	TypeCommand TypeCommand
}

type Store struct {
	store map[int64]interface{}

	mu sync.RWMutex
}

func NewStore() *Store {
	return &Store{
		store: make(map[int64]interface{}, 15),
	}
}

func (s *Store) Set(data interface{}, userID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[userID] = data
}

func (s *Store) Read(userID int64) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	d, ok := s.store[userID]
	if !ok {
		return nil, false
	}

	return d, true
}

func (s *Store) Delete(userID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.store, userID)
}
