# kxctl - kubectl for cross-cluster activities

A command-line utility that enhances the usability of kubectl by applying filters and various commands across multiple Kubernetes contexts.

## Features

- Execute kubectl commands across multiple contexts
- Filter contexts using include/exclude patterns
- Safety checks for write operations requiring explicit confirmation
- Simplify common Kubernetes operations across multiple clusters

## Installation

```
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
  version     Display version information
  help        Display help information

Notes:
  - No command provided: shows help information
  - Leading flags without command: treated as 'exec'

Flags:
  -i, --include pattern   Include contexts matching pattern (can be used multiple times)
  -e, --exclude pattern   Exclude contexts matching pattern (can be used multiple times)
  -f, --force             Force execution of write operations
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
```

## License

MIT