package kubernetes

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Client represents a kubectl client interface
type Client struct {
	Contexts []string
}

// NewClient creates a new kubernetes client
func NewClient() (*Client, error) {
	contexts, err := GetContexts()
	if err != nil {
		return nil, err
	}

	return &Client{
		Contexts: contexts,
	}, nil
}

// GetContexts returns all available kubectl contexts
func GetContexts() ([]string, error) {
	cmd := exec.Command("kubectl", "config", "get-contexts", "-o", "name")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubectl contexts: %w", err)
	}

	contexts := strings.Split(strings.TrimSpace(string(output)), "\n")
	return contexts, nil
}

// isWriteOperation checks if the kubectl command is a write operation (non-read-only)
func isWriteOperation(args []string) bool {
	if len(args) == 0 {
		return false
	}

	// List of kubectl verbs that modify resources
	writeVerbs := map[string]bool{
		"apply":     true,
		"create":    true,
		"delete":    true,
		"edit":      true,
		"patch":     true,
		"replace":   true,
		"scale":     true,
		"set":       true,
		"label":     true,
		"annotate":  true,
		"taint":     true,
		"drain":     true,
		"cordon":    true,
		"uncordon":  true,
		"rollout":   true,
		"autoscale": true,
	}

	// Check first argument as the verb
	return writeVerbs[args[0]]
}

// ExecuteCommand executes a kubectl command in the specified contexts
// If timeout is greater than 0, it will be applied to each command execution
// If grepPattern is not empty, output will be filtered to lines matching the pattern
func (c *Client) ExecuteCommand(ctx context.Context, kubectlArgs []string, contexts []string, force bool, timeout time.Duration, grepPattern string) error {
	if isWriteOperation(kubectlArgs) && !force {
		return fmt.Errorf("write operation detected: '%s'. Use --force flag to confirm", strings.Join(kubectlArgs, " "))
	}

	// Create a wait group to wait for all commands to complete
	var wg sync.WaitGroup
	wg.Add(len(contexts))

	// Create a mutex to protect concurrent writes to stdout/stderr
	var outputMutex sync.Mutex

	// Execute commands in parallel
	for _, contextName := range contexts {
		// Capture the current context for goroutine
		currentContext := contextName

		go func() {
			defer wg.Done()

			// Create the kubectl command with context
			args := append([]string{"--context", currentContext}, kubectlArgs...)
			cmd := exec.CommandContext(ctx, "kubectl", args...)

			// Set a timeout if specified
			if timeout > 0 {
				timer := time.AfterFunc(timeout, func() {
					cmd.Process.Kill()
				})
				defer timer.Stop()
			}

			// Capture output
			output, err := cmd.CombinedOutput()
			outputStr := string(output)

			// Print output and handle errors
			outputMutex.Lock()
			fmt.Printf("Context: %s\n", currentContext)

			// Process output line by line to ensure context association
			lines := strings.Split(outputStr, "\n")
			for _, line := range lines {
				if line != "" {
					// Filter lines by grep pattern if provided
					if grepPattern == "" || matchesGrepPattern(line, grepPattern) {
						fmt.Printf("  %s\n", line)
					}
				}
			}

			// Handle errors
			if err != nil {
				// Check if it was killed by timeout
				if timeout > 0 && cmd.ProcessState != nil && cmd.ProcessState.Exited() == false {
					fmt.Fprintf(os.Stderr, "  Timeout after %s\n", timeout)
				} else {
					fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
				}
			}

			// Add a separator line for readability between contexts
			fmt.Println()
			outputMutex.Unlock()
		}()
	}

	// Wait for all commands to complete
	wg.Wait()
	return nil
}

// matchesGrepPattern checks if a line matches the grep pattern
// Supports either standard string contains or regex patterns when enclosed in /pattern/
func matchesGrepPattern(line, pattern string) bool {
	// Check if it's a regex pattern (enclosed in slashes)
	if strings.HasPrefix(pattern, "/") && strings.HasSuffix(pattern, "/") && len(pattern) > 2 {
		regexPattern := pattern[1 : len(pattern)-1]
		reg, err := regexp.Compile(regexPattern)
		if err == nil && reg.MatchString(line) {
			return true
		}
		return false
	}

	// Support for egrep-like alternation with |
	if strings.Contains(pattern, "|") {
		patterns := strings.Split(pattern, "|")
		for _, p := range patterns {
			if strings.Contains(line, p) {
				return true
			}
		}
		return false
	}

	// Simple substring match
	return strings.Contains(line, pattern)
}
