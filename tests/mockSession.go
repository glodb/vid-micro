package tests

import (
	"net/http"

	"github.com/gorilla/sessions"
)

// MockStore is a mock implementation of the sessions.Store interface
type MockStore struct {
	Sessions map[string]*sessions.Session
}

// New creates a new session instance in the mock store
func (ms *MockStore) New(r *http.Request, name string) (*sessions.Session, error) {
	session := sessions.NewSession(ms, name)
	return session, nil
}

// Get returns a mock session from the store
func (ms *MockStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	if session, ok := ms.Sessions[name]; ok {
		return session, nil
	}
	return ms.New(r, name)
}

// Save does nothing in the mock store
func (ms *MockStore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	return nil
}

// NewMockStore creates a new instance of MockStore
func NewMockStore() *MockStore {
	return &MockStore{
		Sessions: make(map[string]*sessions.Session),
	}
}
