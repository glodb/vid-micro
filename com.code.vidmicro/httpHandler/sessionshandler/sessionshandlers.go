package sessionshandler

import (
	"sync"

	"com.code.vidmicro/com.code.vidmicro/settings/configmanager"
	"github.com/gorilla/sessions"
)

type session struct {
	store *sessions.CookieStore
}

var (
	instance *session
	once     sync.Once
)

// Singleton. Returns a single object of Factory
func GetInstance() *session {
	// var instance
	once.Do(func() {
		instance = &session{}
		instance.store = sessions.NewCookieStore([]byte(configmanager.GetInstance().SessionKey))
	})
	return instance
}

func (s *session) GetSession() *sessions.CookieStore {
	return s.store
}
