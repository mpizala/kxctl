package kubernetes

import (
	"context"
	"os"
	"os/exec"
	"testing"
)

func TestGetContexts(t *testing.T) {
	// Skip test if kubectl is not available
	if _, err := exec.LookPath("kubectl"); err != nil {
		t.Skip("kubectl is not available in PATH")
	}

	contexts, err := GetContexts()
	if err != nil {
		t.Fatalf("GetContexts() error = %v", err)
	}

	if len(contexts) == 0 && os.Getenv("CI") != "true" {
		t.Log("No contexts found. This is expected in CI environments but might be an issue locally.")
	}
}

func TestExecuteCommand(t *testing.T) {
	// Skip test if kubectl is not available
	if _, err := exec.LookPath("kubectl"); err != nil {
		t.Skip("kubectl is not available in PATH")
	}

	client := &Client{
		Contexts: []string{"test-context"},
	}

	// This test doesn't actually execute kubectl since the context is fake
	// It just verifies that the function doesn't panic
	err := client.ExecuteCommand(context.Background(), []string{"version", "--client"}, []string{"test-context"}, true)
	if err != nil {
		t.Fatalf("ExecuteCommand() unexpected error: %v", err)
	}
}
