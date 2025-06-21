# HTMX Todo List App

A simple todo list web application built with Go, HTMX, and Tailwind CSS.

## Features

- ✅ Add new todos
- ✅ Mark todos as complete/incomplete
- ✅ Delete todos
- ✅ Real-time updates with HTMX
- ✅ Beautiful UI with Tailwind CSS
- ✅ In-memory storage (todos persist during server runtime)

## Technologies Used

- **Backend**: Go (Golang)
- **Frontend**: HTMX + Tailwind CSS
- **Storage**: In-memory (no database required)

## How to Run

1. Make sure you have Go installed (Go 1.24.4 or later)

2. Navigate to the project directory:
   ```bash
   cd agents/infra-agent/agent/working_dir
   ```

3. Run the application:
   ```bash
   go run main.go
   ```

4. Open your browser and go to:
   ```
   http://localhost:8080
   ```

## Usage

- **Add a todo**: Type in the input field and click "Add Todo" or press Enter
- **Complete a todo**: Click the circle next to the todo text
- **Delete a todo**: Click the trash icon (will ask for confirmation)

## API Endpoints

- `GET /` - Main page
- `GET /todos` - Get all todos (returns HTML fragment)
- `POST /todos/add` - Add a new todo
- `POST /todos/toggle/{id}` - Toggle todo completion status
- `DELETE /todos/delete/{id}` - Delete a todo

## Notes

- Todos are stored in memory and will be lost when the server is restarted
- The app uses HTMX for dynamic updates without full page reloads
- Responsive design works on desktop and mobile devices # ai-agents-demo-webapp
