package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestUserSessionWithoutCookieIsAnonymous(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewUserAuthHandler(nil, false)
	router := gin.New()
	router.GET("/api/v1/session", handler.Session)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/session", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d want=%d body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	var response struct {
		Success bool `json:"success"`
		Data    struct {
			Authenticated bool             `json:"authenticated"`
			User          *json.RawMessage `json:"user"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !response.Success {
		t.Fatal("success=false want true")
	}
	if response.Data.Authenticated {
		t.Fatal("authenticated=true want false")
	}
	if response.Data.User != nil {
		t.Fatalf("user=%s want null", string(*response.Data.User))
	}
}
