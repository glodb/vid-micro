package baserouter

import (
	"sync"

	"github.com/gorilla/mux"
)

type baseRouterHandler struct {
	router map[string]*mux.Router
}

var instance *baseRouterHandler
var once sync.Once

//Singleton. Returns a single object of Factory
func GetInstance() *baseRouterHandler {
	// var instance
	once.Do(func() {
		instance = &baseRouterHandler{}
		instance.router = make(map[string]*mux.Router)
	})
	return instance
}

func (u *baseRouterHandler) SetRouter(name string, router *mux.Router) {
	u.router[name] = router
}

func (u *baseRouterHandler) GetBaseRouter() *mux.Router {
	return u.router["base"]
}

func (u *baseRouterHandler) GetOpenRouter() *mux.Router {
	return u.router["open"]
}

func (u *baseRouterHandler) GetLoginRouter() *mux.Router {
	return u.router["auth"]
}
