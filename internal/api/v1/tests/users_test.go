// Package tests provides integration tests for the v1 API.
package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"interface-api/internal/api/v1/handlers"

	"github.com/labstack/echo/v4"
)

func TestGetUser(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
	resp := httptest.NewRecorder()
	c := e.NewContext(req, resp)
	c.SetParamNames("id")
	c.SetParamValues("123")

	h := handlers.NewUserHandler(nil)

	if err := h.GetUser(c); err != nil {
		t.Errorf("GetUser() error = %v", err)
		return
	}

	if resp.Code != http.StatusOK {
		t.Errorf("GetUser() wrong status code = %v", resp.Code)
		return
	}

	var actual map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&actual); err != nil {
		t.Errorf("GetUser() error decoding response body: %v", err)
		return
	}

	if actual["id"] != "123" {
		t.Errorf("GetUser() expected id 123, got %v", actual["id"])
	}
}
