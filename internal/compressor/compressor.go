// Package compressor provides video compression functionality by invoking
// external ffmpeg processes with appropriate arguments.
package compressor

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Make exec.Command mockable for testing
var execCommand = func(command string, args ...string) interface{ Run() error } {
	return exec.Command(command, args...)
}

// Compress invokes ffmpeg with args based on the preset.
// Parameters:
//   - input: path to input video file
//   - output: path where output will be written
//   - args: additional ffmpeg arguments, typically from presets.BuildFFArgs
//
// Returns an error if ffmpeg execution fails.
func Compress(input, output string, args []string) error {
	// Fix path issues by normalizing all paths (use forward slashes even on Windows)
	input = filepath.Clean(input)
	output = filepath.Clean(output)

	// Convert Windows backslashes to forward slashes for ffmpeg
	input = strings.ReplaceAll(input, "\\", "/")
	output = strings.ReplaceAll(output, "\\", "/")

	// Ensure output directory exists
	outDir := filepath.Dir(output)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	fullArgs := []string{"-i", input}
	fullArgs = append(fullArgs, args...)
	fullArgs = append(fullArgs, output)

	fmt.Printf("Running: ffmpeg %s\n", formatArgs(fullArgs))
	cmd := execCommand("ffmpeg", fullArgs...)

	// Capture stderr for better error reporting
	var stderr bytes.Buffer
	// Only set Stderr if the returned type has a Stderr field
	if c, ok := cmd.(*exec.Cmd); ok {
		c.Stderr = &stderr
	}

	err := cmd.Run()
	// If we are in a test, and the mock provides its own stderr, use that
	if err != nil {
		// Try to get a .stderr field from the mock (for tests)
		type stderrProvider interface{ StderrBuf() *bytes.Buffer }
		if sprov, ok := cmd.(stderrProvider); ok && sprov.StderrBuf() != nil && sprov.StderrBuf().Len() > 0 {
			return fmt.Errorf("%w: %s", err, strings.TrimSpace(sprov.StderrBuf().String()))
		}
		if stderr.Len() > 0 {
			return fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
		}
		return err
	}

	return nil
}

// formatArgs formats command arguments for printing, handling special characters properly
func formatArgs(args []string) string {
	quoted := make([]string, len(args))
	for i, arg := range args {
		if strings.Contains(arg, " ") || strings.Contains(arg, "\"") {
			quoted[i] = fmt.Sprintf("%q", arg) // Properly quote arguments with spaces
		} else {
			quoted[i] = arg
		}
	}
	return strings.Join(quoted, " ")
}

// Test helper to cover the execCommand default implementation
func execCommandDefaultIsExecCmd() bool {
	cmd := execCommand("echo", "foo")
	_, ok := cmd.(*exec.Cmd)
	return ok
}
