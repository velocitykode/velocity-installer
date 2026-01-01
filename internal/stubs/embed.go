package stubs

import "embed"

//go:embed app/http/controllers/*.stub app/middleware/*.stub routes/*.stub config/*.stub main.go.stub
var FS embed.FS

// Get returns the content of a stub file
func Get(name string) ([]byte, error) {
	return FS.ReadFile(name)
}
