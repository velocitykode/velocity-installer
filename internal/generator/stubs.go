package generator

import (
	"bytes"
	"os"
	"path/filepath"
	"text/template"

	"github.com/velocitykode/velocity-installer/internal/stubs"
)

// copyStubFile copies and processes a stub file from internal/stubs to the destination
func copyStubFile(stubName, destPath string) error {
	return copyStubFileWithConfig(stubName, destPath, nil)
}

// copyStubFileWithConfig copies and processes a stub file with template data
func copyStubFileWithConfig(stubName, destPath string, config interface{}) error {
	// Read from embedded stubs
	content, err := stubs.Get(stubName)
	if err != nil {
		return err
	}

	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}

	// If config is provided, process as template
	if config != nil {
		tmpl, err := template.New(stubName).Parse(string(content))
		if err != nil {
			return err
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, config); err != nil {
			return err
		}
		content = buf.Bytes()
	}

	// Write to destination
	return os.WriteFile(destPath, content, 0644)
}

// generateFilesFromStubs copies all necessary stub files to the project
func generateFilesFromStubs(config ProjectConfig) error {
	// Copy main.go with template processing
	if err := copyStubFileWithConfig("main.go.stub", filepath.Join(config.Name, "main.go"), config); err != nil {
		return err
	}

	// Copy home controller
	if err := copyStubFile("app/http/controllers/home_controller.go.stub", filepath.Join(config.Name, "app", "http", "controllers", "home_controller.go")); err != nil {
		return err
	}

	// Copy middleware
	if err := copyStubFile("app/http/middleware/middleware.go.stub", filepath.Join(config.Name, "app", "http", "middleware", "middleware.go")); err != nil {
		return err
	}

	// Copy web routes with template processing
	if err := copyStubFileWithConfig("routes/web.go.stub", filepath.Join(config.Name, "routes", "web.go"), config); err != nil {
		return err
	}

	// Copy config with template processing
	if err := copyStubFileWithConfig("config/config.go.stub", filepath.Join(config.Name, "config", "config.go"), config); err != nil {
		return err
	}

	// Copy API routes if API mode
	if config.API {
		if err := copyStubFileWithConfig("routes/api.go.stub", filepath.Join(config.Name, "routes", "api.go"), config); err != nil {
			return err
		}
	}

	// Copy auth files if auth is enabled
	if config.Auth {
		if err := copyStubFile("app/http/middleware/auth.go.stub", filepath.Join(config.Name, "app", "http", "middleware", "auth.go")); err != nil {
			return err
		}
	}

	return nil
}
