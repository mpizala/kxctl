package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/mpizala/kxctl/pkg/filter"
	"github.com/mpizala/kxctl/pkg/kubernetes"
)

const version = "0.1.0"

type stringSliceFlag []string

func (s *stringSliceFlag) String() string {
	return strings.Join(*s, ", ")
}

func (s *stringSliceFlag) Set(value string) error {
	*s = append(*s, value)
	return nil
}

type commandFlags struct {
	include stringSliceFlag
	exclude stringSliceFlag
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		// No command provided, default to help
		printHelp()
		return nil
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	// Special case for help flags
	if cmd == "--help" || cmd == "-h" {
		printHelp()
		return nil
	}

	// Special case for flags without command (treat as exec)
	if strings.HasPrefix(cmd, "-") {
		return runExec(os.Args[1:])
	}

	switch cmd {
	case "version":
		return runVersion()
	case "help":
		printHelp()
		return nil
	case "list":
		return runList(args)
	case "exec":
		return runExec(args)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", cmd)
		printHelp()
		return fmt.Errorf("unknown command: %s", cmd)
	}
}

func runVersion() error {
	fmt.Printf("kxctl version %s\n", version)
	return nil
}

func runList(args []string) error {
	flags, err := parseFlags(args)
	if err != nil {
		return err
	}

	client, err := kubernetes.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	filteredContexts := filter.FilterContexts(client.Contexts, flags.include, flags.exclude)
	if len(filteredContexts) == 0 {
		return errors.New("no contexts match the provided filters")
	}

	for _, ctx := range filteredContexts {
		fmt.Println(ctx)
	}
	return nil
}

func runExec(args []string) error {
	// Extract kubectl args after --
	var kubectlArgs []string
	cmdIndex := -1

	for i, arg := range args {
		if arg == "--" {
			cmdIndex = i
			break
		}
	}

	var flagArgs []string
	if cmdIndex != -1 {
		flagArgs = args[:cmdIndex]
		kubectlArgs = args[cmdIndex+1:]
	} else {
		// Check for any flags, assume the rest are kubectl args
		var i int
		for i = 0; i < len(args); i++ {
			if !strings.HasPrefix(args[i], "-") {
				break
			}
			flagArgs = append(flagArgs, args[i])
			if strings.HasPrefix(args[i], "-") && !strings.HasPrefix(args[i], "--") && len(args[i]) == 2 && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				// Handle -i value and similar
				flagArgs = append(flagArgs, args[i+1])
				i++
			}
		}
		kubectlArgs = args[i:]
	}

	flags, err := parseFlags(flagArgs)
	if err != nil {
		return err
	}

	client, err := kubernetes.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	filteredContexts := filter.FilterContexts(client.Contexts, flags.include, flags.exclude)
	if len(filteredContexts) == 0 {
		return errors.New("no contexts match the provided filters")
	}

	if len(kubectlArgs) == 0 {
		// No kubectl command specified, just list the contexts
		for _, ctx := range filteredContexts {
			fmt.Println(ctx)
		}
		return nil
	}

	// Execute kubectl command on filtered contexts
	return client.ExecuteCommand(context.Background(), kubectlArgs, filteredContexts)
}

func parseFlags(args []string) (commandFlags, error) {
	var flags commandFlags

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-i", "--include":
			if i+1 >= len(args) || strings.HasPrefix(args[i+1], "-") {
				return flags, errors.New("include flag requires a value")
			}
			flags.include = append(flags.include, args[i+1])
			i++
		case "-e", "--exclude":
			if i+1 >= len(args) || strings.HasPrefix(args[i+1], "-") {
				return flags, errors.New("exclude flag requires a value")
			}
			flags.exclude = append(flags.exclude, args[i+1])
			i++
		default:
			if strings.HasPrefix(args[i], "-") {
				return flags, fmt.Errorf("unknown flag: %s", args[i])
			}
			// If not a flag, stop processing
			break
		}
	}

	return flags, nil
}

func printHelp() {
	fmt.Print(`kxctl - Kubernetes Context Control

Usage:
  kxctl [command] [flags] [-- kubectl_command]

Commands:
  list        List available contexts
  exec        Execute kubectl command on filtered contexts
  version     Display version information
  help        Display help information

Notes:
  - No command provided: shows this help information
  - Leading flags without command: treated as 'exec'

Flags:
  -i, --include pattern   Include contexts matching pattern (can be used multiple times)
  -e, --exclude pattern   Exclude contexts matching pattern (can be used multiple times)
  -h, --help              Display this help information

Examples:
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
`)
}
