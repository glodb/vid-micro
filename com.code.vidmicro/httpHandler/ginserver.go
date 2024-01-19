package httpHandler

import (
	"net/http"
	"sync"

	"com.code.vidmicro/com.code.vidmicro/app/middlewares"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/baserouter"
	"com.code.vidmicro/com.code.vidmicro/settings/configmanager"
	"github.com/gin-gonic/gin"
)

type GINServer struct {
	engine *gin.Engine
}

var instance *GINServer
var once sync.Once

// Singleton. Returns a single object of Factory
func GetInstance() *GINServer {
	// var instance
	once.Do(func() {
		instance = &GINServer{}
		if configmanager.GetInstance().IsProduction {
			gin.SetMode(gin.ReleaseMode)
			instance.engine = gin.Default()
		} else {
			instance.engine = gin.New()
			instance.engine.Use(gin.Recovery())
		}
		instance.Setup()
	})
	return instance
}
func (u *GINServer) GetEngine() *gin.Engine {
	return u.engine
}

func (u *GINServer) HandleBlank() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, map[string]interface{}{"msg": "Welcome to gpsina gin api"})
	}
}

func (u *GINServer) Setup() {
	corsMiddleware := middlewares.CORSMiddleware{}
	u.engine.Use(corsMiddleware.GetHandlerFunc())

	u.engine.GET("/", u.HandleBlank())

	// Check if Api available
	apiMiddleware := middlewares.ApiMiddleware{}
	u.engine.Use(apiMiddleware.GetHandlerFunc())

	baseginrouter := u.engine.Group("/")
	baserouter.GetInstance().SetRouter("base", baseginrouter)

	sessionMiddleware := middlewares.SessionMiddleware{}

	baseopenrouter := baseginrouter.Group("/")
	baseopenrouter.Use(sessionMiddleware.GetHandlerFunc())

	baserouter.GetInstance().SetRouter("open", baseopenrouter)

	authMiddleware := middlewares.AuthMiddleware{}
	baseauthwriter := baseopenrouter.Group("/", authMiddleware.GetHandlerFunc())

	aclMiddleware := middlewares.ACLMiddleware{}
	baseauthwriter.Use(aclMiddleware.GetHandlerFunc())

	baserouter.GetInstance().SetRouter("auth", baseauthwriter)

}
