package test

import (
	"chat-module/auth"
	"chat-module/util"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimit(t *testing.T) {
	go util.RateLimit()

	// Generate a JWT token, which we will be used for authentication
	token, err := auth.GenerateJWTToken("test-user")

	if err != nil {
		t.Fatalf("failed to generate jwt token: %v", err.Error())
	}

	// Create the request
	req, err := http.NewRequest("GET", "/api/test/success", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	for i := 0; i < util.MaximumRequests; i++ {
		// Use httptest to create a ResponseRecorder to record the response
		rr := httptest.NewRecorder()
		handler := http.Handler(util.RateLimitMiddleware(Test200ResponseHandler))

		log.Printf("Running request %d", i+1)

		// Call the handler function, passing the Request and ResponseRecorder
		handler.ServeHTTP(rr, req)

		// Check the status code of the response
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		// Check the response body
		expected := "TEST endpoint OK"
		if rr.Body.String() != expected {
			t.Errorf("handler returned unexpected body: got %v want %v",
				rr.Body.String(), expected)
		}
	}

	// Use httptest to create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()
	handler := http.Handler(util.RateLimitMiddleware(Test200ResponseHandler))

	// Next one should fail
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Rate limit not working, should have triggered after %d requests", util.MaximumRequests)
	}

	// Now wait the required time
	time.Sleep((util.TimeSlotSeconds + 1) * time.Second)

	rr = httptest.NewRecorder()
	handler = http.Handler(util.RateLimitMiddleware(Test200ResponseHandler))

	handler.ServeHTTP(rr, req)

	// Check the status code of the response
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body
	expected := "TEST endpoint OK"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}
