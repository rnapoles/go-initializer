package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// ProjectInitializer handles the Go project initialization
type ProjectInitializer struct {
	ProjectName string
	ProjectDir  string
	GitUsername string
	ModuleName  string
	IsRestAPI   bool
}

func main() {

	// Check for project name argument
	if len(os.Args) < 2 {
		fmt.Println("Error: Please provide a project name.")
		fmt.Println("Usage: go-init project-name [--rest-api]")
		os.Exit(1)
	}

	initializer := &ProjectInitializer{
		ProjectName: os.Args[1],
	}

	// Check for --rest-api flag
	for i := 2; i < len(os.Args); i++ {
		if os.Args[i] == "--rest-api" {
			initializer.IsRestAPI = true
		}
	}

	// Set project directory
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		os.Exit(1)
	}
	initializer.ProjectDir = filepath.Join(currentDir, initializer.ProjectName)

	// Run the initialization process
	initializer.Run()
}

// Run executes the full project initialization process
func (p *ProjectInitializer) Run() {
	fmt.Println("Go Project Initializer v1.0")
	fmt.Println("----------------------------")

	if p.IsRestAPI {
		fmt.Println("REST API mode enabled")
	}

	// Check if project directory already exists
	if _, err := os.Stat(p.ProjectDir); !os.IsNotExist(err) {
		fmt.Printf("Error: The directory \"%s\" already exists.\n", p.ProjectName)
		os.Exit(1)
	}

	// Get Git username from email
	p.GitUsername = p.getGitUsername()

	// Create project directory
	fmt.Printf("Creating project directory: %s...\n", p.ProjectName)
	if err := os.MkdirAll(p.ProjectDir, 0755); err != nil {
		fmt.Printf("Error creating project directory: %v\n", err)
		os.Exit(1)
	}

	// Change to project directory
	if err := os.Chdir(p.ProjectDir); err != nil {
		fmt.Printf("Error changing to project directory: %v\n", err)
		os.Exit(1)
	}

	// Initialize Go module
	p.initGoModule()

	// Create project structure
	p.createProjectStructure()

	// Create project files
	if p.IsRestAPI {
		p.createRestAPIFiles()
	} else {
		p.createMainGoFile()
	}

	p.createReadmeFile()
	p.createGitignoreFile()
	p.createMakefileFile()

	// Initialize Git repository
	p.initGitRepository()

	// Print success message
	p.printSuccessMessage()
}

// getGitUsername extracts username from git email
func (p *ProjectInitializer) getGitUsername() string {
	// Try to get git email
	cmd := exec.Command("git", "config", "user.email")
	output, err := cmd.Output()

	if err == nil && len(output) > 0 {
		email := strings.TrimSpace(string(output))
		// Extract username part (before @)
		if parts := strings.Split(email, "@"); len(parts) > 0 {
			return parts[0]
		}
	}

	// If no email found or couldn't extract username, prompt user
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("GitHub username (default: github-user): ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	if username == "" {
		return "github-user"
	}
	return username
}

// initGoModule initializes the Go module
func (p *ProjectInitializer) initGoModule() {
	fmt.Println("Initializing Go module...")

	// Prompt for module name
	defaultModule := fmt.Sprintf("github.com/%s/%s", p.GitUsername, p.ProjectName)

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Module name (default %s): ", defaultModule)
	moduleName, _ := reader.ReadString('\n')
	moduleName = strings.TrimSpace(moduleName)

	if moduleName == "" {
		moduleName = defaultModule
	}
	p.ModuleName = moduleName

	// Run go mod init
	cmd := exec.Command("go", "mod", "init", p.ModuleName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error initializing Go module: %v\n", err)
		os.Exit(1)
	}
}

// createProjectStructure creates the standard Go project directory structure
func (p *ProjectInitializer) createProjectStructure() {
	fmt.Println("Creating standard Go project structure...")

	// Define base directories
	dirs := []string{
		filepath.Join("cmd", p.ProjectName),
		"internal",
		"pkg",
	}

	// Add REST API specific directories if flag is set
	if p.IsRestAPI {
		apiDirs := []string{
			"api",
			"api/handlers",
			"api/middleware",
			"api/routes",
			"web",
			"web/templates",
			"web/static",
			"web/static/css",
			"web/static/js",
			"internal/models",
			"internal/database",
			"configs",
			"test",
		}
		dirs = append(dirs, apiDirs...)
	} else {
		// Default directories for non-API projects
		standardDirs := []string{
			"configs",
			"test",
		}
		dirs = append(dirs, standardDirs...)
	}

	// Create each directory
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("Error creating directory %s: %v\n", dir, err)
			os.Exit(1)
		}
	}
}

// createRestAPIFiles creates files for a REST API project
func (p *ProjectInitializer) createRestAPIFiles() {
	fmt.Println("Creating REST API files...")

	// Create main.go
	mainContent := fmt.Sprintf(`package main

import (
	"fmt"
	"log"
	"net/http"
	"github.com/%s/%s/api/routes"
)

func main() {
	// Initialize router
	router := routes.InitRoutes()
	
	// Start server
	port := ":8080"
	fmt.Printf("Server starting on port %%s...\n", port)
	log.Fatal(http.ListenAndServe(port, router))
}
`, p.GitUsername, p.ProjectName)

	mainPath := filepath.Join("cmd", p.ProjectName, "main.go")
	if err := ioutil.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
		fmt.Printf("Error creating main.go file: %v\n", err)
		os.Exit(1)
	}

	// Create routes.go
	routesContent := `package routes

import (
	// "net/http"
	"github.com/gorilla/mux"
	"` + p.ModuleName + `/api/handlers"
)

// InitRoutes initializes the router and sets up routes
func InitRoutes() *mux.Router {
	router := mux.NewRouter()
	
	// Define API routes
	router.HandleFunc("/api/health", handlers.HealthCheckHandler).Methods("GET")
	
	// Add your routes here
	// router.HandleFunc("/api/resource", handlers.GetResourceHandler).Methods("GET")
	
	return router
}
`

	routesPath := filepath.Join("api", "routes", "routes.go")
	if err := ioutil.WriteFile(routesPath, []byte(routesContent), 0644); err != nil {
		fmt.Printf("Error creating routes.go file: %v\n", err)
		os.Exit(1)
	}

	// Create handlers.go
	handlersContent := `package handlers

import (
	"encoding/json"
	"net/http"
)

// Response represents a standard API response
type Response struct {
	Status  string      ` + "`json:\"status\"`" + `
	Message string      ` + "`json:\"message,omitempty\"`" + `
	Data    interface{} ` + "`json:\"data,omitempty\"`" + `
}

// HealthCheckHandler returns a 200 OK status when the API is available
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	response := Response{
		Status:  "success",
		Message: "API is up and running",
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
`

	handlersPath := filepath.Join("api", "handlers", "handlers.go")
	if err := ioutil.WriteFile(handlersPath, []byte(handlersContent), 0644); err != nil {
		fmt.Printf("Error creating handlers.go file: %v\n", err)
		os.Exit(1)
	}

	// Create middleware example
	middlewareContent := `package middleware

import (
	"net/http"
	"time"
)

// Logger is a middleware that logs request details
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Call the next handler
		next.ServeHTTP(w, r)
		
		// Log the request after handling it
		duration := time.Since(start)
		
		// You can use a proper logger here
		// log.Printf("%s %s %s %s", r.Method, r.RequestURI, r.RemoteAddr, duration)
		println(r.Method, r.RequestURI, duration.String())
	})
}
`

	middlewarePath := filepath.Join("api", "middleware", "middleware.go")
	if err := ioutil.WriteFile(middlewarePath, []byte(middlewareContent), 0644); err != nil {
		fmt.Printf("Error creating middleware.go file: %v\n", err)
		os.Exit(1)
	}

	// Add a go.mod file with dependencies for a REST API
	cmd := exec.Command("go", "get", "github.com/gorilla/mux")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Warning: Unable to add gorilla/mux dependency: %v\n", err)
	}
}

// createMainGoFile creates the main.go file for a standard project
func (p *ProjectInitializer) createMainGoFile() {
	fmt.Println("Creating main.go file...")

	content := fmt.Sprintf(`package main

import (
    "fmt"
)

func main() {
    fmt.Println("Hello from %s!")
}
`, p.ProjectName)

	filePath := filepath.Join("cmd", p.ProjectName, "main.go")
	if err := ioutil.WriteFile(filePath, []byte(content), 0644); err != nil {
		fmt.Printf("Error creating main.go file: %v\n", err)
		os.Exit(1)
	}
}

// createReadmeFile creates the README.md file
func (p *ProjectInitializer) createReadmeFile() {
	fmt.Println("Creating README.md...")

	version := runtime.Version()

	restApiContent := ""
	if p.IsRestAPI {
		restApiContent = `
### API Endpoints

- GET /api/health - Health check endpoint

### Running the API

` + "```" + `bash
go run cmd/` + p.ProjectName + `/main.go
` + "```" + `

The API will be available at http://localhost:8080
`
	}

	content := fmt.Sprintf(`# %s

A Go project created with go-init.%s

## Getting Started

### Prerequisites

- Go %s or higher

### Installation

`+"```"+`bash
go get %s
`+"```"+`

## Usage

`+"```"+`bash
go run cmd/%s/main.go
`+"```"+`
`, p.ProjectName, restApiContent, version, p.ModuleName, p.ProjectName)

	if err := ioutil.WriteFile("README.md", []byte(content), 0644); err != nil {
		fmt.Printf("Error creating README.md file: %v\n", err)
		os.Exit(1)
	}
}

// createGitignoreFile creates the .gitignore file
func (p *ProjectInitializer) createGitignoreFile() {
	fmt.Println("Creating .gitignore file...")

	content := `# Binaries for programs and plugins
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary, built with ` + "`go test -c`" + `
*.test

# Output of the go coverage tool
*.out

# Go workspace file
go.work

# Dependency directories
vendor/

# IDEs
.idea/
.vscode/

# Environment files
.env
.env.local
`

	if err := ioutil.WriteFile(".gitignore", []byte(content), 0644); err != nil {
		fmt.Printf("Error creating .gitignore file: %v\n", err)
		os.Exit(1)
	}
}

// createMakefileFile creates the Makefile
func (p *ProjectInitializer) createMakefileFile() {
	fmt.Println("Creating Makefile...")

	content := fmt.Sprintf(`.PHONY: build clean test run

build:
	go build -o bin/%s cmd/%s/main.go

test:
	go test ./...

clean:
	rm -rf bin/

run:
	go run cmd/%s/main.go
`, p.ProjectName, p.ProjectName, p.ProjectName)

	if err := ioutil.WriteFile("Makefile", []byte(content), 0644); err != nil {
		fmt.Printf("Error creating Makefile: %v\n", err)
		os.Exit(1)
	}
}

// initGitRepository initializes a Git repository
func (p *ProjectInitializer) initGitRepository() {
	fmt.Println("Initializing Git repository...")

	cmd := exec.Command("git", "init")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error initializing Git repository: %v\n", err)
		// Don't exit - this isn't critical
	}
}

// printSuccessMessage prints information about the created project
func (p *ProjectInitializer) printSuccessMessage() {

	fmt.Println()
	//dirtree.Write(os.Stdout, p.ProjectDir)
	//dirtree.Write(os.Stdout, p.ProjectDir, dirtree.Depth(2), dirtree.ModeSize|dirtree.ModeCRC32)

	fmt.Printf("Successfully created Go project: %s\n", p.ProjectName)
	fmt.Println()
	fmt.Println("Project structure:")
	fmt.Printf("- %s/\n", p.ProjectName)
	fmt.Println("  |- cmd/")
	fmt.Printf("  |  \\- %s/ (application entrypoints)\n", p.ProjectName)
	fmt.Println("  |     \\- main.go")
	fmt.Println("  |- internal/ (private code)")

	if p.IsRestAPI {
		fmt.Println("  |  |- models/ (data models)")
		fmt.Println("  |  \\- database/ (database connections)")
	}

	fmt.Println("  |- pkg/ (public code)")

	if p.IsRestAPI {
		fmt.Println("  |- api/ (API definitions)")
		fmt.Println("  |  |- handlers/ (request handlers)")
		fmt.Println("  |  |- middleware/ (HTTP middleware)")
		fmt.Println("  |  \\- routes/ (route definitions)")
		fmt.Println("  |- web/ (web assets)")
		fmt.Println("  |  |- templates/ (HTML templates)")
		fmt.Println("  |  \\- static/ (static assets)")
		fmt.Println("  |     |- css/ (stylesheets)")
		fmt.Println("  |     \\- js/ (javascript files)")
	} else {
		fmt.Println("  |- api/ (API definitions)")
	}

	fmt.Println("  |- configs/ (configuration files)")
	fmt.Println("  |- test/ (test files)")
	fmt.Println("  |- README.md")
	fmt.Println("  |- .gitignore")
	fmt.Println("  |- Makefile")
	fmt.Println("  \\- go.mod")
	fmt.Println()

	if p.IsRestAPI {
		fmt.Println("To run your REST API:")
		fmt.Printf("  cd %s\n", p.ProjectName)
		fmt.Printf("  go run cmd/%s/main.go\n", p.ProjectName)
		fmt.Println()
		fmt.Println("Your API will be available at: http://localhost:8080")
		fmt.Println("Health check endpoint: http://localhost:8080/api/health")
	} else {
		fmt.Println("To run your project:")
		fmt.Printf("  cd %s\n", p.ProjectName)
		fmt.Printf("  go run cmd/%s/main.go\n", p.ProjectName)
	}

	fmt.Println()
	fmt.Println("Happy coding!")
}
