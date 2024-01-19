package middlewares

import (
	"log"
	"strings"

	"github.com/gin-gonic/gin"
)

type ApiMiddleware struct {
}

func (u *ApiMiddleware) GetHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		urlPath := "/" + strings.Split(c.Request.URL.Path, "/")[1] + "/" + strings.Split(c.Request.URL.Path, "/")[2]
		log.Println(urlPath)
		// if _, ok := config.GetSet("apis")[urlPath]; ok {
		// 	c.Next()
		// } else {
		// 	c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.API_NOT_AVAILABLE, nil, nil))
		// }
	}
}
