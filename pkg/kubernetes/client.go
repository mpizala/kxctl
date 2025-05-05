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

// ExecuteCommand executes a kubectl command in the specified contexts
func (c *Client) ExecuteCommand(ctx context.Context, kubectlArgs []string, contexts []string) error {
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
