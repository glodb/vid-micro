package middlewares

import (
	"log"

	"github.com/gin-gonic/gin"
)

type ACLMiddleware struct {
}

func (u *ACLMiddleware) GetHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println(c.GetInt("role"))
		// roleKey := config.GetMapKeyString("mapAcl", strconv.FormatInt(int64(c.GetInt("role")), 10))
		// aclMap := config.GetMapSet("acl", roleKey)
		// urlPath := "/" + strings.Split(c.Request.URL.Path, "/")[1] + "/" + strings.Split(c.Request.URL.Path, "/")[2]
		// if _, ok := aclMap[urlPath]; !ok {
		// 	c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.API_NOT_ACCESSABLE, nil, nil))
		// 	return
		// }
		c.Next()
	}
}
