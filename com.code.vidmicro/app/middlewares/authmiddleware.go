package middlewares

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"

	"com.code.vidmicro/com.code.vidmicro/app/models"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/responses"
	"com.code.vidmicro/com.code.vidmicro/settings/configmanager"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type AuthMiddleware struct {
}

func (u *AuthMiddleware) isTokenValid(tokenString string) (bool, error) {

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Provide the key for token validation
		return []byte(configmanager.GetInstance().SessionSecret), nil
	})

	if err != nil {
		return false, err
	}

	// Check if the token is valid
	if token.Valid {
		// Token is valid, check the expiration time
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			expirationTime := time.Unix(int64(claims["exp"].(float64)), 0)
			currentTime := time.Now()

			if expirationTime.After(currentTime) {
				return true, nil
			} else {
				return false, errors.New("token has expired")
			}
		} else {
			return false, errors.New("error extracting claims")
		}
	} else {
		return false, errors.New("error parsing token")
	}
}

func (u *AuthMiddleware) GetHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {

		auth := strings.SplitN(c.Request.Header.Get("Authorization"), " ", 2)
		if len(auth) != 2 || auth[0] != "Basic" {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.BASIC_AUTH_FAILED, nil, nil))
		}

		payload, _ := base64.StdEncoding.DecodeString(auth[1])
		pair := strings.SplitN(string(payload), ":", 2)
		if len(pair) != 2 || !u.validate(pair[0], pair[1], c) {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.BASIC_AUTH_FAILED, nil, nil))
		} else {
			ok, err := u.isTokenValid(strings.TrimSpace(pair[1]))

			if ok && err == nil {
				c.Set("userId", strings.TrimSpace(pair[0]))
				c.Set("token", strings.TrimSpace(pair[1]))
				c.Next()
			} else {
				c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.TOKEN_EXPIRED, nil, nil))
			}
		}
	}
}

func (u *AuthMiddleware) validate(username string, password string, c *gin.Context) bool {
	//If it passes session middleware means session exists
	sessionGeneric, _ := c.Get("session")
	session := sessionGeneric.(models.Session)

	if session.Username == username && session.Token == password {
		c.Set("role", int(session.Role))
		c.Set("email", strings.ToLower(session.Email))
		c.Set("username", strings.ToLower(session.Username))
		c.Set("roleName", strings.ToLower(session.RoleName))
		return true
	}

	return false
}
