package middlewares

import (
	"log"
	"net/http"
	"strings"

	"com.code.vidmicro/com.code.vidmicro/httpHandler/responses"
	"com.code.vidmicro/com.code.vidmicro/settings/configmanager"
	"github.com/gin-gonic/gin"
)

type ApiMiddleware struct {
}

func (u *ApiMiddleware) GetHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		splittedString := strings.Split(c.Request.URL.Path, "/")
		if len(splittedString) < 3 {
			c.AbortWithStatusJSON(http.StatusNotFound, responses.GetInstance().WriteResponse(c, responses.NOT_FOUND, nil, nil))
			return
		}
		urlPath := "/" + splittedString[1] + "/" + splittedString[2]

		log.Println(urlPath, c.Request.Method, configmanager.GetInstance().Apis[urlPath])

		if methods, ok := configmanager.GetInstance().Apis[urlPath]; ok {
			if methods.Contains(c.Request.Method) {
				c.Set("apiPath", urlPath)
				c.Next()
			} else {
				c.AbortWithStatusJSON(http.StatusMethodNotAllowed, responses.GetInstance().WriteResponse(c, responses.METHOD_NOT_AVAILABLE, nil, nil))
				return
			}
		} else {
			c.AbortWithStatusJSON(http.StatusNotFound, responses.GetInstance().WriteResponse(c, responses.NOT_FOUND, nil, nil))
			return
		}
	}
}
