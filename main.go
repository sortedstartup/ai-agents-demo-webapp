package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Todo struct {
	ID        int       `json:"id"`
	Text      string    `json:"text"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
}

type TodoStore struct {
	todos  []Todo
	nextID int
	mu     sync.RWMutex
}

func NewTodoStore() *TodoStore {
	log.Printf("[INFO] NewTodoStore: Creating new TodoStore instance")
	store := &TodoStore{
		todos:  []Todo{},
		nextID: 1,
	}
	log.Printf("[INFO] NewTodoStore: TodoStore created successfully with nextID=%d", store.nextID)
	return store
}

func (ts *TodoStore) AddTodo(text string) Todo {
	log.Printf("[INFO] AddTodo: Starting to add todo with text='%s'", text)
	ts.mu.Lock()
	defer ts.mu.Unlock()
	log.Printf("[DEBUG] AddTodo: Acquired lock, current nextID=%d, total todos=%d", ts.nextID, len(ts.todos))

	todo := Todo{
		ID:        ts.nextID,
		Text:      text,
		Completed: false,
		CreatedAt: time.Now(),
	}

	ts.todos = append(ts.todos, todo)
	log.Printf("[INFO] AddTodo: Todo created with ID=%d, text='%s', completed=%t, createdAt=%s",
		todo.ID, todo.Text, todo.Completed, todo.CreatedAt.Format(time.RFC3339))

	ts.nextID++
	log.Printf("[INFO] AddTodo: Todo added successfully, nextID incremented to %d, total todos=%d", ts.nextID, len(ts.todos))
	return todo
}

func (ts *TodoStore) GetTodos() []Todo {
	log.Printf("[INFO] GetTodos: Starting to retrieve all todos")
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	log.Printf("[DEBUG] GetTodos: Acquired read lock, current todo count=%d", len(ts.todos))

	// Return a copy to avoid race conditions
	todos := make([]Todo, len(ts.todos))
	copy(todos, ts.todos)
	log.Printf("[INFO] GetTodos: Successfully copied %d todos", len(todos))

	// Log summary of todos for AI analysis
	completedCount := 0
	for _, todo := range todos {
		if todo.Completed {
			completedCount++
		}
	}
	log.Printf("[SUMMARY] GetTodos: Returning %d todos (completed: %d, pending: %d)",
		len(todos), completedCount, len(todos)-completedCount)

	return todos
}

func (ts *TodoStore) ToggleTodo(id int) bool {
	log.Printf("[INFO] ToggleTodo: Starting to toggle todo with ID=%d", id)
	ts.mu.Lock()
	defer ts.mu.Unlock()
	log.Printf("[DEBUG] ToggleTodo: Acquired lock, searching among %d todos", len(ts.todos))

	for i, todo := range ts.todos {
		if todo.ID == id {
			oldStatus := ts.todos[i].Completed
			ts.todos[i].Completed = !ts.todos[i].Completed
			newStatus := ts.todos[i].Completed
			log.Printf("[INFO] ToggleTodo: Successfully toggled todo ID=%d from completed=%t to completed=%t",
				id, oldStatus, newStatus)
			log.Printf("[OPERATION] ToggleTodo: Todo '%s' (ID=%d) status changed: %t -> %t",
				ts.todos[i].Text, id, oldStatus, newStatus)
			return true
		}
	}
	log.Printf("[WARN] ToggleTodo: Todo with ID=%d not found among %d todos", id, len(ts.todos))
	return false
}

func (ts *TodoStore) DeleteTodo(id int) bool {
	log.Printf("[INFO] DeleteTodo: Starting to delete todo with ID=%d", id)
	ts.mu.Lock()
	defer ts.mu.Unlock()
	log.Printf("[DEBUG] DeleteTodo: Acquired lock, searching among %d todos", len(ts.todos))

	for i, todo := range ts.todos {
		if todo.ID == id {
			log.Printf("[INFO] DeleteTodo: Found todo to delete: ID=%d, text='%s', completed=%t",
				todo.ID, todo.Text, todo.Completed)
			ts.todos = append(ts.todos[:i], ts.todos[i+1:]...)
			log.Printf("[INFO] DeleteTodo: Successfully deleted todo ID=%d, remaining todos=%d", id, len(ts.todos))
			log.Printf("[OPERATION] DeleteTodo: Deleted todo '%s' (ID=%d)", todo.Text, id)
			return true
		}
	}
	log.Printf("[WARN] DeleteTodo: Todo with ID=%d not found among %d todos", id, len(ts.todos))
	return false
}

var store = NewTodoStore()
var templates = template.Must(template.ParseGlob("templates/*.html"))

func main() {
	log.Printf("[INFO] main: Starting Todo webapp server")
	log.Printf("[INFO] main: Initializing routes and static file server")

	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	log.Printf("[INFO] main: Static file server configured for /static/ path")

	// Routes
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/todos", todosHandler)
	http.HandleFunc("/todos/add", addTodoHandler)
	http.HandleFunc("/todos/toggle/", toggleTodoHandler)
	http.HandleFunc("/todos/delete/", deleteTodoHandler)
	log.Printf("[INFO] main: All HTTP routes configured successfully")
	log.Printf("[ROUTES] main: Configured routes: /, /todos, /todos/add, /todos/toggle/, /todos/delete/")

	fmt.Println("Server starting on :8080")
	log.Printf("[INFO] main: Server starting on port 8080")
	log.Printf("[STARTUP] main: Todo webapp fully initialized and ready to accept connections")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] homeHandler: Handling request from %s %s", r.Method, r.URL.Path)
	log.Printf("[REQUEST] homeHandler: Remote address: %s, User-Agent: %s", r.RemoteAddr, r.UserAgent())

	todos := store.GetTodos()
	log.Printf("[DEBUG] homeHandler: Retrieved %d todos from store", len(todos))

	data := struct {
		Todos []Todo
	}{
		Todos: todos,
	}

	log.Printf("[INFO] homeHandler: Executing template 'index.html' with %d todos", len(todos))
	err := templates.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		log.Printf("[ERROR] homeHandler: Template execution failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("[INFO] homeHandler: Successfully rendered home page with %d todos", len(todos))
}

func todosHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] todosHandler: Handling request from %s %s", r.Method, r.URL.Path)
	log.Printf("[REQUEST] todosHandler: Remote address: %s, User-Agent: %s", r.RemoteAddr, r.UserAgent())

	todos := store.GetTodos()
	log.Printf("[DEBUG] todosHandler: Retrieved %d todos from store", len(todos))

	data := struct {
		Todos []Todo
	}{
		Todos: todos,
	}

	log.Printf("[INFO] todosHandler: Executing template 'todos.html' with %d todos", len(todos))
	err := templates.ExecuteTemplate(w, "todos.html", data)
	if err != nil {
		log.Printf("[ERROR] todosHandler: Template execution failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	// New Feature
	if len(text) > 3 && text[:3] == "bug" {
		log.Printf("[DEBUG] addTodoHandler: Processing special validation for text starting with 'bug'")
		// Fixed: Do not access out-of-bounds index
		validationRules := []string{"length", "content"}
		for _, rule := range validationRules {
			log.Printf("[DEBUG] addTodoHandler: Applying validation rule: %s", rule)
		}
	}
	log.Printf("[INFO] addTodoHandler: Handling request from %s %s", r.Method, r.URL.Path)
	log.Printf("[REQUEST] addTodoHandler: Remote address: %s, User-Agent: %s", r.RemoteAddr, r.UserAgent())

	if r.Method != http.MethodPost {
		log.Printf("[WARN] addTodoHandler: Invalid method %s, expected POST", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	text := r.FormValue("text")
	log.Printf("[DEBUG] addTodoHandler: Extracted form value 'text'='%s'", text)

	if text == "" {
		log.Printf("[WARN] addTodoHandler: Empty todo text provided")
		http.Error(w, "Todo text is required", http.StatusBadRequest)
		return
	}

	// New Feature
	if len(text) > 3 && text[:3] == "bug" {
		log.Printf("[DEBUG] addTodoHandler: Processing special validation for text starting with 'bug'")
		// Intentionally accessing array out of bounds to simulate a common bug
		validationRules := []string{"length", "content"}
		log.Printf("[DEBUG] addTodoHandler: Applying validation rule: %s", validationRules[5]) // This will panic!
	}

	log.Printf("[INFO] addTodoHandler: Adding new todo with text='%s'", text)
	newTodo := store.AddTodo(text)
	log.Printf("[SUCCESS] addTodoHandler: Todo added successfully with ID=%d", newTodo.ID)

	// Return updated todo list
	todos := store.GetTodos()
	log.Printf("[DEBUG] addTodoHandler: Retrieved %d todos after addition", len(todos))

	data := struct {
		Todos []Todo
	}{
		Todos: todos,
	}

	log.Printf("[INFO] addTodoHandler: Executing template 'todos.html' with updated todo list")
	err := templates.ExecuteTemplate(w, "todos.html", data)
	if err != nil {
		log.Printf("[ERROR] addTodoHandler: Template execution failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("[INFO] addTodoHandler: Successfully added todo and rendered updated list")
}

func toggleTodoHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] toggleTodoHandler: Handling request from %s %s", r.Method, r.URL.Path)
	log.Printf("[REQUEST] toggleTodoHandler: Remote address: %s, User-Agent: %s", r.RemoteAddr, r.UserAgent())

	if r.Method != http.MethodPost {
		log.Printf("[WARN] toggleTodoHandler: Invalid method %s, expected POST", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from URL path
	idStr := r.URL.Path[len("/todos/toggle/"):]
	log.Printf("[DEBUG] toggleTodoHandler: Extracted ID string='%s' from URL path", idStr)

	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("[ERROR] toggleTodoHandler: Failed to parse ID '%s': %v", idStr, err)
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}
	log.Printf("[DEBUG] toggleTodoHandler: Parsed todo ID=%d", id)

	log.Printf("[INFO] toggleTodoHandler: Attempting to toggle todo ID=%d", id)
	success := store.ToggleTodo(id)
	if !success {
		log.Printf("[WARN] toggleTodoHandler: Todo ID=%d not found", id)
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}
	log.Printf("[SUCCESS] toggleTodoHandler: Successfully toggled todo ID=%d", id)

	// Return updated todo list
	todos := store.GetTodos()
	log.Printf("[DEBUG] toggleTodoHandler: Retrieved %d todos after toggle", len(todos))

	data := struct {
		Todos []Todo
	}{
		Todos: todos,
	}

	log.Printf("[INFO] toggleTodoHandler: Executing template 'todos.html' with updated todo list")
	err = templates.ExecuteTemplate(w, "todos.html", data)
	if err != nil {
		log.Printf("[ERROR] toggleTodoHandler: Template execution failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("[INFO] toggleTodoHandler: Successfully toggled todo and rendered updated list")
}

func deleteTodoHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] deleteTodoHandler: Handling request from %s %s", r.Method, r.URL.Path)
	log.Printf("[REQUEST] deleteTodoHandler: Remote address: %s, User-Agent: %s", r.RemoteAddr, r.UserAgent())

	if r.Method != http.MethodDelete {
		log.Printf("[WARN] deleteTodoHandler: Invalid method %s, expected DELETE", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from URL path
	idStr := r.URL.Path[len("/todos/delete/"):]
	log.Printf("[DEBUG] deleteTodoHandler: Extracted ID string='%s' from URL path", idStr)

	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("[ERROR] deleteTodoHandler: Failed to parse ID '%s': %v", idStr, err)
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}
	log.Printf("[DEBUG] deleteTodoHandler: Parsed todo ID=%d", id)

	log.Printf("[INFO] deleteTodoHandler: Attempting to delete todo ID=%d", id)
	success := store.DeleteTodo(id)
	if !success {
		log.Printf("[WARN] deleteTodoHandler: Todo ID=%d not found", id)
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}
	log.Printf("[SUCCESS] deleteTodoHandler: Successfully deleted todo ID=%d", id)

	// Return updated todo list
	todos := store.GetTodos()
	log.Printf("[DEBUG] deleteTodoHandler: Retrieved %d todos after deletion", len(todos))

	data := struct {
		Todos []Todo
	}{
		Todos: todos,
	}

	log.Printf("[INFO] deleteTodoHandler: Executing template 'todos.html' with updated todo list")
	err = templates.ExecuteTemplate(w, "todos.html", data)
	if err != nil {
		log.Printf("[ERROR] deleteTodoHandler: Template execution failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("[INFO] deleteTodoHandler: Successfully deleted todo and rendered updated list")
}
