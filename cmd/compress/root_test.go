// Tests for the CLI root command and compress subcommand
package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/yourorg/video-compressor/internal/presets"
)

// TestRoot_UnknownPreset verifies that the CLI returns an error when the preset is unknown
func TestRoot_UnknownPreset(t *testing.T) {
	origLoad := loadPresetsFunc
	defer func() { loadPresetsFunc = origLoad }()
	loadPresetsFunc = func() (map[string]presets.Preset, error) {
		return map[string]presets.Preset{"foo": {}}, nil
	}

	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"compress", "--preset=bar", "file.mp4"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for unknown preset, got nil")
	}
	if !strings.Contains(err.Error(), "unknown preset") {
		t.Errorf("expected 'unknown preset' error, got: %v", err)
	}
}

// TestRoot_LoadPresetsError verifies that a LoadAll error is propagated
func TestRoot_LoadPresetsError(t *testing.T) {
	origLoad := loadPresetsFunc
	defer func() { loadPresetsFunc = origLoad }()
	loadPresetsFunc = func() (map[string]presets.Preset, error) {
		return nil, fmt.Errorf("boom")
	}

	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"compress", "file.mp4"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for LoadAll failure, got nil")
	}
	if !strings.Contains(err.Error(), "loading presets") {
		t.Errorf("expected 'loading presets' error, got: %v", err)
	}
}

// TestRoot_CompressError verifies that a compression failure is reported
func TestRoot_CompressError(t *testing.T) {
	origLoad := loadPresetsFunc
	origComp := compressFunc
	defer func() {
		loadPresetsFunc = origLoad
		compressFunc = origComp
	}()

	loadPresetsFunc = func() (map[string]presets.Preset, error) {
		return map[string]presets.Preset{"default": {}}, nil
	}
	compressFunc = func(in, out string, args []string) error {
		return fmt.Errorf("fail")
	}

	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"compress", "file.mp4"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected compression error, got nil")
	}
	if !strings.Contains(err.Error(), "one or more files failed to compress") {
		t.Errorf("expected 'one or more files failed to compress' error, got: %v", err)
	}
}

// TestRoot_Success verifies the CLI prints 'All done!' when compression succeeds
func TestRoot_Success(t *testing.T) {
	origLoad := loadPresetsFunc
	origComp := compressFunc

	// Get the compressFunc from the queue package
	origQueueCompress := getQueueCompressFunc()

	defer func() {
		loadPresetsFunc = origLoad
		compressFunc = origComp
		setQueueCompressFunc(origQueueCompress)
	}()

	loadPresetsFunc = func() (map[string]presets.Preset, error) {
		return map[string]presets.Preset{"default": {}}, nil
	}

	// Create a mock compress function that always returns nil
	mockCompressFunc := func(in, out string, args []string) error {
		// Skip actual ffmpeg execution and just return success
		return nil
	}

	// Set the mock in both places
	compressFunc = mockCompressFunc
	setQueueCompressFunc(mockCompressFunc)

	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"compress", "my.mp4"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !strings.Contains(buf.String(), "All done!") {
		t.Errorf("expected 'All done!' in output, got: %s", buf.String())
	}
}

// TestRoot_NoSubcommand verifies the application displays help when no subcommand is provided
func TestRoot_NoSubcommand(t *testing.T) {
	// capture os.Exit
	var code int
	oldExit := osExit
	osExit = func(c int) { code = c }
	defer func() { osExit = oldExit }()

	// The root command without subcommands shows help and doesn't exit with error
	// Update test to expect exit code 0
	os.Args = []string{"video-compress"}
	main()
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
}

// TestRoot_UnknownCommand verifies the application exits with code 1 for unknown commands
func TestRoot_UnknownCommand(t *testing.T) {
	var code int
	old := osExit
	osExit = func(c int) { code = c }
	defer func() { osExit = old }()

	os.Args = []string{"video-compress", "bizazzle"}
	main()
	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}

// TestMustGetFunctions tests that mustGetString panics when the flag doesn't exist
func TestMustGetFunctions(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}

	// Test mustGetString panics
	defer func() {
		if r := recover(); r == nil {
			t.Error("mustGetString should have panicked for non-existent flag")
		}
	}()
	mustGetString(cmd, "non-existent-flag")
	t.Error("Expected panic, got none")
}

// TestMustGetInt tests that mustGetInt panics when the flag doesn't exist
func TestMustGetInt(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}

	defer func() {
		if r := recover(); r == nil {
			t.Error("mustGetInt should have panicked for non-existent flag")
		}
	}()
	mustGetInt(cmd, "non-existent-flag")
	t.Error("Expected panic, got none")
}

// TestRoot_CompressWithOverrides verifies that the --video-codec, --ffpreset, and --crf flags
// properly override values from the preset
func TestRoot_CompressWithOverrides(t *testing.T) {
	origLoad := loadPresetsFunc
	origCompressFunc := compressFunc
	origQueueCompressFunc := getQueueCompressFunc()

	defer func() {
		loadPresetsFunc = origLoad
		compressFunc = origCompressFunc
		setQueueCompressFunc(origQueueCompressFunc)
	}()

	// Create a preset with known values to be overridden
	loadPresetsFunc = func() (map[string]presets.Preset, error) {
		return map[string]presets.Preset{
			"default": {
				VideoCodec: "libx264", // default codec
				Preset:     "medium",  // default preset
				CRF:        23,        // default CRF
			},
		}, nil
	}

	// Capture the args passed to ffmpeg to verify overrides were applied
	var capturedArgs []string
	mockCompressFunc := func(in, out string, args []string) error {
		capturedArgs = args
		return nil
	}

	compressFunc = mockCompressFunc
	setQueueCompressFunc(mockCompressFunc)

	// Test case 1: All overrides
	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// Run with all override flags
	cmd.SetArgs([]string{
		"compress",
		"--video-codec", "libvpx-vp9", // override codec
		"--ffpreset", "fast", // override preset
		"--crf", "30", // override CRF
		"input.mp4",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Expected success, got: %v", err)
	}

	// Verify the arguments contain our overridden values
	expectedArgs := []string{
		"-c:v", "libvpx-vp9", // overridden codec
		"-preset", "fast", // overridden preset
		"-crf", "30", // overridden CRF
	}

	// Check that each expected arg appears in the captured args
	for i := 0; i < len(expectedArgs); i += 2 {
		flag := expectedArgs[i]
		value := expectedArgs[i+1]

		found := false
		for j := 0; j < len(capturedArgs); j += 2 {
			if j+1 < len(capturedArgs) && capturedArgs[j] == flag && capturedArgs[j+1] == value {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Expected %s=%s in ffmpeg args, got: %v", flag, value, capturedArgs)
		}
	}

	// Test case 2: Single override
	// Create a new command instance to avoid flags persisting between executions
	cmd = newRootCmd()
	buf = new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{
		"compress",
		"--crf", "18", // only override CRF
		"input.mp4",
	})
	capturedArgs = nil // Clear previous captured args

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Expected success, got: %v", err)
	}

	// Check that CRF was overridden but other values remained default
	foundCRF := false
	foundCodec := false
	foundPreset := false

	for j := 0; j < len(capturedArgs); j += 2 {
		if j+1 < len(capturedArgs) {
			switch capturedArgs[j] {
			case "-crf":
				foundCRF = capturedArgs[j+1] == "18" // should be overridden
			case "-c:v":
				foundCodec = capturedArgs[j+1] == "libx264" // should be default
			case "-preset":
				foundPreset = capturedArgs[j+1] == "medium" // should be default
			}
		}
	}

	if !foundCRF {
		t.Errorf("CRF=18 override not found in args: %v", capturedArgs)
	}
	if !foundCodec {
		t.Errorf("Default codec not preserved: %v", capturedArgs)
	}
	if !foundPreset {
		t.Errorf("Default preset not preserved: %v", capturedArgs)
	}
}
