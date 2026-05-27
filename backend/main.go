package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

var db *sql.DB
var tableCreated bool

type Todo struct {
	ID   uint64 `json:"id"`
	Text string `json:"text"`
	Done bool   `json:"done"`
}

func ensureTableExists() error {
	if tableCreated {
		return nil
	}
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS todos (
		id SERIAL PRIMARY KEY,
		text TEXT NOT NULL,
		done BOOLEAN NOT NULL DEFAULT FALSE
	)`)
	if err == nil {
		tableCreated = true
	}
	return err
}

func main() {
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "password"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "todoapp"
	}
	listenPort := os.Getenv("listenPort")
	if listenPort == "" {
		listenPort = "8080"
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("{\"level\":\"error\", \"message\":\"open db: %v\"}", err)
	}
	defer db.Close()

	// Test the connection (warn but don't fail startup if down)
	if err = db.Ping(); err != nil {
		log.Printf("{\"level\":\"warn\", \"message\":\"Database connection ping failed on startup (will retry on requests): %v\"}", err)
	} else {
		// Try to create table if database is accessible on startup
		if err := ensureTableExists(); err != nil {
			log.Printf("{\"level\":\"warn\", \"message\":\"Failed to create table on startup: %v\"}", err)
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthHandler)
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/todos", todosHandler)
	mux.HandleFunc("/todos/", todoHandler)
	mux.HandleFunc("/metrics", metrics)

	addr := ":" + listenPort
	log.Printf("{\"level\":\"info\", \"listening on\":\"%s\", \"db\":\"%s\"}", addr, dbName)
	log.Fatal(http.ListenAndServe(addr, enableCORSMiddleware(mux)))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if err := db.Ping(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "{\"level\":\"error\", \"message\":\"DB not accessible\"}")
		return
	}
	if err := ensureTableExists(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "{\"level\":\"error\", \"message\":\"Database initialization failed: %v\"}", err)
		return
	}
	AppMetrics.IncRequests()
	AppMetrics.IncHealthChecks()
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}

// enableCORS is a middleware that adds CORS headers and handles preflight OPTIONS requests.
func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// enableCORSMiddleware wraps a standard http.Handler with enableCORS middleware.
func enableCORSMiddleware(next http.Handler) http.Handler {
	return enableCORS(next.ServeHTTP)
}

func todosHandler(w http.ResponseWriter, r *http.Request) {
	if err := ensureTableExists(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "{\"level\":\"error\", \"message\":\"Database initialization failed: %v\"}", err)
		return
	}
	switch r.Method {
	case http.MethodGet:
		listTodos(w)
	case http.MethodPost:
		createTodo(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func todoHandler(w http.ResponseWriter, r *http.Request) {
	if err := ensureTableExists(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "{\"level\":\"error\", \"message\":\"Database initialization failed: %v\"}", err)
		return
	}
	idStr := strings.TrimPrefix(r.URL.Path, "/todos/")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("{\"level\":\"error\", \"message\":\"Invalid todo ID %s\"}", idStr)
		fmt.Fprint(w, "{\"level\":\"error\", \"message\":\"Invalid todo ID\"}")
		return
	}
	switch r.Method {
	case http.MethodGet:
		getTodo(w, id)
	case http.MethodPut:
		updateTodo(w, r, id)
	case http.MethodDelete:
		deleteTodo(w, id)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func listTodos(w http.ResponseWriter) {
	AppMetrics.IncRequests()
	rows, err := db.Query("SELECT id, text, done FROM todos ORDER BY id")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}
	defer rows.Close()

	var out []Todo
	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.Text, &t.Done); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, err.Error())
			return
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}

	AppMetrics.IncTodoListFetched()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func createTodo(w http.ResponseWriter, r *http.Request) {
	AppMetrics.IncRequests()
	var in struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("{\"level\":\"error\", \"message\":\"invalid body\"}")
		fmt.Fprint(w, "{\"level\":\"error\", \"message\":\"invalid body\"}")
		return
	}

	var created Todo
	err := db.QueryRow("INSERT INTO todos (text, done) VALUES ($1, $2) RETURNING id", in.Text, false).Scan(&created.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}
	created.Text = in.Text
	created.Done = false

	log.Printf("{\"level\":\"info\", \"Todo created\":\"%+v\"}", created)
	AppMetrics.IncTodoCreated()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(created)
}

func getTodo(w http.ResponseWriter, id uint64) {
	AppMetrics.IncRequests()
	var t Todo
	err := db.QueryRow("SELECT id, text, done FROM todos WHERE id = $1", id).Scan(&t.ID, &t.Text, &t.Done)
	if err != nil {
		if err == sql.ErrNoRows {
			AppMetrics.IncTodoNotFound()
			w.WriteHeader(http.StatusNotFound)
			log.Printf("{\"level\":\"error\", \"message\":\"Todo not found %d\"}", id)
			fmt.Fprint(w, "{\"level\":\"error\", \"message\":\"Todo not found\"}")
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}
	AppMetrics.IncTodoUpdated()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(t)
}

func updateTodo(w http.ResponseWriter, r *http.Request, id uint64) {
	AppMetrics.IncRequests()
	var in struct {
		Text *string `json:"text"`
		Done *bool   `json:"done"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("{\"level\":\"error\", \"message\":\"invalid body\"}")
		fmt.Fprint(w, "{\"level\":\"error\", \"message\":\"invalid body\"}")
		return
	}

	// First check if todo exists
	var updated Todo
	err := db.QueryRow("SELECT id, text, done FROM todos WHERE id = $1", id).Scan(&updated.ID, &updated.Text, &updated.Done)
	if err != nil {
		if err == sql.ErrNoRows {
			AppMetrics.IncTodoNotFound()
			w.WriteHeader(http.StatusNotFound)
			log.Printf("{\"level\":\"error\", \"message\":\"Todo not found %d\"}", id)
			fmt.Fprint(w, "{\"level\":\"error\", \"message\":\"Todo not found\"}")
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}

	// Update fields
	if in.Text != nil {
		updated.Text = *in.Text
	}
	if in.Done != nil {
		updated.Done = *in.Done
	}

	_, err = db.Exec("UPDATE todos SET text = $1, done = $2 WHERE id = $3", updated.Text, updated.Done, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}

	log.Printf("{\"level\":\"info\", \"Todo updated\":\"%+v\"}", updated)
	AppMetrics.IncTodoUpdated()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(updated)
}

func deleteTodo(w http.ResponseWriter, id uint64) {
	AppMetrics.IncRequests()
	result, err := db.Exec("DELETE FROM todos WHERE id = $1", id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}
	if rowsAffected == 0 {
		AppMetrics.IncTodoNotFound()
		w.WriteHeader(http.StatusNotFound)
		log.Printf("{\"level\":\"error\", \"message\":\"Todo not found %d\"}", id)
		fmt.Fprint(w, "{\"level\":\"error\", \"message\":\"Todo not found\"}")
		return
	}
	AppMetrics.IncTodoDeleted()
	log.Printf("{\"level\":\"info\", \"Todo deleted\":\"%d\"}", id)
	w.WriteHeader(http.StatusNoContent)
}

func metrics(w http.ResponseWriter, r *http.Request) {
	AppMetrics.IncRequests()
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	fmt.Fprint(w, AppMetrics.Render())
}
