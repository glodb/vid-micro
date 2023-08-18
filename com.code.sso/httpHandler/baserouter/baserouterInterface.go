package baserouter

import (
	"github.com/gorilla/mux"
)

type BaseRouter interface {
	SetRouter(name string, router *mux.Router)
	GetOpenRouter() *mux.Router
	GetLoginRouter() *mux.Router
	GetBaseRouter() *mux.Router
}
