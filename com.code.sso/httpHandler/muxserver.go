package httpHandler

import (
	"log"
	"net/http"
	"sync"

	"com.code.sso/com.code.sso/app/middlewares"
	"com.code.sso/com.code.sso/config"
	"com.code.sso/com.code.sso/httpHandler/basecontrollers"
	"com.code.sso/com.code.sso/httpHandler/baserouter"
	"com.code.sso/com.code.sso/httpHandler/responses"
	"github.com/gorilla/mux"
)

type muxServer struct {
	base *mux.Router
}

var (
	instance *muxServer
	once     sync.Once
)

//Singleton. Returns a single object
func GetInstance() *muxServer {
	// var instance
	once.Do(func() {
		instance = &muxServer{}
		instance.setup()
	})
	return instance
}

func (u *muxServer) HandleBlank(w http.ResponseWriter, r *http.Request) {
	responses.GetInstance().WriteJsonResponse(w, r, responses.WELCOME_TO_SSO, nil, nil)
}

func (u *muxServer) setup() {
	corsMiddleware := middlewares.CORSMiddleware{}
	u.base = &mux.Router{}

	u.base.Use(corsMiddleware.GetHandlerFunc)
	u.base.HandleFunc("/", u.HandleBlank).Methods("GET")

	baserouter.GetInstance().SetRouter("base", u.base)

	sessionMiddleware := middlewares.SessionMiddleware{}
	open := u.base.PathPrefix("/").Subrouter()
	open.Use(sessionMiddleware.GetHandlerFunc)
	baserouter.GetInstance().SetRouter("open", open)

	authMiddleware := middlewares.AuthMiddleware{}
	auth := open.PathPrefix("/").Subrouter()
	auth.Use(authMiddleware.GetHandlerFunc)

	baserouter.GetInstance().SetRouter("auth", auth)
}

func (u *muxServer) Start() {
	log.Println("Server listening on ", config.GetInstance().Server.Address, ":", config.GetInstance().Server.Port)
	basecontrollers.GetInstance().RegisterControllers()

	http.Handle("/", u.base)
	err := http.ListenAndServe(config.GetInstance().Server.Address+":"+config.GetInstance().Server.Port, nil)
	if err != nil {
		log.Println("Error in running server:", err)
	}
}
