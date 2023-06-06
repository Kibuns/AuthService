package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

func TestGetJwtWithIncorrectCredentials(t *testing.T) {
	// Create a request body with incorrect credentials
	requestBody := map[string]string{
		"userName": "incorrectUser",
		"password": "incorrectPassword893789178956187936891356589136050890163581635789613597813875137865789136578915",
	}
	jsonBody, _ := json.Marshal(requestBody)

	// Create a new request with the JSON body
	req, err := http.NewRequest("POST", "/jwt", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatal(err)
	}

	// Set the request Content-Type header to application/json
	req.Header.Set("Content-Type", "application/json")

	// Create a response recorder to record the response
	rr := httptest.NewRecorder()

	// Create a new router
	router := mux.NewRouter()
	router.HandleFunc("/jwt", GetJwt)

	// Call the GetJwt handler function
	router.ServeHTTP(rr, req)

	// Check the response status code
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, but got %d", http.StatusUnauthorized, rr.Code)
	}

	// Check the response body message
	expectedBody := "incorrect credentials"
	if strings.TrimSpace(rr.Body.String()) != expectedBody {
		t.Errorf("Expected body %s, but got %s", expectedBody, rr.Body.String())
	}
}

func TestGetJwtWithCorrectCredentials(t *testing.T) {
	// Create a request body with incorrect credentials
	requestBody := map[string]string{
		"userName": "jaicy",
		"password": "1234",
	}
	jsonBody, _ := json.Marshal(requestBody)

	// Create a new request with the JSON body
	req, err := http.NewRequest("POST", "/jwt", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatal(err)
	}

	// Set the request Content-Type header to application/json
	req.Header.Set("Content-Type", "application/json")

	// Create a response recorder to record the response
	rr := httptest.NewRecorder()

	// Create a new router
	router := mux.NewRouter()
	router.HandleFunc("/jwt", GetJwt)

	// Call the GetJwt handler function
	router.ServeHTTP(rr, req)

	// Check the response status code
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, but got %d", http.StatusOK, rr.Code)
	}

	// Check the response body message
	expectedBody := "incorrect credentials"
	if strings.TrimSpace(rr.Body.String()) != expectedBody {
		t.Errorf("Expected body %s, but got %s", expectedBody, rr.Body.String())
	}
}

func TestEndpointWithInvalidJWT(t *testing.T) {
	// Create a new request without a valid JWT
	req, err := http.NewRequest("GET", "/api", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Set an invalid JWT token in the Authorization header
	req.Header.Set("Token", "invalid token lol")

	// Create a response recorder to record the response
	rr := httptest.NewRecorder()

	// Create a new router
	router := mux.NewRouter()
	router.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		// This is a mock implementation of the handler function for testing purposes
		if validateToken(req.Header.Get("Token")){
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusForbidden)
		}
		
		
		w.Write([]byte("Mock Handler"))
	}).Methods("GET")

	// Call the handler function
	router.ServeHTTP(rr, req)

	// Check the response status code
	if rr.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, but got %d", http.StatusForbidden, rr.Code)
	}
}

func TestEndpointWithValidJWT(t *testing.T) {
	// Create a new request without a valid JWT
	req, err := http.NewRequest("GET", "/api", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Set a valid JWT token in the Authorization header
	jwt, err := CreateJWT("testuser")
	if err != nil {
		return
	}
	req.Header.Set("Token", jwt)

	// Create a response recorder to record the response
	rr := httptest.NewRecorder()

	// Create a new router
	router := mux.NewRouter()
	router.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		// This is a mock implementation of the handler function for testing purposes
		if validateToken(req.Header.Get("Token")){
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusForbidden)
		}
		
		
		w.Write([]byte("Mock Handler"))
	}).Methods("GET")

	// Call the handler function
	router.ServeHTTP(rr, req)

	// Check the response status code
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, but got %d", http.StatusOK, rr.Code)
	}
}

func TestJWTValidationWithValidJWT(t *testing.T) {
	validToken, err := CreateJWT("testuser")
	if err != nil {
		t.Errorf("Error creating JWT token")
	}
	if(!validateToken(validToken)){
		t.Errorf("Token determined invalid according to validateToken")
	}
}

func TestJWTValidationWithInvalidJWT(t *testing.T) {
	invalidToken := "invalidtoken838165985613978657813"
	if(validateToken(invalidToken)){
		t.Errorf("Invalid token determined valid according to validateToken")
	}
}