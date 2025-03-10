# GitHub Repository Management

A tool for managing and tracking GitHub repositories, pull requests, and issues.

## Features

- Track multiple GitHub repositories
- View pull requests and issues across all tracked repositories
- Filter by state, author, repository, and more
- Automatic synchronization with GitHub
- In-memory caching for fast access

## Installation

### Prerequisites

- Go 1.16 or higher
- GitHub CLI (`gh`) installed and authenticated

### Building from source

1. Clone the repository:
   ```
   git clone https://github.com/yourusername/github-repos-management.git
   cd github-repos-management
   ```

2. Build the server and CLI:
   ```
   go build -o bin/server ./cmd/server
   go build -o bin/ghrepos ./cmd/cli
   ```

## Configuration

1. Copy the example configuration file:
   ```
   cp config.yaml.example config.yaml
   ```

2. Edit the configuration file to match your requirements:
   ```yaml
   server:
     host: "localhost"
     port: 8080

   database:
     type: "memory"

   github:
     items_per_fetch: 100
   ```

## Usage

### Starting the server

```
./bin/server -config config.yaml
```

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

#### Service commands

```
# Get service status
./bin/ghrepos service status

# Refresh all repositories
./bin/ghrepos service refresh
```

## License

This project is licensed under the MIT License - see the LICENSE file for details. 