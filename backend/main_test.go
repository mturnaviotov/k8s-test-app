package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// --- Helper to create temporary DB for CRUD tests ---
func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
	})

	// create table
	_, err = db.Exec(`CREATE TABLE todos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		text TEXT NOT NULL,
		done BOOLEAN NOT NULL DEFAULT FALSE
	)`)
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	return db
}

// --- Health Check Test ---
func TestHealthHandler(t *testing.T) {
	db = setupTestDB(t)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	handler := http.HandlerFunc(healthHandler)
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	expected := `OK`
	if strings.TrimSpace(w.Body.String()) != expected {
		t.Fatalf("expected body %q, got %q", expected, w.Body.String())
	}

}

// --- Basic CRUD Smoke Test ---
func TestTodoCRUD(t *testing.T) {
	db = setupTestDB(t)

	// Create Todo
	reqBody := strings.NewReader(`{"text":"test todo"}`)
	req := httptest.NewRequest(http.MethodPost, "/todos", reqBody)
	w := httptest.NewRecorder()
	todosHandler(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", w.Code)
	}

	var created Todo
	if err := json.NewDecoder(w.Body).Decode(&created); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Get Todo
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/todos/%d", created.ID), nil)
	w = httptest.NewRecorder()
	todoHandler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	var fetched Todo
	if err := json.NewDecoder(w.Body).Decode(&fetched); err != nil {
		t.Fatalf("failed to decode fetched todo: %v", err)
	}

	if fetched.Text != created.Text {
		t.Fatalf("expected text %q, got %q", created.Text, fetched.Text)
	}

	// Update Todo
	updateBody := strings.NewReader(`{"done":true}`)
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/todos/%d", created.ID), updateBody)
	w = httptest.NewRecorder()
	todoHandler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on update, got %d", w.Code)
	}

	var updated Todo
	if err := json.NewDecoder(w.Body).Decode(&updated); err != nil {
		t.Fatalf("failed to decode updated todo: %v", err)
	}

	if !updated.Done {
		t.Fatalf("expected Done=true, got false")
	}

	// Delete Todo
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/todos/%d", created.ID), nil)
	w = httptest.NewRecorder()
	todoHandler(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204 No Content on delete, got %d", w.Code)
	}

}
