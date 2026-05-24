package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/normastars/frame"
)

// newTestApp creates an App with the example config for testing.
func newTestApp() *frame.App {
	return frame.New("./conf/default.yaml")
}

func TestHello(t *testing.T) {
	app := newTestApp()
	app.GET("/hello", Hello)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/hello", nil)
	app.Engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp frame.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Code != "0" {
		t.Errorf("expected code 0, got %s", resp.Code)
	}
}

func TestGetUser(t *testing.T) {
	app := newTestApp()
	app.GET("/user/:id", GetUser)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/user/1", nil)
	app.Engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestCreateUser(t *testing.T) {
	app := newTestApp()
	app.POST("/user", CreateUser)

	body := `{"name":"alice"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/user", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	app.Engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestCreateUserInvalidBody(t *testing.T) {
	app := newTestApp()
	app.POST("/user", CreateUser)

	body := `{invalid json}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/user", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	app.Engine.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestListItems(t *testing.T) {
	app := newTestApp()
	app.GET("/api/v1/items", ListItems)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/items", nil)
	app.Engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp frame.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatal("expected map for Data")
	}
	total, ok := data["total"].(float64)
	if !ok || total <= 0 {
		t.Errorf("expected Total > 0, got %v", data["total"])
	}
}

func TestAdd(t *testing.T) {
	if got := Add(1, 2); got != 3 {
		t.Errorf("Add() = %v, want 3", got)
	}
}
