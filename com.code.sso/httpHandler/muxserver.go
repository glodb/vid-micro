package httpHandler

import (
	"log"
	"net/http"
	"sync"

	"com.code.sso/com.code.sso/httpHandler/baserouter"
	"com.code.sso/com.code.sso/httpHandler/middlewares"
	"github.com/gorilla/mux"
)

type MuxServer struct {
}

var (
	instance *MuxServer
	once     sync.Once
)

//Singleton. Returns a single object of Factory
func GetInstance() *MuxServer {
	// var instance
	once.Do(func() {
		instance = &MuxServer{}
		instance.setup()
	})
	return instance
}

func (u *MuxServer) setup() {
	corsMiddleware, apiMiddleware := middlewares.CORSMiddleware{}, middlewares.ApiMiddleware{}
	base := &mux.Router{}
	base.Use(corsMiddleware.GetHandlerFunc, apiMiddleware.GetHandlerFunc)

	baserouter.GetInstance().SetRouter("base", base)

	sessionMiddleware := middlewares.SessionMiddleware{}
	open := &mux.Router{}
	open.Use(corsMiddleware.GetHandlerFunc, apiMiddleware.GetHandlerFunc, sessionMiddleware.GetHandlerFunc)
	baserouter.GetInstance().SetRouter("open", base)

	authMiddleware := middlewares.AuthMiddleware{}
	auth := &mux.Router{}
	auth.Use(corsMiddleware.GetHandlerFunc, apiMiddleware.GetHandlerFunc, sessionMiddleware.GetHandlerFunc, authMiddleware.GetHandlerFunc)

	baserouter.GetInstance().SetRouter("auth", auth)

}

func (u *MuxServer) Start() {
	log.Println("Server listening on :8080")
	http.ListenAndServe(":8080", nil)
}
