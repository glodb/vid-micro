package cookie

import (
	"sync"

	"com.code.vidmicro/com.code.vidmicro/settings/configmanager"
	"github.com/gorilla/securecookie"
)

var (
	once     sync.Once
	instance *cookie
)

type cookie struct {
	secureCookie *securecookie.SecureCookie
}

func GetInstance() *cookie {
	once.Do(func() {
		instance = &cookie{}
		instance.secureCookie = securecookie.New([]byte(configmanager.GetInstance().SecureCookieHash), []byte(configmanager.GetInstance().SecureCookieBlock))
	})
	return instance
}

func (c *cookie) GetCookie() *securecookie.SecureCookie {
	return c.secureCookie
}
