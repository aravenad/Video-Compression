package compressor

import (
	"errors"
	"os/exec"
	"testing"
)

// CommandRunner is an interface for command execution
type CommandRunner interface {
	Run() error
}

// mockCmd implements a simplified version of exec.Cmd for testing
type mockCmd struct {
	shouldSucceed bool
	stderr        string
}

func (m *mockCmd) Run() error {
	if !m.shouldSucceed {
		return errors.New("mock error")
	}
	return nil
}

// mockExecCmd is a wrapper that satisfies the *exec.Cmd interface
type mockExecCmd struct {
	*exec.Cmd
	mock *mockCmd
}

// TestCompressFailure tests that compression properly handles command failures
func TestCompressFailure(t *testing.T) {
	// Save the original execCommand and restore it after the test
	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()

	execCommand = func(command string, args ...string) interface{ Run() error } {
		return &mockCmd{shouldSucceed: false}
	}

	// Run Compress with our mock
	err := Compress("input.mp4", "output.mp4", []string{})

	// Verify we got an error
	if err == nil {
		t.Fatal("Expected error but got nil")
	}
	// Check error message
	if err.Error() != "mock error" {
		t.Errorf("Expected error message 'mock error', got: %s", err.Error())
	}
}
