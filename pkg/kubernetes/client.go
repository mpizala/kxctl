package kubernetes

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
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
func (c *Client) ExecuteCommand(ctx context.Context, kubectlArgs []string, contexts []string, force bool) error {
	if isWriteOperation(kubectlArgs) && !force {
		return fmt.Errorf("write operation detected: '%s'. Use --force flag to confirm", strings.Join(kubectlArgs, " "))
	}

	for _, context := range contexts {
		fmt.Printf("Context: %s\n", context)

		args := append([]string{"--context", context}, kubectlArgs...)
		cmd := exec.CommandContext(ctx, "kubectl", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error in context %s: %v\n", context, err)
		}
	}

	return nil
}
