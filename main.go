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
	return &TodoStore{
		todos:  []Todo{},
		nextID: 1,
	}
}

func (ts *TodoStore) AddTodo(text string) Todo {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	todo := Todo{
		ID:        ts.nextID,
		Text:      text,
		Completed: false,
		CreatedAt: time.Now(),
	}

	ts.todos = append(ts.todos, todo)
	ts.nextID++
	return todo
}

func (ts *TodoStore) GetTodos() []Todo {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	// Return a copy to avoid race conditions
	todos := make([]Todo, len(ts.todos))
	copy(todos, ts.todos)
	return todos
}

func (ts *TodoStore) ToggleTodo(id int) bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	for i, todo := range ts.todos {
		if todo.ID == id {
			ts.todos[i].Completed = !ts.todos[i].Completed
			return true
		}
	}
	return false
}

func (ts *TodoStore) DeleteTodo(id int) bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	for i, todo := range ts.todos {
		if todo.ID == id {
			ts.todos = append(ts.todos[:i], ts.todos[i+1:]...)
			return true
		}
	}
	return false
}

var store = NewTodoStore()
var templates = template.Must(template.ParseGlob("templates/*.html"))

func main() {
	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))

	// Routes
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/todos", todosHandler)
	http.HandleFunc("/todos/add", addTodoHandler)
	http.HandleFunc("/todos/toggle/", toggleTodoHandler)
	http.HandleFunc("/todos/delete/", deleteTodoHandler)

	fmt.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	todos := store.GetTodos()
	data := struct {
		Todos []Todo
	}{
		Todos: todos,
	}

	err := templates.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func todosHandler(w http.ResponseWriter, r *http.Request) {
	todos := store.GetTodos()
	data := struct {
		Todos []Todo
	}{
		Todos: todos,
	}

	err := templates.ExecuteTemplate(w, "todos.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func addTodoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	text := r.FormValue("text")
	if text == "" {
		http.Error(w, "Todo text is required", http.StatusBadRequest)
		return
	}

	store.AddTodo(text)

	// Return updated todo list
	todos := store.GetTodos()
	data := struct {
		Todos []Todo
	}{
		Todos: todos,
	}

	err := templates.ExecuteTemplate(w, "todos.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func toggleTodoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from URL path
	idStr := r.URL.Path[len("/todos/toggle/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	if !store.ToggleTodo(id) {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}

	// Return updated todo list
	todos := store.GetTodos()
	data := struct {
		Todos []Todo
	}{
		Todos: todos,
	}

	err = templates.ExecuteTemplate(w, "todos.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func deleteTodoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from URL path
	idStr := r.URL.Path[len("/todos/delete/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	if !store.DeleteTodo(id) {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}

	// Return updated todo list
	todos := store.GetTodos()
	data := struct {
		Todos []Todo
	}{
		Todos: todos,
	}

	err = templates.ExecuteTemplate(w, "todos.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
