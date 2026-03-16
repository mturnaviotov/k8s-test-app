package main

import (
	"fmt"
	"sync"
)

type Metrics struct {
	mu sync.Mutex

	TotalRequests   int
	TodoListFetched int
	TodoCreated     int
	TodoUpdated     int
	TodoDeleted     int
	TodoNotFound    int
	HealthChecks    int
}

var AppMetrics = &Metrics{}

func (m *Metrics) IncRequests() {
	m.mu.Lock()
	m.TotalRequests++
	m.mu.Unlock()
}

func (m *Metrics) IncTodoListFetched() {
	m.mu.Lock()
	m.TodoListFetched++
	m.mu.Unlock()
}

func (m *Metrics) IncTodoCreated() {
	m.mu.Lock()
	m.TodoCreated++
	m.mu.Unlock()
}

func (m *Metrics) IncTodoUpdated() {
	m.mu.Lock()
	m.TodoUpdated++
	m.mu.Unlock()
}

func (m *Metrics) IncTodoDeleted() {
	m.mu.Lock()
	m.TodoDeleted++
	m.mu.Unlock()
}

func (m *Metrics) IncTodoNotFound() {
	m.mu.Lock()
	m.TodoNotFound++
	m.mu.Unlock()
}

func (m *Metrics) IncHealthChecks() {
	m.mu.Lock()
	m.HealthChecks++
	m.mu.Unlock()
}

// Render for Prometheus
func (m *Metrics) Render() string {
	m.mu.Lock()
	defer m.mu.Unlock()

	return fmt.Sprintf(`
# HELP todoapp_requests_total Total HTTP requests
# TYPE todoapp_requests_total counter
todoapp_requests_total %d

# HELP todoapp_todos_fetched_total Total times todo list was fetched
# TYPE todoapp_todos_fetched_total counter
todoapp_todos_fetched_total %d

# HELP todoapp_todos_created_total Total todos created
# TYPE todoapp_todos_created_total counter
todoapp_todos_created_total %d

# HELP todoapp_todos_updated_total Total todos updated
# TYPE todoapp_todos_updated_total counter
todoapp_todos_updated_total %d

# HELP todoapp_todos_deleted_total Total todos deleted
# TYPE todoapp_todos_deleted_total counter
todoapp_todos_deleted_total %d

# HELP todoapp_todos_not_found_total Total todos deleted
# TYPE todoapp_todos_not_found_total counter
todoapp_todos_not_found_total %d

# HELP todoapp_healthchecks_total Total calls to /healthz
# TYPE todoapp_healthchecks_total counter
todoapp_healthchecks_total %d
`,
		m.TotalRequests,
		m.TodoListFetched,
		m.TodoCreated,
		m.TodoUpdated,
		m.TodoDeleted,
		m.TodoNotFound,
		m.HealthChecks,
	)
}
