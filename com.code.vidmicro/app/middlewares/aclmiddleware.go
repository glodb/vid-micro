package middlewares

import (
	"log"
	"net/http"

	"com.code.vidmicro/com.code.vidmicro/httpHandler/responses"
	"com.code.vidmicro/com.code.vidmicro/settings/configmanager"
	"github.com/gin-gonic/gin"
)

type ACLMiddleware struct {
}

func (u *ACLMiddleware) GetHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println(c.GetString("roleName"))
		if val, ok := configmanager.GetInstance().Acl[c.GetString("roleName")]; !ok {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.API_NOT_ACCESSABLE, nil, nil))
			return
		} else {
			if innerVal, ok := val[c.GetString("apiPath")]; !ok {
			} else {
				if !innerVal.Contains(c.Request.Method) {
					c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.API_NOT_ACCESSABLE, nil, nil))
					return
				} else {
					c.Next()
				}
			}
		}
	}
}
