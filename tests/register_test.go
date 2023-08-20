package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"com.code.sso/com.code.sso/app/controllers"
	"com.code.sso/com.code.sso/app/models"
	"com.code.sso/com.code.sso/config"
	"com.code.sso/com.code.sso/httpHandler/basecontrollers"
	"com.code.sso/com.code.sso/httpHandler/responses"
	"com.code.sso/tests/clientmodels"
)

func TestRegisterUserAPI(t *testing.T) {
	// Create a test server with the handler

	config.GetInstance().Setup("../setup/prod.json")

	controller, _ := basecontrollers.GetInstance().GetController("User")
	userController := controller.(*controllers.UserController)

	testCases := []struct {
		Name           string
		RequestData    map[string]string
		ExpectedStatus int
	}{
		{
			Name: "ValidRequest",
			RequestData: map[string]string{
				"email":    "test4@example.com",
				"password": "securepassword",
			},
			ExpectedStatus: responses.REGISTER_USER_SUCCESS,
		},
		{
			Name: "InvalidPasswordLength",
			RequestData: map[string]string{
				"email":    "test@example.com",
				"password": "short",
			},
			ExpectedStatus: responses.VALIDATION_FAILED,
		},
		{
			Name: "DuplicateUser",
			RequestData: map[string]string{
				"email":    "test4@example.com",
				"password": "securepassword",
			},
			ExpectedStatus: responses.DB_ERROR,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			requestJSON, _ := json.Marshal(tc.RequestData)
			// Create a mock request
			req, err := http.NewRequest(http.MethodPost, "/api/registerUser", bytes.NewBuffer(requestJSON))
			if err != nil {
				t.Fatal(err)
			}

			session := models.Session{
				SessionId: "rand-id",
			}
			ctx := context.WithValue(req.Context(), "session", session)
			// Create a ResponseRecorder to record the response
			rr := httptest.NewRecorder()

			// Call the handler function using the mock request
			userController.HandleRegisterUser(rr, req.WithContext(ctx))

			user := clientmodels.ClientUser{}
			decoder := json.NewDecoder(rr.Body)
			err = decoder.Decode(&user)

			if user.Code != tc.ExpectedStatus {
				t.Errorf("Expected status code %d, but got %d", tc.ExpectedStatus, user.Code)
			}
			log.Println(user)
		})
	}
}
