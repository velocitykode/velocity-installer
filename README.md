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
velocity self-update    # Update to latest version
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
| [`vel`](https://github.com/velocitykode/vel) | Built from source | Dev server, migrations, generators |

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

Full documentation at **[velocitykode.com/docs](https://velocitykode.com/docs)**

## License

MIT License - See [LICENSE](LICENSE) for details.
