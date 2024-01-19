package controllers

import (
	"crypto/rand"
	"io"
	"net/http"
	"strconv"
	"time"

	"com.code.vidmicro/com.code.vidmicro/app/models"
	"com.code.vidmicro/com.code.vidmicro/database/basefunctions"
	"com.code.vidmicro/com.code.vidmicro/database/basetypes"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/basecontrollers/baseinterfaces"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/baserouter"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/basevalidators"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/responses"
	"com.code.vidmicro/com.code.vidmicro/settings/cache"
	"com.code.vidmicro/com.code.vidmicro/settings/configmanager"
	"com.code.vidmicro/com.code.vidmicro/settings/utils"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid"
)

type UserController struct {
	baseinterfaces.BaseControllerFactory
	basefunctions.BaseFucntionsInterface
	basevalidators.ValidatorInterface
	entropy io.Reader
}

func (u *UserController) generateJWT(user models.User) (string, error) {
	claims := models.Claims{
		User: user,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Duration(configmanager.GetInstance().TokenExpiry) * time.Minute).Unix(), // Set expiration time
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(configmanager.GetInstance().SessionSecret))
}

func (u *UserController) generateRefreshToken() (string, error) {
	// Generate a random refresh token (you may use a library for better randomness)
	ulid := ulid.MustNew(ulid.Timestamp(time.Now()), u.entropy)
	return string(ulid.String()), nil
}

func (u *UserController) GetDBName() basetypes.DBName {
	return basetypes.DBName(configmanager.GetInstance().Database.DBName)
}

func (u *UserController) GetCollectionName() basetypes.CollectionName {
	return "users"
}

func (u *UserController) DoIndexing() error {
	u.entropy = ulid.Monotonic(rand.Reader, 0)
	u.EnsureIndex(u.GetDBName(), u.GetCollectionName(), models.User{})
	return nil
}

func (u *UserController) SetBaseFunctions(inter basefunctions.BaseFucntionsInterface) {
	u.BaseFucntionsInterface = inter
}

func (u *UserController) handleRegisterUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelUser := models.User{}
		if err := c.ShouldBind(&modelUser); err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		file, _, err := c.Request.FormFile("image")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}
		defer file.Close()

		err = u.Validate(c.GetString("path"), modelUser)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		modelUser.Salt, err = utils.GenerateSalt()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}
		modelUser.Role = 1
		modelUser.CreatedAt = time.Now()
		modelUser.UpdatedAt = time.Now()
		modelUser.Password = utils.HashPassword(modelUser.Password, modelUser.Salt)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.CREATE_HASH_FAILED, err, nil))
			return
		}
		_, err = u.Add(u.GetDBName(), u.GetCollectionName(), modelUser)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.CREATE_HASH_FAILED, err, nil))
			return
		}
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.REGISTER_USER_SUCCESS, err, nil))
	}
}

func (u *UserController) handleLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelUser := models.User{}
		if err := c.ShouldBindJSON(&modelUser); err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		err := u.Validate(c.GetString("path"), modelUser)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		rows, err := u.FindOne(u.GetDBName(), u.GetCollectionName(), "id, username,name, email, password, role, salt, avatar_url, createdAt, updatedAt", map[string]interface{}{"username": modelUser.Username, "email": modelUser.Email}, &modelUser, true, " AND is_verified=TRUE AND black_listed=FALSE", true)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.ERROR_READING_USER, err, nil))
			return
		}
		defer rows.Close()

		var users []models.User

		// Iterate over the rows.
		for rows.Next() {
			// Create a User struct to scan values into.
			var user models.User

			// Scan the row's values into the User struct.
			err := rows.Scan(&user.Id, &user.Username, &user.Name, &user.Email, &user.Password, &user.Role, &user.Salt, &user.AvatarUrl, &user.CreatedAt, &user.UpdatedAt)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.ERROR_READING_USER, err, nil))
				return
			}

			// Append the user to the slice.
			users = append(users, user)
		}

		// Check for errors from iterating over rows.
		if err := rows.Err(); err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.ERROR_READING_USER, err, nil))
			return
		}

		if len(users) >= 1 { //Check the first user as username and email are unique
			user := users[0]
			password := utils.HashPassword(modelUser.Password, user.Salt)

			if password == user.Password {
				jwtToken, err := u.generateJWT(models.User{})

				if err != nil {
					c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.ERROR_READING_USER, err, nil))
					return
				}

				refreshToken, err := u.generateRefreshToken()

				if err != nil {
					c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.ERROR_READING_USER, err, nil))
					return
				}
				cache.GetInstance().SetString(refreshToken, jwtToken)
				var session models.Session
				var sessionId string
				if val, ok := c.Get("session"); ok {
					session = val.(models.Session)
				}
				if val, ok := c.Get("session-id"); ok {
					sessionId = val.(string)
				}
				session.Username = user.Username
				session.Token = jwtToken
				session.Name = user.Name
				session.Email = user.Email
				session.Password = user.Password
				session.AvatarUrl = user.AvatarUrl
				session.IsVerified = true
				session.Salt = user.Salt
				session.Role = user.Role
				session.CreatedAt = user.CreatedAt
				session.UpdatedAt = user.UpdatedAt
				session.UserId = int64(user.Id)

				//TODO: get the role name from db

				cache.GetInstance().SAdd([]interface{}{strconv.FormatInt(int64(user.Id), 10) + "_all_sessions", sessionId})

				cache.GetInstance().Set(sessionId, session.EncodeRedisData())
				c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.LOGIN_SUCCESS, err, map[string]string{"jwtToken": jwtToken, "refreshToken": refreshToken, "username": modelUser.Username, "tokenType": "HTTPBasicAuth"}))

			} else {
				c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.PASSWORD_MISMATCHED, err, nil))
				return
			}
		} else {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.ERROR_READING_USER, err, nil))
			return
		}
	}
}

func (u *UserController) handleRefreshToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse the refresh token from the request
		refreshToken := c.PostForm("refresh_token")
		if refreshToken == "" {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.REFRESH_TOKEN_REQUIRED, nil, nil))
			return
		}
		jwtToken, err := cache.GetInstance().GetString(refreshToken)
		if jwtToken == "" || err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.INVALID_REFRESH_TOKEN, nil, nil))
			return
		}
		newJwtToken, err := u.generateJWT(models.User{})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.ERROR_CREATING_JWT, nil, nil))
			return
		}

		var session models.Session
		var sessionId string
		if val, ok := c.Get("session"); ok {
			session = val.(models.Session)
		}
		if val, ok := c.Get("session-id"); ok {
			sessionId = val.(string)
		}

		session.Token = newJwtToken

		cache.GetInstance().Set(sessionId, session.EncodeRedisData())

		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.REFRESH_TOKEN_SUCCESS, err, map[string]string{"jwtToken": newJwtToken, "refreshToken": refreshToken, "username": session.Username, "tokenType": "HTTPBasicAuth"}))
	}
}

func (u *UserController) handleGetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var session models.Session
		if val, ok := c.Get("session"); ok {
			session = val.(models.Session)
		}

		session.Salt = make([]byte, 0)
		session.Token = ""
		session.Password = ""
		session.SessionId = ""

		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.GET_USER_SUCCESS, nil, session))
	}
}

func (u *UserController) RegisterApis() {
	baserouter.GetInstance().GetOpenRouter().POST("/api/signup", u.handleRegisterUser())
	baserouter.GetInstance().GetOpenRouter().POST("/api/login", u.handleLogin())
	baserouter.GetInstance().GetOpenRouter().POST("/api/refreshToken", u.handleRefreshToken())
	baserouter.GetInstance().GetLoginRouter().GET("/api/getUser", u.handleGetUser())
}
