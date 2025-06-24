# kxctl - kubectl for cross-cluster activities

A command-line utility that enhances the usability of kubectl by applying filters and various commands across multiple Kubernetes contexts.

## Features

- Execute kubectl commands across multiple contexts
- Filter contexts using include/exclude patterns
- Filter output with grep functionality (supports basic patterns and regex)
- Interactive progress tracking (press Enter during execution to see status)
- Safety checks for write operations requiring explicit confirmation
- Simplify common Kubernetes operations across multiple clusters

## Installation

### Using Homebrew (macOS and Linux)

```bash
brew install mpizala/utils/kxctl
```

### Using Go

```bash
go install github.com/mpizala/kxctl/cmd/kxctl@latest
```

## Usage

```
kxctl - Kubernetes Context Control

Usage:
  kxctl [command] [flags] [-- kubectl_command]

Commands:
  list        List available contexts
  exec        Execute kubectl command on filtered contexts
  status      Show pods not in Running or Succeeded state
  version     Display version information
  help        Display help information

Flags:
  -i, --include pattern   Include contexts matching pattern (can be used multiple times)
  -e, --exclude pattern   Exclude contexts matching pattern (can be used multiple times)
  -g, --grep pattern      Filter command output to lines matching pattern
  -p, --parallel number   Limit parallel execution to specified number of contexts
  -t, --timeout duration  Set timeout for kubectl commands (e.g. 30s, 1m, 2m30s)
  -f, --force             Force execution of write operations
  -A, --all-namespaces    Show resources across all namespaces (status command)
  -h, --help              Display help information
```

### Examples

```bash
# List all contexts
kxctl list

# List contexts matching 'prod'
kxctl list -i prod

# Run a command on all contexts
kxctl exec -- get pods

# Run a command on contexts matching a pattern
kxctl exec -i production -- get pods

# Run a command excluding contexts matching a pattern
kxctl exec -e staging -- get pods

# Shorthand syntax (starting with flags implies 'exec')
kxctl -i prod -- get pods

# Run a write operation with force flag
kxctl exec -f -i prod -- apply -f deployment.yaml

# Show problematic pods in the current namespace
kxctl status

# Show problematic pods in production clusters
kxctl status -i prod -A

# Show problematic pods with additional kubectl args
kxctl status -- -o json

# Show problematic pods across namespace webapp
kxctl status -- --namespace webapp

# Run a command with a timeout (useful for slow or unresponsive clusters)
kxctl exec -t 30s -- get pods

# Filter kubectl output with pipe-like syntax using the --grep flag
kxctl exec -g "coredns|web-app" -- get pods -A

# Interactive progress tracking: run a command and press Enter during execution
# to see which clusters are currently being processed
kxctl exec -i prod -- get pods  # Press Enter while waiting to see progress

# Limit parallel execution to prevent kubelogin or authentication issues
kxctl exec -p 3 -i prod -- get pods
```

## License

MIT