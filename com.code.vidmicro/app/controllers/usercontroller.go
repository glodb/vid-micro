package controllers

import (
	//"com.code.vidmicro/com.code.vidmicro/app/middlewares"
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"com.code.vidmicro/com.code.vidmicro/app/models"
	"com.code.vidmicro/com.code.vidmicro/app/models/jsonmodels"
	"com.code.vidmicro/com.code.vidmicro/database/basefunctions"
	"com.code.vidmicro/com.code.vidmicro/database/basetypes"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/basecontrollers/baseconst"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/basecontrollers/baseinterfaces"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/baserouter"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/basevalidators"
	"com.code.vidmicro/com.code.vidmicro/httpHandler/responses"
	"com.code.vidmicro/com.code.vidmicro/settings/cache"
	"com.code.vidmicro/com.code.vidmicro/settings/configmanager"
	"com.code.vidmicro/com.code.vidmicro/settings/emails"
	"com.code.vidmicro/com.code.vidmicro/settings/oauthconfig"
	"com.code.vidmicro/com.code.vidmicro/settings/oauthconfig/services"
	"com.code.vidmicro/com.code.vidmicro/settings/s3uploader"
	"com.code.vidmicro/com.code.vidmicro/settings/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
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
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, err, nil))
			return
		}

		file, err := c.FormFile("image")

		if err == nil && file != nil {
			url, statusCode, err := s3uploader.GetInstance().UploadToSCW(file)
			if err == nil {
				modelUser.AvatarUrl = url
			} else {
				c.AbortWithStatusJSON(statusCode, responses.GetInstance().WriteResponse(c, responses.UPLOADING_AVATAR_FAILED, err, nil))
				return
			}
		}

		err = basevalidators.GetInstance().GetValidator().Struct(modelUser)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, basevalidators.GetInstance().CreateErrors(err), nil))
			return
		}

		count, err := u.Count(u.GetDBName(), u.GetCollectionName(), map[string]interface{}{"username": modelUser.Username, "email": modelUser.Email}, true)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}

		if count > 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.USERNAME_OR_EMAIL_EXISTS, err, nil))
			return
		}

		verificationToken, err := u.generateJWT(configmanager.GetInstance().EmailVerificationTokenExpiry)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.GENERATE_EMAIL_VERIFICATION_TOKEN_FAILED, err, nil))
			return
		}

		modelUser.Salt, err = utils.GenerateSalt()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}
		modelUser.Role = 20
		modelUser.CreatedAt = time.Now()
		modelUser.UpdatedAt = time.Now()
		modelUser.Password = utils.HashPassword(modelUser.Password, modelUser.Salt)
		modelUser.VerificationToken = verificationToken

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}
		userId, err := u.Add(u.GetDBName(), u.GetCollectionName(), modelUser, true)
		modelUser.Id = int(userId)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}

		verificationURL := fmt.Sprintf("%s%s", configmanager.GetInstance().EmailVerificationURL, verificationToken)
		emailBody := fmt.Sprintf(configmanager.GetInstance().EmailBody, modelUser.Username, verificationURL)

		err = emails.GetInstance().SendVerificationEmail(modelUser.Email, configmanager.GetInstance().EmailSubject, emailBody)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}

		controller, _ := u.BaseControllerFactory.GetController(baseconst.RefreshToken)
		refreshTokenController := controller.(*RefreshTokensController)
		refreshToken, err := refreshTokenController.GetRefreshToken(modelUser.Id)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}

		jwtToken, err := u.generateJWT(configmanager.GetInstance().TokenExpiry)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}
		var session models.Session
		if val, ok := c.Get("session"); ok {
			session = val.(models.Session)
		}
		session.Username = modelUser.Username
		session.Token = jwtToken
		session.Name = modelUser.Name
		session.Email = modelUser.Email
		session.AvatarUrl = modelUser.AvatarUrl
		session.IsVerified = modelUser.IsVerified
		session.Salt = modelUser.Salt
		session.Role = modelUser.Role
		session.UserId = int64(modelUser.Id)
		session.RoleName = cache.GetInstance().HashGet("auth_roles_"+strconv.FormatInt(int64(session.Role), 10), "slug")

		userSessionController, _ := u.BaseControllerFactory.GetController(baseconst.UsersSessions)
		query := `INSERT INTO users_sessions (user_id, session_id, created_at, expiring_at)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (session_id) DO UPDATE
			SET user_id = $1;`
		_, err = userSessionController.RawQuery(userSessionController.GetDBName(), userSessionController.GetCollectionName(), query, []interface{}{modelUser.Id, session.SessionId, session.CreatedAt, session.ExpiringAt})

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}
		cache.GetInstance().Set(session.SessionId, session.EncodeRedisData())

		session.Salt = make([]byte, 0)
		session.Token = ""
		session.Password = ""
		session.SessionId = ""
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.REGISTER_USER_SUCCESS, err, map[string]interface{}{"jwtToken": jwtToken, "refreshToken": refreshToken, "username": modelUser.Username, "tokenType": "Bearer", "user": session}))

	}
}

func (u *UserController) handleVerifyEmail() gin.HandlerFunc {
	return func(c *gin.Context) {
		emailVerificationToken := c.Param("token")
		if emailVerificationToken == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.INVALID_EMAIL_OR_TOKEN, nil, nil))
			return
		}

		// Parse the JWT token to check for expiration
		claims := jwt.MapClaims{}
		parsedToken, err := jwt.ParseWithClaims(emailVerificationToken, claims, func(token *jwt.Token) (interface{}, error) {
			// Provide the key or secret used to sign the token
			return []byte(configmanager.GetInstance().SessionSecret), nil
		})

		// Check for errors during token parsing
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}

		// Check if the token is valid
		if parsedToken.Valid {
			// Check for expiration
			expirationDuration := time.Duration(configmanager.GetInstance().EmailVerificationTokenExpiry) * time.Second
			expirationTime := time.Now().Add(expirationDuration)
			if time.Now().After(expirationTime) {
				// Token has expired
				c.AbortWithStatusJSON(http.StatusUnauthorized, responses.GetInstance().WriteResponse(c, responses.TOKEN_EXPIRED, nil, nil))
				return
			}

			// If the tokens match, update the user's isVerified flag in the database
			err = u.UpdateOne(u.GetDBName(), u.GetCollectionName(), "UPDATE "+string(u.GetCollectionName())+" SET is_verified = true, verification_token = '' WHERE verification_token = $1", []interface{}{emailVerificationToken}, false)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
				return
			}

			// Respond with success message
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.EMAIL_VERIFICATION_SUCCESS, nil, nil))
		} else {
			// Token is invalid
			c.AbortWithStatusJSON(http.StatusUnauthorized, responses.GetInstance().WriteResponse(c, responses.INVALID_EMAIL_OR_TOKEN, nil, nil))
		}
	}
}

func (u *UserController) handleGoogleLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		configSet, err := oauthconfig.GetInstance().GetOAuth2Config(services.Google)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}

		sessionId := ""
		if val, ok := c.Get("session-id"); ok {
			sessionId = val.(string)
		}
		url := configSet.AuthCodeURL(sessionId)
		log.Println("Generated URL:", url)
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.URL_GENERATED, err, map[string]interface{}{"redirect_url": url}))
		// c.Redirect(http.StatusSeeOther, url)
	}
}

// 3. Handle Callback
func (u *UserController) handleGoogleCallback() gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Query("code")
		sessionId := c.Request.FormValue("state")

		var session models.Session
		data, err := cache.GetInstance().Get(sessionId)
		if err != nil || len(data) == 0 {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		} else {
			session.DecodeRedisData(data)
		}

		configSet, err := oauthconfig.GetInstance().GetOAuth2Config(services.Google)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}
		token, err := configSet.Exchange(context.Background(), code)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.GOOGLE_LOGIN_FAILED, err, nil))
			return
		}

		// 4. Exchange the token for user information
		client := configSet.Client(context.TODO(), token)
		resp, err := client.Get(configmanager.GetInstance().GoogleUserInfoLink)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.GOOGLE_LOGIN_FAILED, err, nil))
			return
		}
		defer resp.Body.Close()

		var userInfo map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.GOOGLE_LOGIN_FAILED, err, nil))
			return
		}

		log.Print(userInfo)

		// Construct the upsert query
		upsertQuery := `
		INSERT INTO users (email, username, name, avatar_url, password, is_verified, role)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (email) DO UPDATE
		SET 
			username = $2, 
			avatar_url = $3, 
			is_verified = COALESCE(EXCLUDED.is_verified, users.is_verified);
		`
		userInfo["role"] = 20
		// Execute the upsert query using RawQuery
		_, err = u.RawQuery(u.GetDBName(), u.GetCollectionName(), upsertQuery, []interface{}{userInfo["email"], userInfo["name"], userInfo["name"], userInfo["picture"], userInfo["id"], true, userInfo["role"]})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.GOOGLE_LOGIN_FAILED, err, nil))
			return
		}

		// Retrieve the user from the database
		var users []models.User
		keys := "id, username, name, email, avatar_url, is_verified, salt, role, createdAt, updatedAt"
		condition := map[string]interface{}{"email": userInfo["email"]}
		rows, err := u.Find(u.GetDBName(), u.GetCollectionName(), keys, condition, &users, true, "", true)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.GOOGLE_LOGIN_FAILED, err, nil))
			return
		}
		defer rows.Close()

		// Iterate over the rows.
		for rows.Next() {
			// Create a User struct to scan values into.
			var user models.User

			// Scan the row's values into the User struct.
			avatarUrl := sql.NullString{}
			err := rows.Scan(&user.Id, &user.Username, &user.Name, &user.Email, &avatarUrl, &user.IsVerified, &user.Salt, &user.Role, &user.CreatedAt, &user.UpdatedAt)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.GOOGLE_LOGIN_FAILED, err, nil))
				return
			}

			user.AvatarUrl = avatarUrl.String

			// Append the user to the slice.
			users = append(users, user)
		}

		// Check for errors from iterating over rows.
		if err := rows.Err(); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.GOOGLE_LOGIN_FAILED, err, nil))
			return
		}

		if len(users) >= 1 {
			user := users[0]

			// Generate JWT for the user
			jwtToken, err := u.generateJWT(configmanager.GetInstance().TokenExpiry)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.GOOGLE_LOGIN_FAILED, err, nil))
				return
			}

			// Generate Refresh Token
			controller, _ := u.BaseControllerFactory.GetController(baseconst.RefreshToken)
			refreshTokenController := controller.(*RefreshTokensController)
			refreshToken, err := refreshTokenController.GetRefreshToken(user.Id)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.GOOGLE_LOGIN_FAILED, err, nil))
				return
			}

			session.Username = user.Username
			session.Token = jwtToken
			session.Name = user.Name
			session.Email = user.Email
			session.AvatarUrl = user.AvatarUrl
			session.IsVerified = user.IsVerified
			session.Salt = user.Salt
			session.Role = user.Role
			session.UserId = int64(user.Id)
			session.RoleName = cache.GetInstance().HashGet("auth_roles_"+strconv.FormatInt(int64(session.Role), 10), "slug")

			userSessionController, _ := u.BaseControllerFactory.GetController(baseconst.UsersSessions)
			query := `INSERT INTO users_sessions (user_id, session_id, created_at, expiring_at)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (session_id) DO UPDATE
			SET user_id = $1;`
			_, err = userSessionController.RawQuery(userSessionController.GetDBName(), userSessionController.GetCollectionName(), query, []interface{}{user.Id, session.SessionId, session.CreatedAt, session.ExpiringAt})

			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
				return
			}

			cache.GetInstance().Set(sessionId, session.EncodeRedisData())

			session.Salt = make([]byte, 0)
			session.Token = ""
			session.Password = ""
			session.SessionId = ""
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.LOGIN_SUCCESS, err, map[string]interface{}{"jwtToken": jwtToken, "refreshToken": refreshToken, "username": user.Username, "tokenType": "Bearer", "user": session}))
		} else {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.GOOGLE_LOGIN_FAILED, errors.New("user not found in database"), nil))
		}
	}
}

func (u *UserController) handleLogin() gin.HandlerFunc {
	return func(c *gin.Context) {

		modelUser := jsonmodels.Login{}
		if err := c.ShouldBindJSON(&modelUser); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, err, nil))
			return
		}

		err := basevalidators.GetInstance().GetValidator().Struct(modelUser)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, basevalidators.GetInstance().CreateErrors(err), nil))
			return
		}

		appendingQuery := " AND is_verified=TRUE AND black_listed=FALSE"
		if configmanager.GetInstance().AllowUnverified {
			appendingQuery = " AND black_listed=FALSE"
		}

		rows, err := u.Find(u.GetDBName(), u.GetCollectionName(), "id, username,name, email, password, role, salt, avatar_url, createdAt, updatedAt", map[string]interface{}{"username": modelUser.Username, "email": modelUser.Email}, &modelUser, true, appendingQuery, true)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}
		defer rows.Close()

		var users []models.User

		// Iterate over the rows.
		for rows.Next() {
			// Create a User struct to scan values into.
			var user models.User

			avatarUrl := sql.NullString{}
			// Scan the row's values into the User struct.
			err := rows.Scan(&user.Id, &user.Username, &user.Name, &user.Email, &user.Password, &user.Role, &user.Salt, &avatarUrl, &user.CreatedAt, &user.UpdatedAt)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
				return
			}
			user.AvatarUrl = avatarUrl.String

			// Append the user to the slice.
			users = append(users, user)
		}

		// Check for errors from iterating over rows.
		if err := rows.Err(); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}

		if len(users) >= 1 { //Check the first user as username and email are unique
			user := users[0]
			password := utils.HashPassword(modelUser.Password, user.Salt)

			if password == user.Password { //
				jwtToken, err := u.generateJWT(configmanager.GetInstance().TokenExpiry)

				if err != nil {
					c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
					return
				}

				controller, _ := u.BaseControllerFactory.GetController(baseconst.RefreshToken)
				refreshTokenController := controller.(*RefreshTokensController)
				refreshToken, err := refreshTokenController.GetRefreshToken(user.Id)
				if err != nil {
					c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
					return
				}
				err = cache.GetInstance().SetString(refreshToken, jwtToken)
				if err != nil {
					c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
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
				session.Username = user.Username
				session.Token = jwtToken
				session.Name = user.Name
				session.Email = user.Email
				session.Password = user.Password
				session.AvatarUrl = user.AvatarUrl
				session.IsVerified = true
				session.Salt = user.Salt
				session.Role = user.Role
				session.UserId = int64(user.Id)
				session.RoleName = cache.GetInstance().HashGet("auth_roles_"+strconv.FormatInt(int64(session.Role), 10), "slug")

				userSessionController, _ := u.BaseControllerFactory.GetController(baseconst.UsersSessions)
				query := `INSERT INTO users_sessions (user_id, session_id, created_at, expiring_at)
						VALUES ($1, $2, $3, $4)
						ON CONFLICT (session_id) DO UPDATE
						SET user_id = $1;`
				_, err = userSessionController.RawQuery(userSessionController.GetDBName(), userSessionController.GetCollectionName(), query, []interface{}{user.Id, session.SessionId, session.CreatedAt, session.ExpiringAt})

				if err != nil {
					c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
					return
				}

				err = cache.GetInstance().Set(sessionId, session.EncodeRedisData())

				if err != nil {
					c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
					return
				}

				session.Salt = make([]byte, 0)
				session.Token = ""
				session.Password = ""
				session.SessionId = ""
				c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.LOGIN_SUCCESS, err, map[string]interface{}{"jwtToken": jwtToken, "refreshToken": refreshToken, "username": user.Username, "tokenType": "Bearer", "user": session}))

			} else {
				c.AbortWithStatusJSON(http.StatusUnauthorized, responses.GetInstance().WriteResponse(c, responses.PASSWORD_MISMATCHED, err, nil))
				return
			}
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, responses.GetInstance().WriteResponse(c, responses.ERROR_READING_USER, err, nil))
			return
		}
	}
}

func (u *UserController) handleRefreshToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse the refresh token from the request
		refreshToken := c.PostForm("refresh_token")
		if refreshToken == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.REFRESH_TOKEN_REQUIRED, nil, nil))
			return
		}

		controller, _ := u.BaseControllerFactory.GetController(baseconst.RefreshToken)
		refreshTokenController := controller.(*RefreshTokensController)
		validated, err := refreshTokenController.ValidateRefreshToken(refreshToken)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, nil, nil))
			return
		}

		if !validated {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, errors.New("refresh token doesn't exists"), nil))
			return
		}

		newJwtToken, err := u.generateJWT(configmanager.GetInstance().TokenExpiry)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, nil, nil))
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

		err = cache.GetInstance().Set(sessionId, session.EncodeRedisData())
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}

		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.REFRESH_TOKEN_SUCCESS, err, map[string]string{"jwtToken": newJwtToken, "refreshToken": refreshToken, "username": session.Username, "tokenType": "Bearer"}))
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
		modelUser := jsonmodels.BlackListUser{}
		if err := c.ShouldBind(&modelUser); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, err, nil))
			return
		}

		err := basevalidators.GetInstance().GetValidator().Struct(modelUser)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, basevalidators.GetInstance().CreateErrors(err), nil))
			return
		}

		var currentSession models.Session
		if val, ok := c.Get("session"); ok {
			currentSession = val.(models.Session)
		}

		if int(currentSession.UserId) == modelUser.Id {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.FAILED_BLACK_LISTING, errors.New("can't black list your own user"), nil))
			return
		}

		userSessionsController, _ := u.BaseControllerFactory.GetController(baseconst.UsersSessions)
		rows, err := userSessionsController.Find(userSessionsController.GetDBName(), userSessionsController.GetCollectionName(), " session_id", map[string]interface{}{"user_id": modelUser.Id}, &models.UserSessions{}, false, "", false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}

		defer rows.Close()

		var sessionId string
		for rows.Next() {
			err := rows.Scan(&sessionId)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
				return
			}
			data, err := cache.GetInstance().Get(sessionId)
			if err == nil || len(data) != 0 {
				var session models.Session
				session.DecodeRedisData(data)
				session.BlackListed = true
				err = cache.GetInstance().Set(sessionId, session.EncodeRedisData())
				if err != nil {
					c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
					return
				}
			}
		}

		data := []interface{}{true, modelUser.Id}

		err = u.UpdateOne(u.GetDBName(), u.GetCollectionName(), "UPDATE "+string(u.GetCollectionName())+" SET black_listed = $1 WHERE id = $2 ", data, false)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.BLACK_LIST_SUCCESS, err, nil))
	}
}

func (u *UserController) handleEditUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		setPart := " SET "
		data := make([]interface{}, 0)
		modelUser := jsonmodels.EditUser{}
		if err := c.ShouldBind(&modelUser); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, err, nil))
			return
		}

		err := basevalidators.GetInstance().GetValidator().Struct(modelUser)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, basevalidators.GetInstance().CreateErrors(err), nil))
			return
		}

		var currentSession models.Session
		if val, ok := c.Get("session"); ok {
			currentSession = val.(models.Session)
		}

		file, err := c.FormFile("image")

		if err == nil && file != nil {
			url, statusCode, err := s3uploader.GetInstance().UploadToSCW(file)
			if err == nil {

				if currentSession.AvatarUrl != url {
					modelUser.AvatarUrl.String = url
					modelUser.AvatarUrl.Valid = true
					currentSession.AvatarUrl = url
					setPart += "avatar_url = $1"
					data = append(data, url)
				}
			} else {
				c.AbortWithStatusJSON(statusCode, responses.GetInstance().WriteResponse(c, responses.UPLOADING_AVATAR_FAILED, err, nil))
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

		err = cache.GetInstance().Set(currentSession.SessionId, currentSession.EncodeRedisData())
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}

		if len(data) > 0 {
			lengthString := strconv.FormatInt(int64(len(data)+1), 10)
			setPart += " WHERE id =$" + lengthString
			data = append(data, currentSession.UserId)

			err = u.UpdateOne(u.GetDBName(), u.GetCollectionName(), "UPDATE "+string(u.GetCollectionName())+setPart, data, false)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
				return
			}
			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.UPDATING_USER_SUCCESS, err, nil))
			return
		}
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.NOTHIN_TO_UPDATE, err, nil))
	}
}
func (u *UserController) handleResetPassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		modelPassword := jsonmodels.ResetPassword{}
		if err := c.ShouldBind(&modelPassword); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, err, nil))
			return
		}

		err := basevalidators.GetInstance().GetValidator().Struct(modelPassword)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, basevalidators.GetInstance().CreateErrors(err), nil))
			return
		}
		var user models.User
		appendingQuery := " AND is_verified=TRUE AND black_listed=FALSE"
		if configmanager.GetInstance().AllowUnverified {
			appendingQuery = " AND black_listed=FALSE"
		}
		rows, err := u.Find(u.GetDBName(), u.GetCollectionName(), "id, username,name, email, password, password_hash, role, salt", map[string]interface{}{"email": modelPassword.Email}, &models.User{}, true, appendingQuery, true)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}
		defer rows.Close()

		if rows.Next() {
			passwordHash := sql.NullString{}
			err := rows.Scan(&user.Id, &user.Username, &user.Name, &user.Email, &user.Password, &passwordHash, &user.Role, &user.Salt)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
				return
			}
			user.PasswordHash = passwordHash.String

		} else {
			// Handle the case when there are no rows
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.NOT_VARIFIED_USER, err, nil))
			return
		}
		fmt.Println("check passwordHash from db: ", user.PasswordHash)

		if user.PasswordHash != "" {
			ok, _ := utils.IsTokenValid(user.PasswordHash)
			if ok {
				c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.TOKEN_ALREADY_SENT, err, nil))
				return
			}
			jwtToken, err := u.generateJWT(configmanager.GetInstance().PasswordTokenExpiry)

			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, err, nil))
				return
			}

			emailBody := fmt.Sprintf(configmanager.GetInstance().ResetPasswordEmailBody, user.Email, jwtToken)

			err = emails.GetInstance().SendVerificationEmail(user.Email, configmanager.GetInstance().ResetPasswordEmailSubject, emailBody)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.SEND_VERIFICATION_EMAIL_FAILED, err, nil))
				return
			}

			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.TOKEN_SENT_VIA_EMAIL, err, nil))

		} else {

			jwtToken, err := u.generateJWT(configmanager.GetInstance().PasswordTokenExpiry)

			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadGateway, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, err, nil))
				return
			}

			err = u.UpdateOne(u.GetDBName(), u.GetCollectionName(), "UPDATE "+string(u.GetCollectionName())+" SET password_hash=$1 WHERE email=$2", []interface{}{jwtToken, user.Email}, true)

			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
				return
			}
			emailBody := fmt.Sprintf(configmanager.GetInstance().ResetPasswordEmailBody, user.Email, jwtToken)

			err = emails.GetInstance().SendVerificationEmail(user.Email, configmanager.GetInstance().ResetPasswordEmailSubject, emailBody)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SEND_VERIFICATION_EMAIL_FAILED, err, nil))
				return
			}

			c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.TOKEN_SENT_VIA_EMAIL, err, nil))
		}

	}

}

func (u *UserController) handleVerifyPasswordHash() gin.HandlerFunc {
	return func(c *gin.Context) {

		modelUser := jsonmodels.VerifyPassword{}
		if err := c.ShouldBind(&modelUser); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.BAD_REQUEST, err, nil))
			return
		}

		err := basevalidators.GetInstance().GetValidator().Struct(modelUser)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.VALIDATION_FAILED, basevalidators.GetInstance().CreateErrors(err), nil))
			return
		}

		appendingQuery := " AND is_verified=TRUE AND black_listed=FALSE"
		if configmanager.GetInstance().AllowUnverified {
			appendingQuery = " AND black_listed=FALSE"
		}
		rows, err := u.Find(u.GetDBName(), u.GetCollectionName(), "id, username,name, email, password, password_hash, role, salt", map[string]interface{}{"password_hash": modelUser.PasswordHash}, &models.User{}, true, appendingQuery, true)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}

		var user models.User
		if rows.Next() {
			passwordHash := sql.NullString{}
			err := rows.Scan(&user.Id, &user.Username, &user.Name, &user.Email, &user.Password, &passwordHash, &user.Role, &user.Salt)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
				return
			}
			user.PasswordHash = passwordHash.String

		} else {
			// Handle the case when there are no rows
			c.AbortWithStatusJSON(http.StatusBadRequest, responses.GetInstance().WriteResponse(c, responses.NOT_VARIFIED_USER, err, nil))
			return
		}

		user.Password = utils.HashPassword(modelUser.NewPassword, user.Salt)
		user.PasswordHash = ""

		err = u.UpdateOne(u.GetDBName(), u.GetCollectionName(), "UPDATE "+string(u.GetCollectionName())+" SET password = $1, password_hash=$2", []interface{}{user.Password, user.PasswordHash}, false)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.TOKEN_VERIFICTION_SUCCESS, err, nil))

	}
}

func (u *UserController) handleLogoutUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionId := c.GetString("session-id")
		err := cache.GetInstance().Del(sessionId)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}
		c.SetCookie(configmanager.GetInstance().CookieName, sessionId, -1, configmanager.GetInstance().CookiePath, configmanager.GetInstance().CookieDomain, false, true)

		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.LOGOUT_SUCCESS, err, nil))
	}
}

func (u *UserController) generateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}

func (u *UserController) handleTwitterLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		configSet, err := oauthconfig.GetInstance().GetOAuth2Config(services.Twitter)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}

		sessionId := ""
		if val, ok := c.Get("session-id"); ok {
			sessionId = val.(string)
		}

		// codeChallenge, err := u.generateRandomString(43) // 43 bytes base64 URL encoded = 32 bytes random + padding
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
			return
		}

		// Build the authorization URL with the generated state and code challenge
		// url := configSet.AuthCodeURL(sessionId) // oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		// oauth2.SetAuthURLParam("code_challenge_method", "plain"),

		url := configSet.AuthCodeURL(sessionId,
			oauth2.SetAuthURLParam("code_challenge", "codeChallenge"),
			oauth2.SetAuthURLParam("code_challenge_method", "plain"),
		)

		log.Println("Generated URL:", url)
		c.AbortWithStatusJSON(http.StatusOK, responses.GetInstance().WriteResponse(c, responses.URL_GENERATED, err, map[string]interface{}{"redirect_url": url}))
	}
}

func (u *UserController) handleTwitterCallback() gin.HandlerFunc {
	return func(c *gin.Context) {
		state := c.Query("state")
		if state == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state parameter"})
			return
		}

		code := c.Query("code")

		authString := configmanager.GetInstance().TwitterLoginConfig.ClientId + ":" + configmanager.GetInstance().TwitterLoginConfig.ClientSecret
		basicAuth := base64.StdEncoding.EncodeToString([]byte(authString))

		data := url.Values{}
		data.Set("code", code)
		data.Set("grant_type", "authorization_code")
		data.Set("client_id", configmanager.GetInstance().TwitterLoginConfig.ClientId)
		data.Set("redirect_uri", configmanager.GetInstance().TwitterLoginConfig.RedirectUrl)
		data.Set("code_verifier", "codeChallenge")

		// Create a new request
		req, err := http.NewRequest("POST", "https://api.twitter.com/2/oauth2/token", bytes.NewBufferString(data.Encode()))
		if err != nil {
			fmt.Println("Error creating request:", err)
			return
		}

		// Set the Content-Type header
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Authorization", "Basic "+basicAuth)

		// Send the request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error sending request:", err)
			return
		}
		defer resp.Body.Close()

		// Read the response
		fmt.Println("Response status:", resp.Status)
		// Print the response body
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		fmt.Println("Response body:", buf.String())

		// authCode := c.Query("auth_code")
		// configSet, err := oauthconfig.GetInstance().GetOAuth2Config(services.Twitter)
		// if err != nil {
		// 	c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.SERVER_ERROR, err, nil))
		// 	return
		// }
		// log.Println(authCode)
		// token, err := configSet.Exchange(context.Background(), code)
		// if err != nil {
		// 	c.AbortWithStatusJSON(http.StatusInternalServerError, responses.GetInstance().WriteResponse(c, responses.GOOGLE_LOGIN_FAILED, err, nil))
		// 	return
		// }

	}
}

func (u *UserController) RegisterApis() {
	baserouter.GetInstance().GetOpenRouter().POST("/api/signup", u.handleRegisterUser())
	baserouter.GetInstance().GetOpenRouter().POST("/api/login", u.handleLogin())
	baserouter.GetInstance().GetOpenRouter().POST("/api/refreshToken", u.handleRefreshToken())
	baserouter.GetInstance().GetLoginRouter().GET("/api/getUser", u.handleGetUser())
	baserouter.GetInstance().GetLoginRouter().POST("/api/blackListUser", u.handleBlackListUser())
	baserouter.GetInstance().GetLoginRouter().POST("/api/editUser", u.handleEditUser())
	baserouter.GetInstance().GetLoginRouter().GET("/api/logout", u.handleLogoutUser())
	baserouter.GetInstance().GetBaseRouter(configmanager.GetInstance().SessionKey).GET("/api/verifyEmail/:token", u.handleVerifyEmail())
	baserouter.GetInstance().GetBaseRouter(configmanager.GetInstance().SessionKey).POST("/api/resetPassword", u.handleResetPassword())
	baserouter.GetInstance().GetBaseRouter(configmanager.GetInstance().SessionKey).POST("/api/verifyPasswordHash", u.handleVerifyPasswordHash())
	baserouter.GetInstance().GetOpenRouter().GET("/api/googleLogin", u.handleGoogleLogin())
	baserouter.GetInstance().GetBaseRouter(configmanager.GetInstance().SessionKey).GET("/api/googleCallback", u.handleGoogleCallback())
	baserouter.GetInstance().GetOpenRouter().GET("/api/twitterLogin", u.handleTwitterLogin())
	baserouter.GetInstance().GetBaseRouter(configmanager.GetInstance().SessionKey).GET("/api/twitterCallback", u.handleTwitterCallback())
}
