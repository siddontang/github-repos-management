# GitHub Repository Management

A CLI tool for managing and tracking GitHub repositories, pull requests, and issues.

## Features

- Track multiple GitHub repositories
- View pull requests and issues across all tracked repositories
- Filter by state, author, repository, and more
- Direct integration with GitHub API
- File-based persistence for data storage

## Installation

### Prerequisites

- Go 1.16 or higher
- GitHub personal access token with appropriate permissions

### Building from source

1. Clone the repository:
   ```
   git clone https://github.com/siddontang/github-repos-management.git
   cd github-repos-management
   ```

2. Build the CLI using the Makefile:
   ```
   make build
   ```

   Or build directly:
   ```
   go build -o bin/ghrepos ./cmd/cli
   ```

## Configuration

The CLI uses GitHub authentication from your environment. Make sure you have the `GITHUB_TOKEN` environment variable set:

```bash
export GITHUB_TOKEN=your_github_personal_access_token
```

You can also configure the application by creating a `config.yaml` file:

```yaml
database:
  type: "file"
  path: "data/github-repos.db"

github:
  items_per_fetch: 100
```

## Usage

### Using the CLI

The CLI provides commands for managing repositories, pull requests, and issues.

#### Repository commands

```
# List all tracked repositories
./bin/ghrepos repo list

# Add a repository
./bin/ghrepos repo add owner/repo

# Remove a repository
./bin/ghrepos repo remove owner/repo

# Refresh a repository
./bin/ghrepos repo refresh owner/repo

# Refresh all repositories
./bin/ghrepos repo refresh
```

#### Pull request commands

```
# List all pull requests
./bin/ghrepos pr list

# List pull requests for a specific repository
./bin/ghrepos pr list --repo owner/repo

# List open pull requests
./bin/ghrepos pr list --state open

# List pull requests by author
./bin/ghrepos pr list --author username
```

#### Issue commands

```
# List all issues
./bin/ghrepos issue list

# List issues for a specific repository
./bin/ghrepos issue list --repo owner/repo

# List open issues
./bin/ghrepos issue list --state open

# List issues by author
./bin/ghrepos issue list --author username
```

#### Status command

```
# Get service status
./bin/ghrepos status
```

## Architecture

The CLI directly integrates with the GitHub API through a service layer, providing:

- Direct GitHub API access
- File-based database for persistent storage
- Simplified data model for repositories, pull requests, and issues

## Development

### Available make commands

```
make build      - Build CLI
make test       - Run tests
make clean      - Clean build artifacts
make clean-empty - Remove empty directories
make dist       - Create distribution package
make help       - Show help
make push       - Push to GitHub repository
```

## License

This project is licensed under the MIT License - see the LICENSE file for details. 