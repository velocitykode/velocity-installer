// Package tui drives the interactive "velocity new" wizard using velocity-cli prompts.
package tui

import (
	"errors"
	"fmt"

	cli "github.com/velocitykode/velocity-cli"
	"github.com/velocitykode/velocity-installer/internal/generator"
)

// LaunchNewProjectWizard runs the interactive wizard to collect project config,
// then creates the project.
func LaunchNewProjectWizard(projectName string) {
	cli.Header("velocity new")

	if projectName == "" {
		projectName = cli.Text(
			"Project name:",
			cli.WithRequired(),
			cli.WithPlaceholder("my-awesome-app"),
		)
	}

	database := cli.Select(
		"Database:",
		[]string{"postgres", "mysql", "sqlite", "none"},
		cli.WithSelectDefault("postgres"),
	)
	if database == "none" {
		database = ""
	}

	cache := cli.Select(
		"Cache:",
		[]string{"redis", "memory", "none"},
		cli.WithSelectDefault("redis"),
	)
	if cache == "none" {
		cache = ""
	}

	features := cli.Multiselect(
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
		stack = cli.Select(
			"Frontend stack:",
			generator.ValidStacks,
			cli.WithSelectDefault("react"),
		)
		ssr = cli.Confirm("Enable Inertia server-side rendering?", cli.WithDefaultNo())
	}

	cli.Newline()
	cli.Bold("Review:")
	cli.KeyValue("Project", cli.Highlight(projectName))
	if database != "" {
		cli.KeyValue("Database", database)
	}
	if cache != "" {
		cli.KeyValue("Cache", cache)
	}
	if featureSet["auth"] {
		cli.KeyValue("Auth", "yes")
	}
	if featureSet["api-only"] {
		cli.KeyValue("API-only", "yes")
	}
	if stack != "" {
		cli.KeyValue("Stack", stack)
	}
	if ssr {
		cli.KeyValue("SSR", "yes")
	}
	cli.Newline()

	if !cli.Confirm("Create project?", cli.WithDefaultYes()) {
		cli.Warning("Cancelled.")
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
		cli.Error(fmt.Sprintf("Error creating project: %v", err))
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
		cli.Error(fmt.Sprintf("Error creating project: %v", err))
		return
	}

	showSuccess(projectName)
}

func showSuccess(projectName string) {
	cli.Newline()
	cli.Success("Project created successfully!")
	cli.Newline()
	cli.Bold(fmt.Sprintf("Your Velocity project '%s' is ready.", projectName))
	cli.Newline()
	cli.NextSteps([]string{
		"cd " + projectName,
		"go run . serve",
	})
	cli.Muted("Default port: 4000 (set PORT env to change)")
	cli.Newline()
}
