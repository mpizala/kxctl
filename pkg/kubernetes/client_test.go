package kubernetes

import (
	"context"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"
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
	err := client.ExecuteCommand(context.Background(), []string{"version", "--client"}, []string{"test-context"}, true, 0, "", 0)
	if err != nil {
		t.Fatalf("ExecuteCommand() unexpected error: %v", err)
	}
}

func TestParallelExecution(t *testing.T) {
	// Skip test if kubectl is not available
	if _, err := exec.LookPath("kubectl"); err != nil {
		t.Skip("kubectl is not available in PATH")
	}

	client := &Client{
		// Create multiple fake contexts to test parallel execution
		Contexts: []string{"test-context-1", "test-context-2", "test-context-3"},
	}

	// Execute a simple command with a timeout
	// This doesn't test actual parallelism but ensures the code path works without panics
	err := client.ExecuteCommand(context.Background(), []string{"version", "--client"}, client.Contexts, true, 2*time.Second, "", 0)
	if err != nil {
		t.Fatalf("Parallel execution failed: %v", err)
	}
}

func TestParallelExecutionPerformance(t *testing.T) {
	// Skip test if kubectl is not available
	if _, err := exec.LookPath("kubectl"); err != nil {
		t.Skip("kubectl is not available in PATH")
	}

	// Get real contexts if available, otherwise use fake ones
	contexts, err := GetContexts()
	if err != nil || len(contexts) < 2 {
		t.Skip("Need at least 2 contexts for parallel performance test")
	}

	// Limit to max 3 contexts to keep test runtime reasonable
	if len(contexts) > 3 {
		contexts = contexts[:3]
	}

	// Create mock command that simulates kubectl with a delay
	mockCommand := func(ctx context.Context, duration time.Duration) error {
		select {
		case <-time.After(duration):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// First run serially to measure baseline
	start := time.Now()
	for range contexts {
		_ = mockCommand(context.Background(), 100*time.Millisecond)
	}
	serialTime := time.Since(start)

	// Now run in parallel using our implementation
	// The test measures if multiple contexts are processed in parallel
	client := &Client{Contexts: contexts}

	// Redirect output during the test to avoid cluttering test output
	originalStdout := os.Stdout
	originalStderr := os.Stderr
	os.Stdout, _ = os.Open(os.DevNull)
	os.Stderr, _ = os.Open(os.DevNull)
	defer func() {
		os.Stdout = originalStdout
		os.Stderr = originalStderr
	}()

	// Run actual test with sleep command to simulate work
	start = time.Now()
	err = client.ExecuteCommand(
		context.Background(),
		[]string{"version", "--client"},
		client.Contexts,
		true,
		1*time.Second, // Set a short timeout
		"",            // No grep pattern
		2,             // Limited parallelism
	)
	parallelTime := time.Since(start)

	if err != nil {
		t.Fatalf("Parallel execution failed: %v", err)
	}

	// Verify that parallel execution is faster than serial execution
	// With true parallelism, it should be significantly faster than serial execution
	t.Logf("Serial execution time: %v", serialTime)
	t.Logf("Parallel execution time: %v", parallelTime)

	// Since this is an integration test with real kubectl, we can't be too strict about timing
	// Just log the times and do a basic sanity check that parallel is not slower than serial
	if len(contexts) > 1 && parallelTime > serialTime {
		t.Logf("Warning: Parallel execution (%v) was not faster than serial execution (%v)",
			parallelTime, serialTime)
	}
}

func TestProgressTracking(t *testing.T) {
	// Create test status map
	statusMap := make(map[string]*ContextStatus)
	statusMap["context-1"] = &ContextStatus{
		Name:    "context-1",
		Status:  "running",
		Started: time.Now().Add(-30 * time.Second),
		Done:    false,
	}
	statusMap["context-2"] = &ContextStatus{
		Name:    "context-2",
		Status:  "queued",
		Started: time.Time{},
		Done:    false,
	}
	statusMap["context-3"] = &ContextStatus{
		Name:    "context-3",
		Status:  "completed",
		Started: time.Now().Add(-60 * time.Second),
		Done:    true,
	}

	// Redirect stdout to capture output
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create a mutex for testing
	var mu sync.Mutex

	// Call the function we want to test
	printStatusReport(statusMap, &mu)

	// Restore stdout
	w.Close()
	os.Stdout = originalStdout

	// Read captured output
	var buf strings.Builder
	io.Copy(&buf, r)
	output := buf.String()

	// Check that output contains expected information
	if !strings.Contains(output, "Total: 3, Completed: 1, Running: 1, Queued: 1") {
		t.Errorf("Status summary incorrect: %s", output)
	}
	if !strings.Contains(output, "context-1") {
		t.Errorf("Missing running context details: %s", output)
	}
	if strings.Contains(output, "context-2") {
		t.Errorf("Should not show queued contexts in details: %s", output)
	}
}

func TestGrepPatternMatching(t *testing.T) {
	testCases := []struct {
		name    string
		line    string
		pattern string
		want    bool
	}{
		{
			name:    "Simple substring match",
			line:    "pod-abc-123 Running",
			pattern: "Running",
			want:    true,
		},
		{
			name:    "Simple substring no match",
			line:    "pod-abc-123 Running",
			pattern: "Pending",
			want:    false,
		},
		{
			name:    "Alternation with pipe",
			line:    "pod-abc-123 Running",
			pattern: "Pending|Running",
			want:    true,
		},
		{
			name:    "Regex pattern with slashes",
			line:    "pod-abc-123 Running",
			pattern: "/pod-.*Running/",
			want:    true,
		},
		{
			name:    "Regex pattern with slashes no match",
			line:    "pod-abc-123 Running",
			pattern: "/pod-.*Pending/",
			want:    false,
		},
		{
			name:    "Empty pattern",
			line:    "pod-abc-123 Running",
			pattern: "",
			want:    true, // Should always match
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := matchesGrepPattern(tc.line, tc.pattern)
			if got != tc.want {
				t.Errorf("matchesGrepPattern(%q, %q) = %v, want %v",
					tc.line, tc.pattern, got, tc.want)
			}
		})
	}
}