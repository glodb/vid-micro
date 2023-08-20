package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"com.code.sso/com.code.sso/app/controllers"
	"com.code.sso/com.code.sso/httpHandler/basecontrollers"
)

func TestRegisterUserAPI(t *testing.T) {
	// Create a test server with the handler
	controller, _ := basecontrollers.GetInstance().GetController("User")
	userController := controller.(*controllers.UserController)
	ts := httptest.NewServer(http.HandlerFunc(userController.HandleRegisterUser))
	defer ts.Close()

	// Define test cases
	testCases := []struct {
		Name           string
		RequestData    map[string]string
		ExpectedStatus int
	}{
		{
			Name: "ValidRequest",
			RequestData: map[string]string{
				"email":    "test@example.com",
				"password": "securepassword",
			},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name: "InvalidPasswordLength",
			RequestData: map[string]string{
				"email":    "test@example.com",
				"password": "short",
			},
			ExpectedStatus: http.StatusBadRequest,
		},
		// Add more test cases here
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Prepare request JSON
			requestJSON, _ := json.Marshal(tc.RequestData)

			// Make a POST request to the test server
			resp, err := http.Post(ts.URL, "application/json", bytes.NewBuffer(requestJSON))
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			// Check response status code
			if resp.StatusCode != tc.ExpectedStatus {
				t.Errorf("Expected status code %d, but got %d", tc.ExpectedStatus, resp.StatusCode)
			}
		})
	}
}
