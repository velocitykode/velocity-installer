package generator

import (
	"os"
	"path/filepath"
	"text/template"
)

func generateProjectFiles(config ProjectConfig) error {
	// Generate .env.example
	if err := generateEnvFile(config); err != nil {
		return err
	}

	// Generate .gitignore
	if err := generateGitignore(config); err != nil {
		return err
	}

	// Generate README
	if err := generateReadme(config); err != nil {
		return err
	}

	return nil
}

func generateEnvFile(config ProjectConfig) error {
	envTemplate := `# Application
APP_NAME={{ .Name }}
APP_ENV=development
APP_PORT=4000
APP_URL=http://localhost:4000{{ if .Database }}

# Database
DB_CONNECTION={{ .Database }}
DB_HOST=localhost{{ if eq .Database "postgres" }}
DB_PORT=5432
DB_DATABASE={{ .Name }}_db
DB_USERNAME=postgres
DB_PASSWORD=password{{ else if eq .Database "mysql" }}
DB_PORT=3306
DB_DATABASE={{ .Name }}_db
DB_USERNAME=root
DB_PASSWORD=password{{ else if eq .Database "sqlite" }}
DB_PATH=./database/database.db{{ end }}{{ end }}{{ if .Cache }}

# Cache
CACHE_DRIVER={{ .Cache }}{{ if eq .Cache "redis" }}
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0{{ end }}{{ end }}{{ if .Auth }}

# Authentication
AUTH_SECRET=your-secret-key-here
AUTH_EXPIRY=24h{{ end }}

# Logging
LOG_LEVEL=debug
LOG_OUTPUT=stdout
`

	filePath := filepath.Join(config.Name, ".env.example")
	if err := executeTemplate(filePath, envTemplate, config); err != nil {
		return err
	}

	// Also create .env file
	filePath = filepath.Join(config.Name, ".env")
	return executeTemplate(filePath, envTemplate, config)
}

func generateGitignore(config ProjectConfig) error {
	gitignoreContent := `# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib
/{{ .Name }}
/build/
/dist/

# Test binary
*.test

# Output of go coverage
*.out
coverage.html

# Environment variables
.env
.env.local

# IDE
.idea/
.vscode/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db

# Velocity
.velocity/

# Logs
*.log
/storage/logs/*
!/storage/logs/.gitkeep

# Dependencies
/vendor/

# Database{{ if eq .Database "sqlite" }}
*.db
*.db-shm
*.db-wal{{ end }}
`

	filePath := filepath.Join(config.Name, ".gitignore")
	return executeTemplate(filePath, gitignoreContent, config)
}

func generateReadme(config ProjectConfig) error {
	readmeTemplate := `# {{ .Name }}

A web application built with the Velocity framework for Go.

## Features
{{ if .Database }}
- Database support ({{ .Database }}){{ end }}{{ if .Cache }}
- Caching ({{ .Cache }}){{ end }}{{ if .Auth }}
- Authentication system{{ end }}{{ if .API }}
- RESTful API{{ end }}

## Requirements

- Go 1.21 or higher{{ if eq .Database "postgres" }}
- PostgreSQL{{ else if eq .Database "mysql" }}
- MySQL{{ else if eq .Database "sqlite" }}
- SQLite{{ end }}{{ if eq .Cache "redis" }}
- Redis{{ end }}

## Installation

1. Clone the repository
2. Copy .env.example to .env and configure
3. Install dependencies:
   ` + "```bash" + `
   go mod download
   ` + "```" + `

## Running the Application

### Development
` + "```bash" + `
go run main.go
` + "```" + `

Or using the Velocity CLI:
` + "```bash" + `
velocity serve
` + "```" + `

### Production
` + "```bash" + `
go build -o {{ .Name }}
./{{ .Name }}
` + "```" + `

The application will start on http://localhost:4000 by default.

## Project Structure

` + "```" + `
.
├── app/
│   ├── controllers/    # HTTP controllers
│   ├── middleware/      # HTTP middleware
│   └── models/          # Data models
├── config/              # Configuration files
├── database/            # Database migrations and seeds
├── public/              # Static assets
├── resources/           # Views and resources
├── routes/              # Route definitions
├── storage/             # File storage and logs
└── main.go              # Application entry point
` + "```" + `

## License

MIT
`

	filePath := filepath.Join(config.Name, "README.md")
	return executeTemplate(filePath, readmeTemplate, config)
}

func executeTemplate(filePath, tmplContent string, data interface{}) error {
	tmpl, err := template.New("file").Parse(tmplContent)
	if err != nil {
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, data)
}
