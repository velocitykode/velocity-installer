// Package tui drives the interactive "velocity new" wizard using prism prompts.
package tui

import (
	"errors"
	"fmt"

	"github.com/velocitykode/prism"
	"github.com/velocitykode/velocity-installer/internal/generator"
)

// LaunchNewProjectWizard runs the interactive wizard to collect project config,
// then creates the project.
func LaunchNewProjectWizard(projectName string) {
	prism.Header("velocity new")

	if projectName == "" {
		projectName = prism.Text(
			"Project name:",
			prism.WithRequired(),
			prism.WithPlaceholder("my-awesome-app"),
		)
	}

	database := prism.Select(
		"Database:",
		[]string{"postgres", "mysql", "sqlite", "none"},
		prism.WithSelectDefault("postgres"),
	)
	if database == "none" {
		database = ""
	}

	cache := prism.Select(
		"Cache:",
		[]string{"redis", "memory", "none"},
		prism.WithSelectDefault("redis"),
	)
	if cache == "none" {
		cache = ""
	}

	features := prism.Multiselect(
		"Features (space to toggle, enter to confirm):",
		[]string{"auth", "api-only"},
	)
	featureSet := map[string]bool{}
	for _, f := range features {
		featureSet[f] = true
	}

	stack := ""
	ssr := false
	if !featureSet["api-only"] {
		stack = prism.Select(
			"Frontend stack:",
			generator.ValidStacks,
			prism.WithSelectDefault("react"),
		)
		ssr = prism.Confirm("Enable Inertia server-side rendering?", prism.WithDefaultNo())
	}

	prism.Newline()
	prism.Bold("Review:")
	prism.KeyValue("Project", prism.Highlight(projectName))
	if database != "" {
		prism.KeyValue("Database", database)
	}
	if cache != "" {
		prism.KeyValue("Cache", cache)
	}
	if featureSet["auth"] {
		prism.KeyValue("Auth", "yes")
	}
	if featureSet["api-only"] {
		prism.KeyValue("API-only", "yes")
	}
	if stack != "" {
		prism.KeyValue("Stack", stack)
	}
	if ssr {
		prism.KeyValue("SSR", "yes")
	}
	prism.Newline()

	if !prism.Confirm("Create project?", prism.WithDefaultYes()) {
		prism.Warning("Cancelled.")
		return
	}

	config := generator.ProjectConfig{
		Name:     projectName,
		Module:   projectName,
		Database: database,
		Cache:    cache,
		Auth:     featureSet["auth"],
		API:      featureSet["api-only"],
		Stack:    stack,
		SSR:      ssr,
	}

	if err := generator.CreateProject(config); err != nil && !errors.Is(err, generator.ErrMigrationsSkipped) {
		prism.Error(fmt.Sprintf("Error creating project: %v", err))
		return
	}

	showSuccess(projectName)
}

// CreateProjectWithDefaults creates a project with sensible defaults when the
// wizard can't run (e.g. no TTY).
func CreateProjectWithDefaults(projectName string) {
	config := generator.ProjectConfig{
		Name:   projectName,
		Module: projectName,
	}

	if err := generator.CreateProject(config); err != nil && !errors.Is(err, generator.ErrMigrationsSkipped) {
		prism.Error(fmt.Sprintf("Error creating project: %v", err))
		return
	}

	showSuccess(projectName)
}

func showSuccess(projectName string) {
	prism.Newline()
	prism.Success("Project created successfully!")
	prism.Newline()
	prism.Bold(fmt.Sprintf("Your Velocity project '%s' is ready.", projectName))
	prism.Newline()
	prism.NextSteps([]string{
		"cd " + projectName,
		"go run . serve",
	})
	prism.Muted("Default port: 4000 (set APP_PORT env to change)")
	prism.Newline()
}
