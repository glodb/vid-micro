package baserouter

import (
	"sync"

	"com.code.vidmicro/com.code.vidmicro/settings/configmanager"
	"github.com/gin-gonic/gin"
)

type baseRouterHandler struct {
	router map[string]*gin.RouterGroup
}

var instance *baseRouterHandler
var once sync.Once

// Singleton. Returns a single object of Factory
func GetInstance() *baseRouterHandler {
	// var instance
	once.Do(func() {
		instance = &baseRouterHandler{}
		instance.router = make(map[string]*gin.RouterGroup)
	})
	return instance
}

func (u *baseRouterHandler) SetRouter(name string, router *gin.RouterGroup) {
	u.router[name] = router
}

func (u *baseRouterHandler) GetBaseRouter(secret string) *gin.RouterGroup {
	if secret == configmanager.GetInstance().SessionKey {
		return u.router["base"]
	} else {
		return nil
	}
}

func (u *baseRouterHandler) GetOpenRouter() *gin.RouterGroup {
	return u.router["open"]
}

func (u *baseRouterHandler) GetLoginRouter() *gin.RouterGroup {
	return u.router["auth"]
}

func (u *baseRouterHandler) GetServerRouter() *gin.RouterGroup {
	return u.router["server"]
}
