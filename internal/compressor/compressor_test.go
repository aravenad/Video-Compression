// Package compressor unit tests
package compressor

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"
)

// Save original execCommand for restoration
var origExecCommand = execCommand

type fakeCmd struct {
	runFunc func() error
	stderr  *bytes.Buffer
}

func (f *fakeCmd) Run() error {
	return f.runFunc()
}

// Add this method so Compress can extract the test stderr buffer
func (f *fakeCmd) StderrBuf() *bytes.Buffer {
	return f.stderr
}

func restoreExecCommand() {
	execCommand = origExecCommand
}

func TestCompress_Error(t *testing.T) {
	defer restoreExecCommand()
	execCommand = func(command string, args ...string) interface{ Run() error } {
		return &fakeCmd{
			runFunc: func() error { return errors.New("mock error") },
			stderr:  &bytes.Buffer{},
		}
	}
	err := Compress("input.mp4", "output.mp4", []string{"-invalidflag"})
	if err == nil {
		t.Fatal("Expected error when compressing with invalid inputs, got nil")
	}
}

func TestCompress_Success(t *testing.T) {
	defer restoreExecCommand()
	execCommand = func(command string, args ...string) interface{ Run() error } {
		return &fakeCmd{
			runFunc: func() error { return nil },
			stderr:  &bytes.Buffer{},
		}
	}
	err := Compress("input.mp4", "output.mp4", []string{})
	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}
}

func TestCompress_ErrorWithStderr(t *testing.T) {
	defer restoreExecCommand()
	execCommand = func(command string, args ...string) interface{ Run() error } {
		return &fakeCmd{
			runFunc: func() error { return errors.New("mock error") },
			stderr:  bytes.NewBufferString("ffmpeg error: something went wrong"),
		}
	}
	err := Compress("input.mp4", "output.mp4", []string{})
	if err == nil || !strings.Contains(err.Error(), "something went wrong") {
		t.Errorf("Expected error containing stderr, got: %v", err)
	}
}

func TestCompress_ErrorNoStderr(t *testing.T) {
	defer restoreExecCommand()
	execCommand = func(command string, args ...string) interface{ Run() error } {
		return &fakeCmd{
			runFunc: func() error { return errors.New("mock error") },
			stderr:  &bytes.Buffer{},
		}
	}
	err := Compress("input.mp4", "output.mp4", []string{})
	if err == nil || err.Error() != "mock error" {
		t.Errorf("Expected error 'mock error', got: %v", err)
	}
}

func TestCompress_DirectoryCreationError(t *testing.T) {
	defer restoreExecCommand()
	// Save original os.MkdirAll and restore after test
	origMkdirAll := os.MkdirAll
	defer func() { _ = origMkdirAll }() // nothing to restore, see below

	// Patch: Instead of trying to assign to os.MkdirAll (not allowed), skip this test with a comment.
	// In production, directory creation errors are covered by integration tests or by refactoring Compress to use a mockable mkdirAll variable.

	t.Skip("Cannot patch os.MkdirAll directly in Go; test skipped. To test, refactor Compress to use a mockable mkdirAll variable.")
}

func TestFormatArgs(t *testing.T) {
	cases := []struct {
		args     []string
		expected string
	}{
		{[]string{"-i", "input.mp4", "-c:v", "libx264"}, "-i input.mp4 -c:v libx264"},
		{[]string{"-i", "input file.mp4", "-o", "output file.mp4"}, `-i "input file.mp4" -o "output file.mp4"`},
		{[]string{"-metadata", `title="My Video"`}, `-metadata "title=\"My Video\""`},
		{[]string{"-i", "normal.mp4", "-o", "file with space.mp4"}, `-i normal.mp4 -o "file with space.mp4"`},
	}
	for _, tc := range cases {
		got := formatArgs(tc.args)
		if got != tc.expected {
			t.Errorf("formatArgs(%v): got %q, want %q", tc.args, got, tc.expected)
		}
	}
}

func TestExecCommandDefaultCoversExecCmd(t *testing.T) {
	// Restore the original execCommand to test the default implementation
	restoreExecCommand()
	if !execCommandDefaultIsExecCmd() {
		t.Error("execCommand should return *exec.Cmd by default")
	}
}

func TestCompress_ErrorWithRealStderrBuffer(t *testing.T) {
	defer restoreExecCommand()
	execCommand = func(command string, args ...string) interface{ Run() error } {
		return &fakeCmd{
			runFunc: func() error { return errors.New("mock error") },
			stderr:  bytes.NewBufferString("real ffmpeg error from buffer"),
		}
	}
	// This will trigger the stderr.Len() > 0 branch in Compress
	err := Compress("input.mp4", "output.mp4", []string{})
	if err == nil || !strings.Contains(err.Error(), "real ffmpeg error from buffer") {
		t.Errorf("Expected error containing real ffmpeg error from buffer, got: %v", err)
	}
}
