package main

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/yourorg/video-compressor/internal/presets"
)

// TestRoot_UnknownPreset verifies that the CLI returns an error when the preset is unknown.
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
	cmd.SetArgs([]string{"--preset=bar", "file.mp4"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for unknown preset, got nil")
	}
	if !strings.Contains(err.Error(), "unknown preset") {
		t.Errorf("expected 'unknown preset' error, got: %v", err)
	}
}

// TestRoot_LoadPresetsError verifies that a LoadAll error is propagated.
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
	cmd.SetArgs([]string{"file.mp4"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for LoadAll failure, got nil")
	}
	if !strings.Contains(err.Error(), "loading presets") {
		t.Errorf("expected 'loading presets' error, got: %v", err)
	}
}

// TestRoot_CompressError verifies that a compression failure is reported.
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
	cmd.SetArgs([]string{"file.mp4"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected compression error, got nil")
	}
	if !strings.Contains(err.Error(), "one or more files failed to compress") {
		t.Errorf("expected 'one or more files failed to compress' error, got: %v", err)
	}
}

// TestRoot_Success verifies the CLI prints 'All done!' when compression succeeds.
func TestRoot_Success(t *testing.T) {
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
		return nil
	}

	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"my.mp4"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !strings.Contains(buf.String(), "All done!") {
		t.Errorf("expected 'All done!' in output, got: %s", buf.String())
	}
}
