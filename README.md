# Velocity Installer

The official installer CLI for [Velocity](https://github.com/velocitykode/velocity) Go web framework.

## Installation

### Homebrew (macOS)

```bash
brew tap velocitykode/tap
brew install velocity
```

## Commands

```bash
velocity new <name>     # Create a new Velocity project
velocity init           # Initialize Velocity in existing project
velocity config         # Manage CLI configuration
velocity self-update    # Update to latest version (velocity and vel)

vel <command>           # Run a command in the current project (vel serve, vel migrate, vel gen ...)
```

## Quick Start

```bash
# Install
brew install velocitykode/tap/velocity

# Create a new project
velocity new myapp

# Project is ready with dev servers running
# Go server: http://localhost:4000
# Vite: http://localhost:5173
```

## Architecture

Velocity uses two CLI tools:

| Tool | Install | Purpose |
|------|---------|---------|
| `velocity` | Homebrew | Create projects, manage config |
| `vel` | Homebrew (ships with `velocity`) | Run project commands: dev server, migrations, generators |

`vel` is a thin launcher. It finds the enclosing project (nearest `go.mod`
requiring the framework), builds that project's own `./vel` with `go build`,
and hands it the command line. Every command lives in the project's framework
version, so updating `vel` never changes how a project behaves. If the build
fails, `vel` runs the last good `./vel` and says so. Without the installer,
`./vel <command>` or `go run . <command>` work the same.

When you run `velocity new`, it:
1. Scaffolds a new project
2. Installs dependencies
3. Runs migrations
4. Builds the `./vel` binary
5. Starts development servers

## Configuration

Set defaults for new projects:

```bash
velocity config set default.database postgres
velocity config set default.cache redis
velocity config set default.auth true
```

Configuration stored in `~/.vel/config.yaml`.

## Documentation

Full documentation at **[velocity.velocitykode.com/docs](https://velocity.velocitykode.com/docs)**

## License

MIT License - See [LICENSE](LICENSE) for details.
