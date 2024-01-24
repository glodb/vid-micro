package controllers

import (
	"errors"
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
	"com.code.vidmicro/com.code.vidmicro/settings/s3uploader"
	"com.code.vidmicro/com.code.vidmicro/settings/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/xid"
)

type UserController struct {
	baseinterfaces.BaseControllerFactory
	basefunctions.BaseFucntionsInterface
	basevalidators.ValidatorInterface
}

func (u *UserController) generateJWT(minutes int64) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(time.Minute * time.Duration(minutes)).Unix(), // Set expiration time to 1 hour from now
	})

	// Sign the token with the secret key
	tokenString, err := token.SignedString([]byte(configmanager.GetInstance().SessionSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (u *UserController) generateRefreshToken() (string, error) {
	// Generate a random refresh token (you may use a library for better randomness)
	newXID := xid.New()
	return string(newXID.String()), nil
}

func (u *UserController) GetDBName() basetypes.DBName {
	return basetypes.DBName(configmanager.GetInstance().Database.DBName)
}

func (u *UserController) GetCollectionName() basetypes.CollectionName {
	return "users"
}

func (u *UserController) DoIndexing() error {
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

		file, err := c.FormFile("image")

		if err == nil && file != nil {
			url, err := s3uploader.GetInstance().UploadToSCW(file)
			if err == nil {
				modelUser.AvatarUrl = url
			} else {
				c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.UPLOADING_AVATAR_FAILED, err, nil))
				return
			}
		}

		err = u.Validate(c.GetString("apiPath"), modelUser)
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
		_, err = u.Add(u.GetDBName(), u.GetCollectionName(), modelUser, true)

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

		err := u.Validate(c.GetString("apiPath"), modelUser)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		rows, err := u.Find(u.GetDBName(), u.GetCollectionName(), "id, username,name, email, password, role, salt, avatar_url, createdAt, updatedAt", map[string]interface{}{"username": modelUser.Username, "email": modelUser.Email}, &modelUser, true, " AND is_verified=TRUE AND black_listed=FALSE", true)

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
				jwtToken, err := u.generateJWT(configmanager.GetInstance().TokenExpiry)

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
				session.RoleName = cache.GetInstance().HashGet("auth_roles_"+strconv.FormatInt(int64(session.Role), 10), "slug")

				cache.GetInstance().SAdd([]interface{}{strconv.FormatInt(int64(user.Id), 10) + "_all_sessions", sessionId})

				cache.GetInstance().Set(sessionId, session.EncodeRedisData())
				c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.LOGIN_SUCCESS, err, map[string]string{"jwtToken": jwtToken, "refreshToken": refreshToken, "username": user.Username, "tokenType": "HTTPBasicAuth"}))

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
		newJwtToken, err := u.generateJWT(configmanager.GetInstance().TokenExpiry)
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

func (u *UserController) handleBlackListUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelUser := models.User{}
		if err := c.ShouldBind(&modelUser); err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		err := u.Validate(c.GetString("apiPath"), modelUser)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		var currentSession models.Session
		if val, ok := c.Get("session"); ok {
			currentSession = val.(models.Session)
		}

		if currentSession.Username == modelUser.Username {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.FAILED_BLACK_LISTING, errors.New("can't black list your own user"), nil))
			return
		}

		sessions := cache.GetInstance().SMembers(strconv.FormatInt(int64(modelUser.Id), 10) + "_all_sessions")

		for _, sessionId := range sessions {
			data, err := cache.GetInstance().Get(sessionId)
			if err == nil || len(data) != 0 {
				var session models.Session
				session.DecodeRedisData(data)
				session.BlackListed = true
				cache.GetInstance().Set(sessionId, session.EncodeRedisData())
			}
		}

		data := []interface{}{true, modelUser.Id}

		err = u.UpdateOne(u.GetDBName(), u.GetCollectionName(), "UPDATE "+string(u.GetCollectionName())+" SET black_listed = $1 WHERE id = $2 ", data, false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.FAILED_BLACK_LISTING, err, nil))
			return
		}
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.BLACK_LIST_SUCCESS, err, nil))
	}
}

func (u *UserController) handleEditUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		setPart := " SET "
		data := make([]interface{}, 0)
		modelUser := models.User{}
		if err := c.ShouldBind(&modelUser); err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		err := u.Validate(c.GetString("apiPath"), modelUser)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, err, nil))
			return
		}

		var currentSession models.Session
		if val, ok := c.Get("session"); ok {
			currentSession = val.(models.Session)
		}

		file, err := c.FormFile("image")

		if err == nil && file != nil {
			url, err := s3uploader.GetInstance().UploadToSCW(file)
			if err == nil {

				if currentSession.AvatarUrl != url {
					modelUser.AvatarUrl = url
					currentSession.AvatarUrl = url
					setPart += "avatar_url = $1"
					data = append(data, url)
				}
			} else {
				c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.UPLOADING_AVATAR_FAILED, err, nil))
				return
			}
		}

		if modelUser.Name != "" {
			if currentSession.Name != modelUser.Name {
				lengthString := strconv.FormatInt(int64(len(data)+1), 10)
				if len(data) > 0 {
					setPart += ","
				}
				setPart += "name = $" + lengthString
				currentSession.Name = modelUser.Name
				data = append(data, modelUser.Name)
			}
		}

		cache.GetInstance().Set(currentSession.SessionId, currentSession.EncodeRedisData())

		if len(data) > 0 {
			lengthString := strconv.FormatInt(int64(len(data)+1), 10)
			setPart += " WHERE id =$" + lengthString
			data = append(data, currentSession.UserId)

			err = u.UpdateOne(u.GetDBName(), u.GetCollectionName(), "UPDATE "+string(u.GetCollectionName())+setPart, data, false)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.FAILED_UPDATING_USER, err, nil))
				return
			}
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.UPDATING_USER_SUCCESS, err, nil))
			return
		}
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.NOTHIN_TO_UPDATE, err, nil))
	}
}

func (u *UserController) RegisterApis() {
	baserouter.GetInstance().GetOpenRouter().POST("/api/signup", u.handleRegisterUser())
	baserouter.GetInstance().GetOpenRouter().POST("/api/login", u.handleLogin())
	baserouter.GetInstance().GetOpenRouter().POST("/api/refreshToken", u.handleRefreshToken())
	baserouter.GetInstance().GetLoginRouter().GET("/api/getUser", u.handleGetUser())
	baserouter.GetInstance().GetLoginRouter().POST("/api/blackListUser", u.handleBlackListUser())
	baserouter.GetInstance().GetLoginRouter().POST("/api/editUser", u.handleEditUser())
}
