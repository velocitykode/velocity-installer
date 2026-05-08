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
	// Mirrors the canonical .env.example shipped in
	// velocity-template-react / velocity-template-api so the init flow
	// (adding Velocity to an existing Go project) emits the same env
	// contract as the new-project flow. Env names track the framework
	// directly: AUTH_JWT_*, VIEW_SSR_*, MAIL_MAILGUN_*, MAIL_POSTMARK_*.
	// APP_KEY / QUEUE_SIGNING_KEY / AUTH_JWT_SECRET are generated below
	// before .env is written, so the bootstrapped app passes the
	// framework's mandatory-key check.
	envTemplate := `# App
APP_NAME={{ .Name }}
APP_ENV=development
APP_URL=http://localhost:4000
APP_PORT=4000

# Logging
LOG_DRIVER=console
LOG_LEVEL=info

# Encryption & signing - the installer populates these with random
# 32-byte base64 values at install time. APP_KEY is also the crypto
# fallback; set CRYPTO_KEY explicitly only if you want a dedicated
# crypto key separate from APP_KEY.
APP_KEY={{ .AppKey }}
QUEUE_SIGNING_KEY={{ .QueueSigningKey }}
AUTH_JWT_SECRET={{ .AuthJWTSecret }}
CRYPTO_CIPHER=AES-256-GCM

# Enable to revoke blacklisted JWT IDs.
AUTH_JWT_BLACKLIST_ENABLED=false
# AUTH_JWT_ALGO=HS256
# AUTH_JWT_TTL=60
# AUTH_JWT_REFRESH_TTL=20160
{{ if .Database }}
# Database
DB_CONNECTION={{ .Database }}{{ if eq .Database "postgres" }}
DB_HOST=127.0.0.1
DB_PORT=5432
DB_DATABASE={{ .Name }}_db
DB_USERNAME=postgres
DB_PASSWORD=
# DB_SSL_MODE=disable{{ else if eq .Database "mysql" }}
DB_HOST=127.0.0.1
DB_PORT=3306
DB_DATABASE={{ .Name }}_db
DB_USERNAME=root
DB_PASSWORD={{ else if eq .Database "sqlite" }}
DB_DATABASE=database.sqlite{{ end }}
{{ end }}{{ if .Cache }}
# Cache
CACHE_DRIVER={{ .Cache }}{{ if eq .Cache "redis" }}
REDIS_HOST=127.0.0.1
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DATABASE=0{{ end }}
{{ end }}
# Session
SESSION_NAME=velocity_session
SESSION_LIFETIME=120
SESSION_PATH=/
SESSION_SECURE=false
SESSION_HTTP_ONLY=true
SESSION_SAME_SITE=lax

# Filesystem
FILESYSTEM_DISK=local

# Mail
# MAIL_DRIVER=smtp
# MAIL_HOST=smtp.mailtrap.io
# MAIL_PORT=587
# MAIL_USERNAME=
# MAIL_PASSWORD=
# MAIL_FROM_ADDRESS=hello@example.com
# MAIL_FROM_NAME="${APP_NAME}"

# Mailgun
# MAIL_MAILGUN_DOMAIN=
# MAIL_MAILGUN_SECRET=
# MAIL_MAILGUN_ENDPOINT=api.mailgun.net

# Postmark
# MAIL_POSTMARK_TOKEN=
# MAIL_POSTMARK_MESSAGE_STREAM=outbound

# AWS / S3
# AWS_ACCESS_KEY_ID=
# AWS_SECRET_ACCESS_KEY=
# AWS_DEFAULT_REGION=us-east-1
# AWS_BUCKET=
{{ if not .API }}
# View SSR (Inertia)
# VIEW_SSR_ENABLED=false
# VIEW_SSR_URL=http://127.0.0.1:13714/render
# VIEW_SSR_TIMEOUT=3s
# VIEW_SSR_EXCEPT=/admin,/internal
{{ end }}`

	// Generate .env.example with placeholder keys (so it can be
	// committed to source control safely) and .env with real keys
	// populated for immediate boot.
	exampleData := envFileData{ProjectConfig: config}
	exampleData.AppKey = ""
	exampleData.QueueSigningKey = ""
	exampleData.AuthJWTSecret = ""
	if err := executeTemplate(filepath.Join(config.Name, ".env.example"), envTemplate, exampleData); err != nil {
		return err
	}

	envData := envFileData{ProjectConfig: config}
	for _, p := range []*string{&envData.AppKey, &envData.QueueSigningKey, &envData.AuthJWTSecret} {
		k, err := generateKey()
		if err != nil {
			return err
		}
		*p = k
	}
	return executeTemplate(filepath.Join(config.Name, ".env"), envTemplate, envData)
}

// envFileData wraps ProjectConfig with installer-generated secret
// material so the env template can render a self-contained .env in one
// pass. ProjectConfig is embedded so existing template references like
// {{ .Name }} / {{ .Database }} keep working.
type envFileData struct {
	ProjectConfig
	AppKey          string
	QueueSigningKey string
	AuthJWTSecret   string
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
.vel/

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
